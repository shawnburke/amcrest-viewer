package ftp

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"go.uber.org/zap"

	fd "github.com/goftp/file-driver"
	ftps "github.com/goftp/server"
	"github.com/shawnburke/amcrest-viewer/common"
)

// For each user, we create a different file root.
// Because the driver doesn't know the user until Init, we have to use a proxy strategy
// which is that we create a proxy driver that looks up the actual driver based on the user.
// That actual driver is responsible for creating the directory location, and for handling
// all requests for that user.
//
// On top of that, we manage an index of the files and send the ones we want down to the notification system.
//
// TODOS:
// 1. Implement a garage collection policy for files based on LRU (and maybe total size)

type driverFactory struct {
	ftps.Perm
	logger     *zap.Logger
	bus        common.EventBus
	userSpaces userSpaces
}

const tempRoot = "amz-ftp"

func getTempRoot() string {
	return path.Join(os.TempDir(), tempRoot)
}

func newDriverFactory(perm ftps.Perm, logger *zap.Logger, bus common.EventBus) *driverFactory {
	// add a subdir to make sure we use a new one each time we start,
	// as an added measure against leaks.
	tmpRoot := getTempRoot()
	tempSubDir := "cams-" + time.Now().Format("2006-01-02")
	spaceDir := path.Join(tmpRoot, tempSubDir)
	df := &driverFactory{
		Perm:       perm,
		logger:     logger,
		bus:        bus,
		userSpaces: newUserSpaces(spaceDir, logger),
	}
	go func() {
		df.gc(tmpRoot, time.Hour*24)
		time.Sleep(time.Hour * 24)
	}()
	return df
}

func (df *driverFactory) findNewestFile(dir string) time.Time {
	var newest time.Time
	fs.WalkDir(os.DirFS(dir), ".",
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				df.logger.Error("Error walking dir", zap.Error(err))
				return err
			}
			stat, _ := d.Info()
			if stat.ModTime().UTC().UnixMilli() > newest.UTC().UnixMilli() {
				newest = stat.ModTime()
			}
			return nil
		})
	return newest
}

func (df *driverFactory) gc(
	dir string,
	maxAge time.Duration,
) {

	// get the subdirectories of the temp root)
	subDirs, err := os.ReadDir(dir)
	if err != nil {
		df.logger.Error("Error reading temp root", zap.Error(err))
		return
	}

	// for each directory, find the newest file and delete
	// the whole directory if it's older than the max age
	for _, d := range subDirs {

		if !d.IsDir() {
			continue
		}

		fullPath := path.Join(dir, d.Name())

		newest := df.findNewestFile(fullPath)

		if time.Since(newest) > maxAge {
			df.logger.Info("Deleting old temp dir", zap.String("dir", d.Name()))
			if err = os.RemoveAll(fullPath); err != nil {
				df.logger.Error("Error deleting old temp dir", zap.Error(err))
			}
		}
	}
}

func DirFS(dir string) {
	panic("unimplemented")
}

type userSpaces struct {
	root   string
	spaces map[string]*userSpace
	logger *zap.Logger
}

func newUserSpaces(root string, logger *zap.Logger) userSpaces {

	return userSpaces{
		root:   root,
		logger: logger,
		spaces: map[string]*userSpace{},
	}
}

func (us userSpaces) Get(user string) *userSpace {
	s, ok := us.spaces[user]

	if ok {
		return s
	}

	s = newUserSpace(user, path.Join(us.root, user))
	us.logger.Info("Creating new user space", zap.String("user", user), zap.String("root", s.root))
	us.spaces[user] = s
	return s
}

func (factory *driverFactory) NewDriver() (ftps.Driver, error) {
	return newDriver(factory.logger, factory.bus, factory.userSpaces), nil
}

type proxyDriver struct {
	sync.Mutex
	conn        *ftps.Conn
	logger      *zap.Logger
	bus         common.EventBus
	userDriver  ftps.Driver
	userSpace   *userSpace
	usersSpaces userSpaces
}

func newDriver(logger *zap.Logger, bus common.EventBus, userSpaces userSpaces) ftps.Driver {
	fd := &proxyDriver{
		logger:      logger,
		bus:         bus,
		usersSpaces: userSpaces,
	}
	return fd
}

type userSpace struct {
	driver *fd.FileDriver
	user   string
	root   string
}

func newUserSpace(u string, r string) *userSpace {

	err := os.MkdirAll(r, os.ModeDir|os.ModePerm)

	if err != nil {
		panic(fmt.Sprintf("Failed to create user dir (%s): %v", r, err))
	}

	return &userSpace{
		user: u,
		root: r,
		driver: &fd.FileDriver{
			RootPath: r,
			Perm:     ftps.NewSimplePerm("owner", "group"),
		},
	}
}

func (us userSpace) getPath(p string) string {
	return p
}

func (us userSpace) stat(p string) (os.FileInfo, error) {
	fullPath := path.Join(us.root, p)

	return os.Stat(fullPath)
}

func (us userSpace) getBytes(p string) ([]byte, error) {
	fullPath := path.Join(us.root, p)

	return ioutil.ReadFile(fullPath)
}

func (us userSpace) getReader(p string) (io.ReadCloser, error) {
	fullPath := path.Join(us.root, p)

	return os.Open(fullPath)

}

func (us *userSpace) CreateDriver() ftps.Driver {
	return us.driver
}

func (fd *proxyDriver) Init(c *ftps.Conn) {
	fd.logger.Debug("Connection initiated")
	fd.conn = c
}

func (fd *proxyDriver) driver() ftps.Driver {
	fd.Lock()
	defer fd.Unlock()
	if fd.userDriver == nil {
		user := fd.conn.LoginUser()
		fd.userSpace = fd.usersSpaces.Get(user)
		fd.userDriver = fd.userSpace.CreateDriver()
	}
	return fd.userDriver
}

const cleanupTime = time.Hour * 1

// params  - a file path
// returns - a time indicating when the requested path was last modified
//   - an error if the file doesn't exist or the user lacks
//     permissions
func (fd *proxyDriver) Stat(p string) (ftps.FileInfo, error) {
	fd.logger.Debug("STAT", zap.String("path", p))
	return fd.driver().Stat(p)
}

// params  - path
// returns - true if the current user is permitted to change to the
//
//	requested path
func (fd *proxyDriver) ChangeDir(p string) error {
	fd.logger.Debug("CWD", zap.String("path", p))

	return fd.driver().ChangeDir(p)
}

// params  - path, function on file or subdir found
// returns - error
//
//	path
func (fd *proxyDriver) ListDir(p string, r func(ftps.FileInfo) error) error {
	fd.logger.Debug("LIST", zap.String("path", p))
	return fd.driver().ListDir(p, r)
}

// params  - path
// returns - nil if the directory was deleted or any error encountered
func (fd *proxyDriver) DeleteDir(p string) error {
	fd.logger.Debug("RMDIR", zap.String("path", p))
	return fd.driver().DeleteDir(p)
}

// params  - path
// returns - nil if the file was deleted or any error encountered
func (fd *proxyDriver) DeleteFile(p string) error {
	fd.logger.Debug("RM", zap.String("path", p))
	return fd.driver().DeleteFile(p)
}

// params  - from_path, to_path
// returns - nil if the file was renamed or any error encountered
func (fd *proxyDriver) Rename(s string, d string) error {
	fd.logger.Debug("REN", zap.String("path", s), zap.String("dest", d))

	srcFile, err := fd.toFtpFile(s)
	if err != nil {
		return err
	}

	err = fd.driver().Rename(s, d)
	if err != nil {
		return err
	}

	destFile, err := fd.toFtpFile(d)
	if err != nil {
		return err
	}

	fd.bus.Send(NewFileRenameEvent(destFile, srcFile.FullName))

	return nil
}

// params  - path
// returns - nil if the new directory was created or any error encountered
func (fd *proxyDriver) MakeDir(p string) error {
	fd.logger.Debug("MKDIR", zap.String("path", p))
	return fd.driver().MakeDir(p)
}

// params  - path
// returns - a string containing the file data to send to the client
func (fd *proxyDriver) GetFile(p string, n int64) (int64, io.ReadCloser, error) {
	fd.logger.Error("GET", zap.String("path", p))
	return 0, nil, os.ErrInvalid
}

// params  - destination path, an io.Reader containing the file data
// returns - the number of bytes writen and the first error encountered while writing, if any.
func (fd *proxyDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {

	fd.logger.Debug("PUT", zap.String("path", destPath))

	n, err := fd.driver().PutFile(destPath, data, appendData)

	if err != nil {
		fd.logger.Error("Error putting file", zap.Error(err), zap.String("path", destPath))
		return n, err
	}

	f, err := fd.toFtpFile(destPath)
	if err != nil {
		return 0, err
	}

	fd.logger.Debug("FTP Received file", zap.String("user", fd.conn.LoginUser()), zap.String("path", f.FullName))

	fd.bus.Send(NewFileCreateEvent(f))

	return int64(n), nil
}

func (fd *proxyDriver) toFtpFile(p string) (*File, error) {

	fullPath := fd.userSpace.getPath(p)
	info, err := fd.userSpace.stat(p)
	if err != nil {
		return nil, err
	}

	bytes, err := fd.userSpace.getBytes(p)
	if err != nil {
		return nil, err
	}

	f := &File{
		User:       fd.conn.LoginUser(),
		FullName:   fullPath,
		Data:       bytes,
		Name:       path.Base(fullPath),
		IP:         fd.conn.PublicIp(),
		ReceivedAt: info.ModTime(),
		fullPath:   path.Join(fd.userSpace.root, fullPath),
		logger:     fd.logger,
	}

	f.AutoClose(cleanupTime)
	return f, nil
}
