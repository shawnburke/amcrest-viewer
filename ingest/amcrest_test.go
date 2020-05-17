package ingest

import (
	"fmt"
	"path"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/ftp"
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

	ingester, err := Amcrest(AmcrestParams{
		TZ:     tz,
		Logger: zap.NewNop(),
	})
	require.NoError(t, err)

	cases := []struct {
		P        string
		TS       time.Time
		Type     models.MediaFileType
		Duration time.Duration
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
	}

	for i, tt := range cases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			f := &ftp.File{
				User:     "amcrest",
				Data:     []byte("abc"),
				Name:     path.Base(tt.P),
				FullName: tt.P,
			}

			mf := ingester.Ingester.IngestFile(f)
			require.Equal(t, tt.Type, mf.Type)
			require.Equal(t, tt.TS, mf.Timestamp)
			if mf.Type == models.MP4 {
				require.Equal(t, tt.Duration, *mf.Duration)
			}
		})
	}

}
