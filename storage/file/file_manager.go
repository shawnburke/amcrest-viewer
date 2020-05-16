package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"go.uber.org/zap"
)

const subdir = "cameras"

type Manager interface {
	AddFile(camera string, data []byte, timestamp time.Time, fileType int) (string, error)
	GetFile(path string) (io.ReadCloser, error)
	ListFiles(camera string, start *time.Time, end *time.Time, fileType *int) ([]string, error)
	DeleteFile(path string) (bool, error)
	DeleteFiles(camera string, start *time.Time, end *time.Time) ([]string, error)
}

type Config struct {
	RootDir string
}

func New(logger *zap.Logger, cfg *Config) (Manager, error) {
	return &fileManager{
		logger:  logger,
		rootDir: cfg.RootDir,
	}, nil
}

type fileManager struct {
	logger  *zap.Logger
	rootDir string
}

func (fm *fileManager) getExtension(ft int) (string, error) {
	switch ft {
	case entities.FileTypeJpg:
		return "jpg", nil
	case entities.FileTypeMp4:
		return "mp4", nil
	default:
		return "", fmt.Errorf("Unknown file type %d", ft)
	}
}

func (fm *fileManager) getPath(camera string, ts time.Time, ft int) (string, error) {
	ext, err := fm.getExtension(ft)
	if err != nil {
		return "", err
	}
	fileName := fmt.Sprintf("%d.%s", ts.Unix(), ext)

	return path.Join(subdir, camera, fileName), nil
}

func (fm *fileManager) AddFile(camera string, data []byte, timestamp time.Time, fileType int) (string, error) {
	filePath, err := fm.getPath(camera, timestamp, fileType)
	if err != nil {
		return "", err
	}

	fullPath := path.Join(fm.rootDir, filePath)

	err = os.MkdirAll(path.Dir(fullPath), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Error creating dir %q: %w", path.Dir(fullPath), err)
	}

	err = ioutil.WriteFile(fullPath, data, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Error writing file: %w", err)
	}

	return filePath, nil
}

func (fm *fileManager) GetFile(p string) (io.ReadCloser, error) {
	fullPath := path.Join(fm.rootDir, p)

	if _, err := os.Stat(fullPath); err != nil {
		return nil, fmt.Errorf("Error reading file %q: %v", fullPath, err)
	}

	return os.Open(fullPath)
}

func (fm *fileManager) getRange(camera string, start *time.Time, end *time.Time, fileType *int) ([]string, error) {
	dir := path.Join(fm.rootDir, subdir, camera)

	pattern := "*.*"

	ext := ""
	var err error
	if fileType != nil {
		ext, err = fm.getExtension(*fileType)
		if err != nil {
			return nil, err
		}
		pattern = "*." + ext
	}
	files, err := filepath.Glob(path.Join(dir, pattern))
	if err != nil {
		return nil, fmt.Errorf("Error listing files in %q: %w", dir, err)
	}

	sort.Strings(files)

	matches := make([]string, 0, 32)

	rootDirLen := len(fm.rootDir)

	for _, f := range files {
		fileBase := path.Base(f)
		fileParts := strings.Split(fileBase, ".")

		if ext != "" && !strings.EqualFold(ext, fileParts[1]) {
			continue
		}

		if start != nil || end != nil {

			fileUnix, err := strconv.ParseInt(fileParts[0], 10, 64)
			if err != nil {
				continue
			}

			fileTS := time.Unix(fileUnix, 0)

			if start != nil && fileTS.Before(*start) {
				continue
			}

			if end != nil && fileTS.After(*end) {
				break
			}
		}
		matches = append(matches, f[rootDirLen:])

	}
	return matches, err
}

func (fm *fileManager) ListFiles(camera string, start *time.Time, end *time.Time, fileType *int) ([]string, error) {

	return fm.getRange(camera, start, end, fileType)

}

func (fm *fileManager) DeleteFile(p string) (bool, error) {
	fullPath := path.Join(fm.rootDir, p)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	err = os.Remove(fullPath)
	return err == nil, err
}

func (fm *fileManager) DeleteFiles(camera string, start *time.Time, end *time.Time) ([]string, error) {

	matches, err := fm.getRange(camera, start, end, nil)
	if err != nil {
		return nil, err
	}

	confirmed := make([]string, 0, len(matches))

	for _, match := range matches {
		fullPath := path.Join(fm.rootDir, match)
		err = os.Remove(fullPath)
		if err != nil {
			return confirmed, err
		}
		confirmed = append(confirmed, match)
	}
	return confirmed, nil
}
