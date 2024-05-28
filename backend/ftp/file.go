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
	User       string
	IP         string
	Name       string
	FullName   string
	Reader     io.Reader
	ReceivedAt time.Time
	fullPath   string
	logger     *zap.Logger
}

func (f *File) Close() {
	f.Lock()
	defer f.Unlock()

	if f.Reader != nil {
		if closer, ok := f.Reader.(io.Closer); ok {
			closer.Close()
		}
		f.Reader = nil
	}

	if f.fullPath != "" {
		_, err := os.Stat(f.fullPath)
		if os.IsNotExist(err) {
			return
		}
		err = os.Remove(f.fullPath)
		if err != nil && f.logger != nil {
			f.logger.Error("Failed to clean up file", zap.String("path", f.fullPath), zap.Error(err))
		}
		f.fullPath = ""
	}
}

func (f *File) AutoClose(t time.Duration) {
	go func() {
		time.Sleep(t)
		f.Close()
	}()
}
