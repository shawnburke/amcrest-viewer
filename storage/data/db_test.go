package data

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func TestNewDB(t *testing.T) {

	db, _, done := createDB(t)
	defer done()

	// ensure the constraint is right
	_, err := db.Exec(`INSERT INTO files (Path, Timestamp, Received, CameraID, Type) VALUES ("/foo/bar", "2020-05-14T12:34:32", "2020-05-14T12:34:32", 23, 0)`)
	require.Error(t, err)

}

func TestCamAddGetList(t *testing.T) {
	_, rep, done := createDB(t)
	defer done()

	host := "1.2.34.4"

	cams := []*entities.Camera{
		{
			Name: time.Now().String(),
			Type: "amcrest",
			Host: &host,
		},
		{
			Name: "foobar",
			Type: "amcrest",
		},
		{
			Name: "foobar",
			Type: "fail",
		},
	}

	for i, cam := range cams {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			cam2, err := rep.AddCamera(cam.Name, cam.Type, cam.Host)
			if cam.Type == "fail" {
				require.Nil(t, cam2)
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cam2)

			require.Greater(t, cam2.ID, 0)
			require.Equal(t, cam.Name, cam2.Name)
			require.Equal(t, cam.Type, cam2.Type)
		})
	}

	res, err := rep.ListCameras()
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 2)

	require.Equal(t, *cams[0].Host, *res[0].Host)
	require.Equal(t, cams[1].Name, res[1].Name)
}

func TestCamAddUpdateDelete(t *testing.T) {
	_, rep, done := createDB(t)
	defer done()

	host := "1.2.34.4"

	cam := &entities.Camera{
		Name: "foobar",
		Type: "amcrest",
		Host: &host,
	}

	cam2, err := rep.AddCamera(cam.Name, cam.Type, cam.Host)

	require.NoError(t, err)
	require.NotNil(t, cam2)

	newName := "newbar"
	cam3, err := rep.UpdateCamera(cam2.CameraID(), &newName, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, cam3)

	require.Equal(t, newName, cam3.Name)

	newHost := "1.1.1.1"
	enabled := false
	cam3, err = rep.UpdateCamera(cam2.CameraID(), nil, &newHost, &enabled)
	require.NoError(t, err)
	require.NotNil(t, cam3)

	require.Equal(t, newHost, *cam3.Host)
	require.Equal(t, false, *cam3.Enabled)

	found, err := rep.DeleteCamera(cam3.CameraID())
	require.True(t, found)
	require.NoError(t, err)

	cam3, err = rep.GetCamera(cam2.CameraID())
	require.Equal(t, os.ErrNotExist, err)
	require.Nil(t, cam3)
}

func TestCamLastSeen(t *testing.T) {
	_, rep, done := createDB(t)
	defer done()

	cam := &entities.Camera{
		Name: "foobar",
		Type: "amcrest",
	}

	cam2, err := rep.AddCamera(cam.Name, cam.Type, nil)

	require.NoError(t, err)
	require.NotNil(t, cam2)
	require.Nil(t, cam2.LastSeen)

	err = rep.SeenCamera(cam2.CameraID())
	require.NoError(t, err)

	cam2, err = rep.GetCamera(cam2.CameraID())
	require.NoError(t, err)
	require.NotNil(t, cam2)
	require.NotNil(t, cam2.LastSeen)
}

func TestFileAddGetList(t *testing.T) {
	_, rep, done := createDB(t)
	defer done()

	cam, err := rep.AddCamera("mycam", "amcrest", nil)
	require.NoError(t, err)
	require.NotNil(t, cam)

	start := time.Now()

	// add 10 files
	for i := 0; i < 10; i++ {
		ts := start.Add(time.Minute * time.Duration(i))
		d := time.Second * time.Duration(i)
		file, err := rep.AddFile(fmt.Sprintf("root/file-%d.mp4", i), entities.FileTypeMp4, cam.CameraID(), 110+i, ts, &d)
		require.NoError(t, err)
		require.NotNil(t, file)
	}

	res, err := rep.ListFiles(cam.CameraID(), nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 10)
	require.Equal(t, 112, res[2].Length)

	s := start.Add(time.Minute)
	e := start.Add(time.Minute * 5)
	lff := &ListFilesFilter{
		Start: &s,
		End:   &e,
	}
	res, err = rep.ListFiles(cam.CameraID(), lff)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 4)

	fileType := entities.FileTypeJpg
	lff = &ListFilesFilter{
		FileType: &fileType,
	}

	res, err = rep.ListFiles(cam.CameraID(), lff)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res, 0)
}

func dumpTables(db *sqlx.DB) {
	fmt.Println("Tables\n", "========")
	res, _ := db.Queryx("SELECT name FROM sqlite_master WHERE type='table'")

	for res.Next() {
		name := ""
		res.Scan(&name)
		fmt.Println(name)
	}
	res.Close()
}

var debugDb = false

func createDB(t *testing.T) (*sqlx.DB, Repository, func()) {

	cfg := &DBConfig{
		DSN: fmt.Sprintf("file:%s-%d.sqlite?cache=shared", t.Name(), time.Now().Unix()),
	}

	if os.Getenv("DEBUGDB") != "" {
		debugDb = true
	}
	if !debugDb {
		cfg.DSN += "&mode=memory"
	}

	lc := &testLifecycle{}

	db, err := New(cfg, lc)
	require.NoError(t, err)

	err = lc.lc.OnStart(context.Background())
	require.NoError(t, err)

	rep, err := NewRepository(db, zap.NewNop())
	require.NoError(t, err)
	require.NotNil(t, rep)

	return db, rep, func() {
		err := lc.lc.OnStop(context.Background())
		require.NoError(t, err)

		require.NoError(t, err)
	}
}

type testLifecycle struct {
	lc fx.Hook
}

func (tc *testLifecycle) Append(hook fx.Hook) {
	tc.lc = hook
}
