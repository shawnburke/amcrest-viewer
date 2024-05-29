package ftp

import (
	"io"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// File should be renamed to FtpFile
type File struct {
	sync.Mutex
	io.ReadCloser
	User       string
	IP         string
	Name       string
	FullName   string
	ReceivedAt time.Time
	fullPath   string
	logger     *zap.Logger
}

func (f *File) Close() error {
	f.Lock()
	defer f.Unlock()

	if f.ReadCloser != nil {
		f.ReadCloser.Close()
	}

	if f.fullPath != "" {
		_, err := os.Stat(f.fullPath)
		if os.IsNotExist(err) {
			return nil
		}
		err = os.Remove(f.fullPath)
		if err != nil && f.logger != nil {
			f.logger.Error("Failed to clean up file", zap.String("path", f.fullPath), zap.Error(err))
		}
		f.fullPath = ""
	}
	return nil
}
