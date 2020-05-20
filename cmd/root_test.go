package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/config"
	"go.uber.org/fx/fxtest"
)

func TestGraph(t *testing.T) {

	testConfig := `

ftp:
  password: pwd

files:
  root_dir: test_data/files

`
	provider, err := config.NewYAMLProviderFromBytes([]byte(testConfig))
	require.NoError(t, err)
	app := fxtest.New(t, buildGraph(provider))
	app.RequireStart().RequireStop()

	require.NotNil(t, app)

}
