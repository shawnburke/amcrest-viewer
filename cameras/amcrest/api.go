package amcrest

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/icholy/digest"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"go.uber.org/zap"
)

const authTypeBasic = "basic"
const authTypeDigest = "digest"
const authTypeFailed = "failed"

type amcrestApi struct {
	*entities.Camera
	Port     int
	client   *http.Client
	authType string
	logger   *zap.Logger
}

func newAmcrestApi(host string, user string, pass string, logger *zap.Logger) *amcrestApi {
	aa := &amcrestApi{
		Camera: &entities.Camera{
			Host: &host,
			CameraCreds: entities.CameraCreds{
				Username: &user,
				Password: &pass,
			},
		},
		logger: logger,
	}

	return aa

}

func (aa *amcrestApi) buildUrl(command string) string {
	port := 80

	if aa.Port != 0 {
		port = aa.Port
	}
	return fmt.Sprintf("%s://%s:%d/cgi-bin/%s", "http", *aa.Host, port, command)
}

func (aa *amcrestApi) getClient() *http.Client {
	if aa.client == nil {
		aa.client = &http.Client{}
	}
	return aa.client
}

func (aa *amcrestApi) getRequest(method string, command string) (*http.Request, error) {

	url := aa.buildUrl(command)

	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	if aa.authType == authTypeBasic {
		req.SetBasicAuth(*aa.Username, *aa.Password)
	}
	return req, nil
}

func (aa *amcrestApi) ensureAuth() error {

	if aa.authType == "" {

		authTypes := []string{authTypeDigest, authTypeBasic}

		for _, authType := range authTypes {
			aa.authType = authType
			client := aa.getClient()
			switch aa.authType {
			case authTypeDigest:

				client.Transport = &digest.Transport{
					Username: *aa.Username,
					Password: *aa.Password,
				}
			default:
				client.Transport = nil
			}

			resp, err := aa.executeCore("GET", "magicBox.cgi?action=getMachineName")
			if err != nil {
				return err
			}

			if resp.StatusCode == 200 {
				aa.logger.Info("Found auth", zap.String("type", authType))
				return nil
			}
			aa.logger.Warn("Failed auth",
				zap.String("type", authType),
				zap.String("status", resp.Status),
				zap.Int("status-code", resp.StatusCode))
		}
		aa.authType = authTypeFailed
		return fmt.Errorf("Failed to find any auth method")
	}
	return nil
}

func (aa *amcrestApi) Execute(method string, command string) (*http.Response, error) {
	if err := aa.ensureAuth(); err != nil {
		return nil, err
	}
	return aa.executeCore(method, command)
}

func (aa *amcrestApi) ExecuteString(method string, command string) (string, error) {
	if err := aa.ensureAuth(); err != nil {
		return "", err
	}
	resp, err := aa.executeCore(method, command)

	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		return string(bytes), nil
	}
	return "", fmt.Errorf("Response not successful: %d %s", resp.StatusCode, resp.Status)
}

func (aa *amcrestApi) executeCore(method string, command string) (*http.Response, error) {
	req, err := aa.getRequest(method, command)
	if err != nil {
		return nil, err
	}
	return aa.getClient().Do(req)
}
