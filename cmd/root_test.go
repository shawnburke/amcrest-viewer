package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"
)

func TestGraph(t *testing.T) {

	os.Chdir("../")
	app := fxtest.New(t, buildGraph())
	app.RequireStart().RequireStop()

	require.NotNil(t, app)

}
