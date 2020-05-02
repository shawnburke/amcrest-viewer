package ingest

import (
	"testing"
	"time"
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

func TestOnNewVideo(t *testing.T) {
	path := "2019-05-09/001/dav/21/21.04.49-21.05.14[M][0@0][0].mp4"

	ingester, err := New(tz)
	require.NoError(t, err)

	f, err := ingester.OnNewFile(path)

	start := time.Date(2019, time.May,9,21,04,49,0,tz)

	require.NoError(t, err)
	require.Equal(t, f.Timestamp, start)

	require.Equal(t, f.Duration.Seconds(), 25.0)
}