package cameras

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/Roverr/rtsp-stream/core"
	"github.com/Roverr/rtsp-stream/core/config"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RtspServer interface {
	StreamPath(cameraID string) (string, error)
	Handle(strID string, path string, w http.ResponseWriter, r *http.Request) error
}

const rtspServerPort = 10099

func NewRtsp(reg Registry, data data.Repository, logger *zap.Logger, lifecycle fx.Lifecycle) (RtspServer, error) {
	r := &rtspServer{
		logger:         logger,
		cameraRegistry: reg,
		data:           data,
	}

	lifecycle.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				return r.start()
			},
			OnStop: func(ctx context.Context) error {
				return r.stop()
			},
		},
	)

	return r, nil
}

type rtspServer struct {
	httpServer     *http.Server
	cameraRegistry Registry
	data           data.Repository
	logger         *zap.Logger
	tempDir        string
	shuttingDown   int
}

func (r *rtspServer) start() error {
	//config := config.InitConfig()

	endpoints := config.EndpointYML{}
	endpoints.Endpoints.List.Enabled = true
	endpoints.Endpoints.Start.Enabled = true
	endpoints.Endpoints.Stop.Enabled = true
	endpoints.Endpoints.Static.Enabled = true

	config := &config.Specification{
		Debug: true,
		Port:  rtspServerPort,
		CORS: config.CORS{
			Enabled: true,
		},
		EndpointYML: endpoints,
	}

	r.tempDir = path.Join(os.TempDir(), "rtsp")
	err := os.MkdirAll(r.tempDir, os.ModeDir)
	if err != nil {
		return err
	}
	config.StoreDir = r.tempDir
	core.SetupLogger(config)
	fileServer := http.FileServer(http.Dir(r.tempDir))
	router := httprouter.New()
	controllers := core.NewController(config, fileServer)
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})
	if config.EndpointYML.Endpoints.List.Enabled {
		router.GET("/list", controllers.ListStreamHandler)
		//	logrus.Infoln("list endpoint enabled | MainProcess")
	}
	if config.EndpointYML.Endpoints.Start.Enabled {
		router.POST("/start", controllers.StartStreamHandler)
		//	logrus.Infoln("start endpoint enabled | MainProcess")
	}
	if config.EndpointYML.Endpoints.Static.Enabled {
		router.GET("/stream/*filepath", controllers.StaticFileHandler)
		//	logrus.Infoln("static endpoint enabled | MainProcess")
	}
	if config.EndpointYML.Endpoints.Stop.Enabled {
		router.POST("/stop", controllers.StopStreamHandler)
		//	logrus.Infoln("stop endpoint enabled | MainProcess")
	}

	done := controllers.ExitPreHook()
	handler := cors.AllowAll().Handler(router)
	if config.CORS.Enabled {
		handler = cors.New(cors.Options{
			AllowCredentials: config.CORS.AllowCredentials,
			AllowedOrigins:   config.CORS.AllowedOrigins,
			MaxAge:           config.CORS.MaxAge,
		}).Handler(router)
	}
	r.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: handler,
	}
	go func() {
		//	logrus.Infof("rtsp-stream transcoder started on %d | MainProcess", config.Port)
		err := r.httpServer.ListenAndServe()
		if r.shuttingDown == 0 && err != nil {
			r.logger.Fatal("Failed to start RTSP server", zap.Error(err))
		}
	}()

	go func() {
		<-done
		r.stop()
	}()
	r.logger.Info("RTSP server started")
	return nil
}

func (r *rtspServer) stop() error {
	if r.httpServer != nil {
		r.shuttingDown = 1
		server := r.httpServer
		r.httpServer = nil
		if err := server.Shutdown(context.Background()); err != nil {
			r.logger.Error("HTTP server Shutdown", zap.Error(err))
			return err
		}
	}
	return nil
}

func (r *rtspServer) Handle(cameraID string, path string, w http.ResponseWriter, req *http.Request) error {

	client := http.DefaultClient

	p := fmt.Sprintf("http://localhost:%d/%s", rtspServerPort, path)
	resp, err := client.Get(p)
	if err != nil {
		r.logger.Error("Error getting stream chunk", zap.Error(err))
		return err
	}

	if resp.StatusCode != http.StatusOK {
		r.logger.Warn("RTSP static handler failed",
			zap.Int("status-code", resp.StatusCode),
			zap.String("path", p),
		)
	}

	for k, v := range resp.Header {
		w.Header().Add(k, v[0])
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		return err
	}

	w.Write(bytes)
	return nil
}

var durationParse = regexp.MustCompile(`(?m)^#EXT-X-TARGETDURATION:(\d+(.\d+)?)$`)

func (r *rtspServer) getTargetDuration(playlist string) float64 {
	match := durationParse.FindAllStringSubmatch(playlist, 1)

	if match == nil {
		return 0
	}

	duration := match[0][1]
	val, err := strconv.ParseFloat(duration, 64)
	if err != nil {
		r.logger.Error("Can't parse media duration", zap.String("val", duration), zap.Error(err))
		return 0
	}
	return val
}

func (r *rtspServer) StreamPath(cameraID string) (string, error) {

	// get the RTSP path for the camera
	//
	cam, err := r.data.GetCamera(cameraID)

	if err != nil {
		return "", err
	}

	cameraType, err := r.cameraRegistry.Get(cam.Type)

	if err != nil {
		return "", err
	}

	caps := cameraType.Capabilities()

	if !caps.LiveStream {
		return "", fmt.Errorf("Camera type %q does not support live stream", cam.Type)
	}

	rtspURL, err := cameraType.RtspUri(cam)

	if err != nil {
		r.logger.Error("Couldn't get RTSP path", zap.String("camera-id", cameraID), zap.Error(err))
		return "", err
	}

	// start the stream
	client := http.DefaultClient

	body := struct {
		URI   string `json:"uri"`
		Alias string `json:"alias,omitempty"`
	}{
		URI: rtspURL,
		//Alias: cam.CameraID(), // causes a bug in the rtsp lib
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("Failed to encode json: %w", err)
	}

	response, err := client.Post(
		fmt.Sprintf("http://localhost:%d/start", rtspServerPort),
		"application/json",
		bytes.NewReader(raw),
	)

	if err != nil {
		return "", fmt.Errorf("Failed to call /start: %w", err)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("Failed to call /start %d: %w", response.StatusCode, err)
	}

	stream := struct {
		URI     string `json:"uri"`
		Running bool   `json:"running"`
		ID      string `json:"id"`
		Alias   string `json:"alias"`
	}{}

	raw = make([]byte, response.ContentLength)

	_, err = io.ReadFull(response.Body, raw)
	err = json.Unmarshal(raw, &stream)

	if err != nil {
		return "", fmt.Errorf("Failed to decode json: %w", err)
	}

	if !stream.Running {
		return "", fmt.Errorf("Started stream but it is not running id=%s", stream.ID)
	}

	// wait until we get a non zero duration result
	//
	target := fmt.Sprintf("http://localhost:%d/%s", rtspServerPort, stream.URI)
	for until := time.Now().Add(time.Second * 10); time.Now().Before(until); {
		response, err := client.Get(target)
		if err != nil {
			r.logger.Error("Error calling stream endpoint", zap.Error(err), zap.String("uri", target))
			break
		}

		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			r.logger.Error("Error reading body from stream endpoint", zap.Error(err), zap.String("uri", target))
			break
		}
		playlist := string(bytes)

		duration := r.getTargetDuration(playlist)
		if duration > 0 {
			break
		}
		sleep := time.Second
		r.logger.Info("RTSP media index is not ready, sleeping", zap.Duration("sleep", sleep))
		time.Sleep(sleep)
	}

	// pick out the URI & return it
	return stream.URI, nil
}
