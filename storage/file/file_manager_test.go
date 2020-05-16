package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestFM(t *testing.T) {

	testdir := path.Join(os.TempDir(), fmt.Sprintf("camtest-%d", time.Now().Unix()))
	defer os.RemoveAll(testdir)
	fm, err := New(zap.NewNop(), &Config{RootDir: testdir})

	require.NoError(t, err)
	require.NotNil(t, fm)

	start := createFiles(t, fm, 25, "camera10")

	s := start.Add(time.Minute)
	e := start.Add(time.Minute * time.Duration(7))
	ft := entities.FileTypeJpg
	files, err := fm.ListFiles("camera10", &s, &e, &ft)

	require.NoError(t, err)
	require.Len(t, files, 6)

}

func TestDeleteFiles(t *testing.T) {
	testdir := path.Join(os.TempDir(), fmt.Sprintf("camtest-%d", time.Now().Unix()))
	defer os.RemoveAll(testdir)
	fm, err := New(zap.NewNop(), &Config{RootDir: testdir})

	start := createFiles(t, fm, 10, "camera11")

	s := start.Add(time.Duration(2) * time.Minute)
	e := s.Add(time.Duration(3) * time.Minute)
	list, err := fm.DeleteFiles("camera11", &s, &e)
	require.NoError(t, err)
	require.Len(t, list, 3)

	remaining, err := fm.ListFiles("camera11", nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, remaining, 7)

	// delete newer than s
	list, err = fm.DeleteFiles("camera11", nil, &s)
	require.NoError(t, err)
	require.Len(t, list, 3)

	remaining, err = fm.ListFiles("camera11", nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, remaining, 4)

}

func createFiles(t *testing.T, fm Manager, n int, camera string) time.Time {
	start := time.Now()
	for i := 0; i < n; i++ {
		data := []byte(fmt.Sprintf("File %d", i))

		p, err := fm.AddFile(camera, data, start.Add(time.Minute*time.Duration(i)), entities.FileTypeJpg)

		require.NoError(t, err)
		require.NotEmpty(t, p)

		reader, err := fm.GetFile(p)
		require.NoError(t, err)

		fileData, err := ioutil.ReadAll(reader)

		require.NoError(t, err)
		defer reader.Close()
		require.Equal(t, data, fileData)
	}
	return start
}
