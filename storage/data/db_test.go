package data

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/stretchr/testify/require"
	"go.uber.org/config"
	"go.uber.org/fx"
)

func TestNewDB(t *testing.T) {

	db, done := createDB(t)
	defer done()

	// ensure the constraint is right
	_, err := db.Exec(`INSERT INTO files (Path, Timestamp, Received, CameraID, Type) VALUES ("/foo/bar", "2020-05-14T12:34:32", "2020-05-14T12:34:32", 23, 0)`)
	require.Error(t, err)

}

func TestAddGet(t *testing.T) {
	db, done := createDB(t)
	defer done()

	rep, err := NewRepository(db)
	require.NoError(t, err)
	require.NotNil(t, rep)

	host := "1.2.34.4"

	cams := []*entities.Camera{
		{
			Name: "foobar",
			Type: "amcrest",
		},
		{
			Name: "foobar",
			Type: "fail",
		},
		{
			Name: time.Now().String(),
			Type: "amcrest",
			Host: &host,
		},
	}

	for i, cam := range cams {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			cam2, err := rep.AddCamera(cam.Name, cam.Type, cam.Host)

			if cam.Type == "fail" {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Greater(t, cam2.ID, 0)
			require.Equal(t, cam.Name, cam2.Name)
			require.Equal(t, cam.Type, cam2.Type)
		})
	}
}

const debugDb = false

func createDB(t *testing.T) (*sqlx.DB, func()) {

	yaml := `
database:
  dsn: ":memory:"
`

	if debugDb {
		dbName := fmt.Sprintf("testdb-%d.db", int(time.Now().Unix()/1000))
		yaml = strings.Replace(yaml, ":memory:", dbName, 1)

	}

	cfg, err := config.NewYAMLProviderFromBytes([]byte(yaml))

	require.NoError(t, err)
	lc := &testLifecycle{}

	db, err := NewDB(cfg, lc)
	require.NoError(t, err)

	err = lc.lc.OnStart(context.Background())
	require.NoError(t, err)

	return db, func() {
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
