package web

import (
	"testing"

	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/stretchr/testify/require"
)

func TestContentType(t *testing.T) {

	ct := getContentType(&entities.File{
		Path: "cameras/amcrest-1/1590669408.mp4",
	})

	require.Equal(t, "video/mp4", ct)

}
