package amcrest

import (
	"bytes"
	"fmt"
	"path"
	"testing"
	"time"

	"go.uber.org/config"
	"go.uber.org/zap"

	c "github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/ftp"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"github.com/stretchr/testify/require"
)

var tz *time.Location

func init() {
	var err error
	tz, err = time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
}

func TestIngestParse(t *testing.T) {

	ac, err := New(zap.NewNop(), tz, config.NopProvider{})
	require.NoError(t, err)

	ct := ac.Instance

	cases := []struct {
		P        string
		TS       time.Time
		Type     models.MediaFileType
		Duration time.Duration
		Fail     bool
		Error    error
	}{
		{
			P:        "2019-05-09/001/dav/21/21.04.49-21.05.14[M][0@0][0].mp4",
			Type:     models.MP4,
			Duration: time.Duration(25 * time.Second),
			TS:       time.Date(2019, time.May, 9, 21, 04, 49, 0, tz),
		},
		{
			P:    "2019-05-09/001/jpg/06/07/27[M][0@0][0].jpg",
			Type: models.JPG,
			TS:   time.Date(2019, time.May, 9, 06, 07, 27, 0, tz),
		},
		{
			P:    "/AMC0009L_M35704/2020-05-28/001/jpg/09/32/52[M][0@0][0].jpg",
			Type: models.JPG,
			TS:   time.Date(2020, time.May, 28, 9, 32, 52, 0, tz),
		},
		{
			P:     "/AMC0009L_M35704/2020-05-28/001/jpg/09/32/52[M][0@0][0].mp4_",
			Error: c.ErrIngestIgnore,
		},
		{
			P:     "/AMC0009L_M35704/2020-05-28/001/jpg/09/32/52[M][0@0][0].idx",
			Error: c.ErrIngestDelete,
		},
	}

	cam := &entities.Camera{
		Name:     "Test Cam",
		ID:       1,
		Type:     "amcrest",
		Timezone: "America/Los_Angeles",
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("Test %d: %s", i, tt.P), func(t *testing.T) {
			f := &ftp.File{
				User:     "amcrest-1",
				Reader:   bytes.NewReader([]byte("abc")),
				Name:     path.Base(tt.P),
				FullName: tt.P,
			}

			mf, err := ct.ParseFilePath(cam, f.FullName)

			if tt.Error != nil {
				require.Equal(t, tt.Error, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, mf)
			require.Equal(t, tt.Type, mf.Type)
			require.Equal(t, tt.TS, mf.Timestamp)
			if mf.Type == models.MP4 {
				require.Equal(t, tt.Duration, *mf.Duration)
			}
		})
	}

}
