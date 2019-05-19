package main

import (
	"testing"
	"time"
	"fmt"
	"github.com/stretchr/testify/require"
)


func TestTimeStamps(t *testing.T) {

	p := "2019-05-10/001/dav/12/12.52.24-12.52.48[M][0@0][0].mp4"

	ts := pathToTimestamps(p)


	require.Len(t, ts, 2)

	require.Equal(t, "2019-05-10 12:52:24 -0700 PDT", ts[0].String())

}

func TestJpgTimeStamps(t *testing.T) {
	p := "2019-05-10/001/jpg/02/29/41[M][0@0][0].jpg"

	ts := jpgPathToTimestamp(p)
	exp, _ := time.Parse(time.RFC3339, "2019-05-10T02:29:41.000Z")
	require.Equal(t, exp, ts)
}

func TestWalk(t *testing.T) {


	files, err := FindFiles("./test_data", false)

	require.NoError(t, err)

	require.NotEqual(t, 0, len(files))

	for _, d := range files {
		for _, f := range d.Videos {
			fmt.Println(f.Path, f.Time, f.Duration, f.Thumb.Path, f.Thumb.Time)
			for _, i := range f.Images {
				fmt.Println("\t", i.Time, i.Path)
			}
		}
	}

}