package web

import (
	"testing"
	"time"

	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/stretchr/testify/require"
)

func TestContentType(t *testing.T) {

	ct := getContentType(&entities.File{
		Path: "cameras/amcrest-1/1590669408.mp4",
	})

	require.Equal(t, "video/mp4", ct)

}

func TestDateParse(t *testing.T) {

	s := &Server{}
	d, err := s.parseTime("1590908400000")
	require.NoError(t, err)
	require.Equal(t, "Sun, 31 May 2020 07:00:00 +0000", d.Format(time.RFC1123Z))

}
