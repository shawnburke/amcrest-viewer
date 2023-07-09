package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"

	openapi_server "github.com/shawnburke/amcrest-viewer/.gen/server"
	"github.com/shawnburke/amcrest-viewer/cameras"
	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/file"
)

type HttpParams struct {
	fx.In
	Args   *common.Params
	Logger *zap.Logger
	Data   data.Repository
	Files  file.Manager
	Config config.Provider
	GC     storage.GCManager
	Rtsp   cameras.RtspServer
}

func New(p HttpParams) HttpServer {

	server := &Server{
		FileRoot: p.Args.DataDir,
		Logger:   p.Logger,
		args:     p.Args,
		handlers: &handlers{
			data:  p.Data,
			files: p.Files,
			gc:    p.GC,
			rtsp:  p.Rtsp,
		},
	}
	server.handlers.s = server

	r := server.Setup(
		p.Config.Get("web.frontend.flutter").String(),
		p.Config.Get("web.frontend.js").String())

	http.Handle("/", r)

	return server
}

type HttpServer interface {
	Start() error
	Stop() error
}

type Server struct {
	FileRoot string
	Logger   *zap.Logger
	r        *mux.Router
	args     *common.Params
	server   *http.Server

	handlers *handlers
}

func (s *Server) Start() error {
	var err error

	s.server = &http.Server{
		Addr: fmt.Sprintf("%s:%d", s.args.Host, s.args.WebPort),
	}

	go func() {
		s.Logger.Info("Server listening", zap.Int("port", s.args.WebPort))
		err = s.server.ListenAndServe()

	}()

	time.Sleep(time.Millisecond * 100)
	if err != nil {
		fmt.Println(err)
	}

	return err

}

func (s *Server) Stop() error {
	if s.server != nil {
		svr := s.server
		s.server = nil
		return svr.Close()
	}
	return nil
}

func (s *Server) writeJson(obj interface{}, w http.ResponseWriter, status int, headers ...string) {

	j, err := json.MarshalIndent(obj, "", "  ")

	if err != nil {
		s.Logger.Error("Error writing JSON", zap.Error(err))
		j = []byte(fmt.Sprintf("Error writing json for type %T: %+v", obj, obj))
	}

	if status == 0 {
		status = 200
	}

	w.Header().Add("Content-Type", "application/json")
	for i := 0; i < len(headers); i += 2 {
		w.Header().Add(headers[i], headers[i+1])
	}
	w.WriteHeader(status)
	w.Write(j)

}

func (s *Server) writeError(err error, w http.ResponseWriter, status int) bool {
	if err == nil {
		return false
	}

	info := struct {
		Message string
		Error   string
	}{
		Message: "Error",
		Error:   err.Error(),
	}

	if status == 0 {

		status = 500

		if errors.Is(err, os.ErrNotExist) {
			status = 404
		}
	}
	s.writeJson(info, w, status)
	return true
}

func (s *Server) enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Setup(frontendFlutter, frontendJS string) http.Handler {

	s.r = mux.NewRouter()

	s.r.Use(s.enableCors)
	s.r.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rwp := &responseWriterProxy{ResponseWriter: w}
			h.ServeHTTP(rwp, r)
			s.Logger.Info(
				"Request",
				zap.String("path", r.URL.Path),
				zap.Any("query", r.URL.Query()),
				zap.Int("status", rwp.statusCode),
				zap.Int("content-length", rwp.contentLength),
			)
		})
	})

	// cameras
	s.r.Methods("POST").Path("/api/cameras").HandlerFunc(s.handlers.createCamera)
	s.r.Methods("GET").Path("/api/cameras/{camera-id}/live").HandlerFunc(s.handlers.getCameraLiveStream)

	// RTSP
	s.r.Methods("GET").PathPrefix("/api/cameras/{camera-id}/live/stream").HandlerFunc(s.handlers.handleCameraLiveStream)

	s.r.Methods("GET").Path("/api/cameras/{camera-id}/stats").HandlerFunc(s.handlers.getCameraStats)

	s.r.Methods("PUT").Path("/api/cameras/{camera-id}").HandlerFunc(s.handlers.updateCamera)
	s.r.Methods("PUT").Path("/api/cameras/{camera-id}/creds").HandlerFunc(s.handlers.updateCameraCreds)
	// files

	s.r.Methods("GET").Path("/api/cameras/{camera-id}/files/{file-id}").HandlerFunc(s.handlers.getFile)
	s.r.Methods("GET").Path("/api/cameras/{camera-id}/files/{file-id}/info").HandlerFunc(s.handlers.getFileInfo)

	s.r.Methods("POST").Path("/api/admin/gc").HandlerFunc(s.handlers.adminTriggerGC)

	s.r.PathPrefix("/api/").Handler(openapi_server.Handler(s.handlers))

	s.Logger.Info("Web server path", zap.String("flutter-path", frontendFlutter), zap.String("js-path", frontendJS))

	// website
	s.r.PathPrefix("/").Handler(
		http.FileServer(http.Dir(frontendFlutter)),
	)

	return s.r
}

func init() {
	// docker containers don't always have the mime.types file
	if mime.TypeByExtension(".mp4") == "" {
		types := map[string]string{
			".mp4": "video/mp4",
		}

		for k, v := range types {
			mime.AddExtensionType(k, v)
		}
	}
}

func getContentType(fi *entities.File) string {

	ct := mime.TypeByExtension(path.Ext(fi.Path))

	return ct
}

type responseWriterProxy struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

func (w *responseWriterProxy) Write(b []byte) (int, error) {
	w.contentLength += len(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriterProxy) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
