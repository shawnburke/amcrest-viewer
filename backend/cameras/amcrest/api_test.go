package amcrest

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEnsureAuth(t *testing.T) {
	user := os.Getenv("AMCREST_USER")
	if user == "" {
		user = "admin"
	}

	pass := os.Getenv("AMCREST_PASSWORD")
	host := os.Getenv("AMCREST_HOST")

	if pass == "" || host == "" {
		fmt.Println("Skipping amcrest test because AMCREST_HOST and AMCREST_PASSWORD are not set.")
		t.Skip()
		return
	}
	logger, _ := zap.NewDevelopment()

	aa := newAmcrestApi(host, user, pass, logger)

	resp, err := aa.ExecuteString("GET", "global.cgi?action=getCurrentTime")

	require.NoError(t, err)

	require.Contains(t, resp, "result=")

}
