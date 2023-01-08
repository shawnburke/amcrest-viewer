package ftp

import (
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
	Data       []byte
	Name       string
	FullName   string
	ReceivedAt time.Time
	fullPath   string
	logger     *zap.Logger
}

func (f *File) Close() {
	f.Lock()
	defer f.Unlock()
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
