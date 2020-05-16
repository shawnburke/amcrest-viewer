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

	start := time.Now()
	for i := 0; i < 25; i++ {
		data := []byte(fmt.Sprintf("File %d", i))

		p, err := fm.AddFile("camera10", data, start.Add(time.Minute*time.Duration(i)), entities.FileTypeJpg)

		require.NoError(t, err)
		require.NotEmpty(t, p)

		reader, err := fm.GetFile(p)
		require.NoError(t, err)

		fileData, err := ioutil.ReadAll(reader)

		require.NoError(t, err)
		defer reader.Close()
		require.Equal(t, data, fileData)
	}

	s := start.Add(time.Minute)
	e := start.Add(time.Minute * time.Duration(7))
	ft := entities.FileTypeJpg
	files, err := fm.ListFiles("camera10", &s, &e, &ft)

	require.NoError(t, err)
	require.Len(t, files, 6)

}
