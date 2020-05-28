package web

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContentType(t *testing.T) {

	ct := getContentType("12343254124.mp4")

	require.Equal(t, "video/mp4", ct)

}