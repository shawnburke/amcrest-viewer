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

func TestOnNewMP4(t *testing.T) {
	path := "2019-05-09/001/dav/21/21.04.49-21.05.14[M][0@0][0].mp4"

	ingester, err := New(tz)
	require.NoError(t, err)

	f, err := ingester.OnNewFile(path)

	start := time.Date(2019, time.May, 9, 21, 04, 49, 0, tz)

	require.NoError(t, err)
	require.Equal(t, MP4, f.Type)
	require.Equal(t, f.Timestamp, start)

	require.Equal(t, f.Duration.Seconds(), 25.0)
}

func TestOnNewJPG(t *testing.T) {
	path := "2019-05-09/001/jpg/06/07/27[M][0@0][0].jpg"

	ingester, err := New(tz)
	require.NoError(t, err)

	f, err := ingester.OnNewFile(path)

	time := time.Date(2019, time.May, 9, 06, 07, 27, 0, tz)

	require.NoError(t, err)
	require.Equal(t, JPG, f.Type)
	require.Equal(t, f.Timestamp, time)

	require.Nil(t, f.Duration)
}
