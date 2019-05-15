package main

import (
	"testing"
	"fmt"
	"github.com/stretchr/testify/require"
)


func TestTimeStamps(t *testing.T) {

	p := "2019-05-10/001/dav/12/12.52.24-12.52.48[M][0@0][0].mp4"

	ts := pathToTimestampes(p)


	require.Len(t, ts, 2)

	require.Equal(t, "2019-05-10 12:52:24 -0700 PDT", ts[0].String())

}

func TestWalk(t *testing.T) {


	files, err := FindFiles("./data")

	require.NoError(t, err)

	require.NotEqual(t, 0, len(files))

	for _, d := range files {
		fmt.Println(d.Date, d.DayOfWeek)
		for _, f := range d.Files {
			fmt.Println(f.Path, f.StartTime, f.EndTime)
		}
	}

}