package common

import (
	"testing"

	"github.com/stretchr/testify/require"
	fxcfg "go.uber.org/config"
)

func TestConfigAuth(t *testing.T) {

	cfgYaml := `
ftp:
  users:
    user_a: pwd_a
    user_B: pwd_b
`

	provider, err := fxcfg.NewYAMLProviderFromBytes([]byte(cfgYaml))
	require.NoError(t, err)
	auth, err := NewConfigAuth(provider)
	require.NoError(t, err)

	require.True(t, auth.IsAllowed("user_a", "pwd_a"))

	require.True(t, auth.IsAllowed("User_A", "pwd_a"))

	require.False(t, auth.IsAllowed("user_x", "pwd_a"))

	require.False(t, auth.IsAllowed("user_a", "pwd_x"))
}
