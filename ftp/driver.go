package ftp

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"go.uber.org/zap"

	filedriver "github.com/goftp/file-driver"
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

type fileDriverFactory struct {
	ftps.Perm
	logger     *zap.Logger
	bus        common.EventBus
	userSpaces userSpaces
}

func newFileDriverFactory(perm ftps.Perm, logger *zap.Logger, bus common.EventBus) *fileDriverFactory {
	return &fileDriverFactory{
		Perm:       perm,
		logger:     logger,
		bus:        bus,
		userSpaces: newUserSpaces(path.Join(os.TempDir(), "cams"), logger),
	}
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
	us.logger.Info("Creating new space", zap.String("user", user), zap.String("root", s.root))
	us.spaces[user] = s
	return s
}

func (factory *fileDriverFactory) NewDriver() (ftps.Driver, error) {
	return newFileDriver(factory.logger, factory.bus, factory.userSpaces), nil
}

type fileDriver struct {
	conn        *ftps.Conn
	logger      *zap.Logger
	bus         common.EventBus
	driver      ftps.Driver
	userSpace   *userSpace
	usersSpaces userSpaces
}

func newFileDriver(logger *zap.Logger, bus common.EventBus, userSpaces userSpaces) ftps.Driver {
	fd := &fileDriver{
		logger:      logger,
		bus:         bus,
		usersSpaces: userSpaces,
	}
	return fd
}

type userSpace struct {
	user string
	root string
}

func newUserSpace(u string, r string) *userSpace {

	err := os.MkdirAll(r, os.ModeDir|os.ModePerm)

	if err != nil {
		panic(fmt.Sprintf("Failed to create user dir (%s): %v", r, err))
	}

	return &userSpace{
		user: u,
		root: r,
	}
}

func (us userSpace) getPath(p string) string {
	return p
}

func (us userSpace) stat(p string) os.FileInfo {
	fullPath := path.Join(us.root, p)

	info, err := os.Stat(fullPath)
	if err != nil {
		// TODO: log
		fmt.Fprintf(os.Stderr, "ERROR stating file %s: %v", fullPath, err)
		return nil
	}
	return info
}

func (us userSpace) getBytes(p string) []byte {
	fullPath := path.Join(us.root, p)

	bytes, err := ioutil.ReadFile(fullPath)

	if err != nil {
		// todo logging
		fmt.Fprintf(os.Stderr, "ERROR reading file %s: %v", fullPath, err)
	}

	return bytes
}

func (us userSpace) getReader(p string) io.ReadCloser {
	fullPath := path.Join(us.root, p)

	reader, err := os.Open(fullPath)

	if err != nil {
		// todo logging
		fmt.Fprintf(os.Stderr, "ERROR reading file %s: %v", fullPath, err)
	}

	return reader

}

func (us *userSpace) CreateDriver(conn *ftps.Conn) ftps.Driver {
	return &filedriver.FileDriver{
		RootPath: us.root,
		Perm:     ftps.NewSimplePerm("owner", "group"),
	}
}

func (fd *fileDriver) Init(c *ftps.Conn) {
	fd.logger.Info("Connection initiated", zap.String("user", c.LoginUser()))

	user := c.LoginUser()
	fd.userSpace = fd.usersSpaces.Get(user)
	fd.driver = fd.userSpace.CreateDriver(c)
	fd.conn = c

}

const cleanupTime = time.Minute * 5

// params  - a file path
// returns - a time indicating when the requested path was last modified
//         - an error if the file doesn't exist or the user lacks
//           permissions
func (fd *fileDriver) Stat(p string) (ftps.FileInfo, error) {
	fd.logger.Info("STAT", zap.String("path", p))
	return fd.driver.Stat(p)
}

// params  - path
// returns - true if the current user is permitted to change to the
//           requested path
func (fd *fileDriver) ChangeDir(p string) error {
	fd.logger.Info("CWD", zap.String("path", p))

	return fd.driver.ChangeDir(p)
}

// params  - path, function on file or subdir found
// returns - error
//           path
func (fd *fileDriver) ListDir(p string, r func(ftps.FileInfo) error) error {
	fd.logger.Info("LIST", zap.String("path", p))
	return fd.driver.ListDir(p, r)
}

// params  - path
// returns - nil if the directory was deleted or any error encountered
func (fd *fileDriver) DeleteDir(p string) error {
	fd.logger.Info("RMDIR", zap.String("path", p))
	return fd.driver.DeleteDir(p)
}

// params  - path
// returns - nil if the file was deleted or any error encountered
func (fd *fileDriver) DeleteFile(p string) error {
	fd.logger.Info("RM", zap.String("path", p))
	return fd.driver.DeleteFile(p)
}

// params  - from_path, to_path
// returns - nil if the file was renamed or any error encountered
func (fd *fileDriver) Rename(s string, d string) error {
	fd.logger.Info("REN", zap.String("path", s), zap.String("dest", d))

	srcFile := fd.toFtpFile(s)

	err := fd.driver.Rename(s, d)
	if err != nil {
		return err
	}

	destFile := fd.toFtpFile(d)

	fd.bus.Send(NewFileRenameEvent(destFile, srcFile.FullName))

	return nil
}

// params  - path
// returns - nil if the new directory was created or any error encountered
func (fd *fileDriver) MakeDir(p string) error {
	fd.logger.Info("MKDIR", zap.String("path", p))
	return fd.driver.MakeDir(p)
}

// params  - path
// returns - a string containing the file data to send to the client
func (fd *fileDriver) GetFile(p string, n int64) (int64, io.ReadCloser, error) {
	fd.logger.Error("GET", zap.String("path", p))

	// return driver.GetFile(p, n)
	return 0, nil, os.ErrInvalid
}

// params  - destination path, an io.Reader containing the file data
// returns - the number of bytes writen and the first error encountered while writing, if any.
func (fd *fileDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {

	fd.logger.Info("PUT", zap.String("path", destPath))

	n, err := fd.driver.PutFile(destPath, data, appendData)

	if err != nil {
		fd.logger.Error("Error putting file", zap.Error(err), zap.String("path", destPath))
		return n, err
	}

	fd.bus.Send(NewFileCreateEvent(fd.toFtpFile(destPath)))

	return int64(n), nil
}

func (fd *fileDriver) toFtpFile(p string) *File {

	fullPath := fd.userSpace.getPath(p)
	info := fd.userSpace.stat(p)

	return &File{
		User:       fd.conn.LoginUser(),
		FullName:   fullPath,
		Data:       fd.userSpace.getBytes(p),
		Name:       path.Base(fullPath),
		IP:         fd.conn.PublicIp(),
		ReceivedAt: info.ModTime(),
	}
}

type ftpFile struct {
	dir      string
	name     string
	data     []byte
	mode     os.FileMode
	isDir    bool
	ts       time.Time
	conn     ftps.Conn
	fullPath string

	owner, group string
}

func (fi *ftpFile) asFile() *File {
	fullPath := path.Join(fi.dir, fi.name)
	f := &File{
		Name:       path.Base(fullPath),
		FullName:   fullPath,
		Data:       fi.data,
		User:       fi.conn.LoginUser(),
		IP:         fi.conn.PublicIp(),
		ReceivedAt: fi.ts,
	}
	return f
}

func (fi *ftpFile) Name() string {
	return fi.name
}

func (fi *ftpFile) Size() int64 {
	return int64(len(fi.data))
}

func (fi *ftpFile) Mode() os.FileMode {
	return fi.mode
} // file mode bits
func (fi *ftpFile) ModTime() time.Time {
	return fi.ts
}

func (fi *ftpFile) IsDir() bool {
	return fi.isDir
} // abbreviation for Mode().IsDir()
func (fi *ftpFile) Sys() interface{} {
	return nil
} // underlying data source (can return nil)

func (fi *ftpFile) Owner() string {
	return fi.owner
}

func (fi *ftpFile) Group() string {
	return fi.group
}
