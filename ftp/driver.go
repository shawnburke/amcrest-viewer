package ftp

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"go.uber.org/zap"

	ftps "github.com/goftp/server"
	"github.com/shawnburke/amcrest-viewer/common"
)

type fileDriverFactory struct {
	RootPath string
	ftps.Perm
	logger *zap.Logger
	bus    common.EventBus
}

func (factory *fileDriverFactory) NewDriver() (ftps.Driver, error) {
	return newFileDriver(factory.logger, factory.bus), nil
}

type fileDriver struct {
	sync.Mutex
	conn   *ftps.Conn
	cwd    string
	files  map[string]*ftpFile
	logger *zap.Logger
	bus    common.EventBus
}

func newFileDriver(logger *zap.Logger, bus common.EventBus) ftps.Driver {
	fd := &fileDriver{
		logger: logger,
		cwd:    "/",
		files: map[string]*ftpFile{
			"/": {
				name:  "",
				isDir: true,
				dir:   "/",
			},
		},
		bus: bus,
	}

	return fd

}

func (fd *fileDriver) Init(c *ftps.Conn) {
	fd.logger.Info("Connection initiated", zap.String("user", c.LoginUser()))
	fd.conn = c
}

const cleanupTime = time.Minute * 5

func (fd *fileDriver) fullPath(p string) string {
	fullPath := path.Join(fd.cwd, p)
	return fullPath
}

// params  - a file path
// returns - a time indicating when the requested path was last modified
//         - an error if the file doesn't exist or the user lacks
//           permissions
func (fd *fileDriver) Stat(p string) (ftps.FileInfo, error) {
	fd.logger.Info("STAT", zap.String("path", p))
	f, ok := fd.files[fd.fullPath(p)]
	if !ok {
		return nil, os.ErrNotExist
	}
	return f, nil
}

// params  - path
// returns - true if the current user is permitted to change to the
//           requested path
func (fd *fileDriver) ChangeDir(p string) error {
	fd.logger.Info("CWD", zap.String("path", p))
	fd.cwd = p
	return nil
}

// params  - path, function on file or subdir found
// returns - error
//           path
func (fd *fileDriver) ListDir(p string, r func(ftps.FileInfo) error) error {

	for filePath, fi := range fd.files {
		dir := path.Dir(filePath)
		if dir != p || filePath == p {
			continue
		}

		err := r(fi)
		if err != nil {
			return err
		}
	}
	return nil
}

// params  - path
// returns - nil if the directory was deleted or any error encountered
func (fd *fileDriver) DeleteDir(p string) error {
	fd.logger.Panic("RMDIR", zap.String("path", p))
	return os.ErrInvalid
}

// params  - path
// returns - nil if the file was deleted or any error encountered
func (fd *fileDriver) DeleteFile(p string) error {
	fd.logger.Panic("RM", zap.String("path", p))
	return os.ErrInvalid
}

// params  - from_path, to_path
// returns - nil if the file was renamed or any error encountered
func (fd *fileDriver) Rename(s string, d string) error {
	fd.logger.Info("REN", zap.String("path", s), zap.String("dest", d))
	srcPath := fd.fullPath(s)

	file, ok := fd.files[srcPath]
	if !ok {
		return os.ErrNotExist
	}

	destPath := fd.fullPath(d)
	if srcPath == destPath {
		return nil
	}

	file.dir = path.Dir(destPath)
	file.name = path.Base(destPath)
	fd.files[destPath] = file
	delete(fd.files, srcPath)

	fd.bus.Send(common.NewFileRenameEvent(file.asFile(), srcPath))

	return nil
}

// params  - path
// returns - nil if the new directory was created or any error encountered
func (fd *fileDriver) MakeDir(p string) error {
	fd.logger.Info("MKDIR", zap.String("path", p))
	fullPath := fd.fullPath(p)

	if f, ok := fd.files[fullPath]; ok {
		if !f.isDir {
			return os.ErrInvalid
		}
		return nil
	}

	fd.files[fullPath] = &ftpFile{
		dir:   path.Dir(fullPath),
		name:  path.Base(fullPath),
		isDir: true,
		ts:    time.Now(),
	}
	return nil
}

// params  - path
// returns - a string containing the file data to send to the client
func (fd *fileDriver) GetFile(p string, n int64) (int64, io.ReadCloser, error) {
	fd.logger.Error("GET", zap.String("path", p))

	return 0, nil, os.ErrInvalid
}

// params  - destination path, an io.Reader containing the file data
// returns - the number of bytes writen and the first error encountered while writing, if any.
func (fd *fileDriver) PutFile(destPath string, data io.Reader, appendData bool) (int64, error) {

	fd.logger.Info("PUT", zap.String("path", destPath))

	fullPath := fd.fullPath(destPath)

	file, ok := fd.files[fullPath]

	if ok && appendData && file.isDir {
		return 0, os.ErrInvalid
	}

	if !ok || !appendData {
		file = &ftpFile{
			dir:  path.Dir(fullPath),
			name: path.Base(fullPath),
			user: fd.conn.LoginUser(),
		}
	}
	fd.Lock()
	fd.files[fullPath] = file
	fd.Unlock()

	cleanup := func() {
		fd.Lock()
		delete(fd.files, fullPath)
		fd.Unlock()
	}

	go func() {
		time.Sleep(cleanupTime)
		cleanup()
	}()

	file.ts = time.Now()

	buffer := make([]byte, 1024*1024)
	read := 0
	for sz, err := data.Read(buffer); sz > 0; sz, err = data.Read(buffer) {
		if err != nil {
			return int64(read), fmt.Errorf("Error reading data: %v", err)
		}
		read += sz
		file.data = append(file.data, buffer[0:sz]...)
	}

	if fd.bus != nil {
		f := file.asFile()
		fd.bus.Send(common.NewFileCreateEvent(f))
	}

	return int64(read), nil
}

type ftpFile struct {
	dir   string
	name  string
	data  []byte
	mode  os.FileMode
	isDir bool
	ts    time.Time
	user  string

	owner, group string
}

func (fi *ftpFile) asFile() *common.File {
	fullPath := path.Join(fi.dir, fi.name)
	f := &common.File{
		Name:       path.Base(fullPath),
		FullName:   fullPath,
		Data:       fi.data,
		User:       fi.user,
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
