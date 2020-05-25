package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"github.com/shawnburke/amcrest-viewer/storage/models"
)

func New(args *common.Params, logger *zap.Logger,
	data data.Repository, files file.Manager) HttpServer {

	server := &Server{
		FileRoot: args.DataDir,
		Logger:   logger,
		args:     args,
		data:     data,
		files:    files,
	}

	r := server.Setup("./public/")

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

	data  data.Repository
	files file.Manager
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
	}
	s.writeJson(info, w, status)
	return true
}

func (s *Server) createCamera(w http.ResponseWriter, r *http.Request) {

	cam := &models.Camera{}

	bytes, err := ioutil.ReadAll(r.Body)

	if s.writeError(err, w, 500) {
		return
	}

	err = json.Unmarshal(bytes, cam)

	if s.writeError(err, w, 400) {
		return
	}

	var host *string
	if cam.Host != "" {
		*host = cam.Host
	}
	camEntity, err := s.data.AddCamera(cam.Name, cam.Type, host)

	if s.writeError(err, w, 400) {
		return
	}

	s.writeJson(cam, w, 201, "Location", "cameras/"+camEntity.CameraID())

}

func (s *Server) getCamera(w http.ResponseWriter, r *http.Request) {

	strID := mux.Vars(r)["id"]

	cam, err := s.data.GetCamera(strID)

	if s.writeError(err, w, 0) {

		return
	}

	s.writeJson(cam, w, 200)

}

func (s *Server) updateCamera(w http.ResponseWriter, r *http.Request) {

	cam := &models.Camera{}

	bytes, err := ioutil.ReadAll(r.Body)

	if s.writeError(err, w, 500) {
		return
	}

	err = json.Unmarshal(bytes, cam)

	if s.writeError(err, w, 400) {
		return
	}

	var name *string
	if cam.Name != "" {
		name = &cam.Name
	}

	newCam, err := s.data.UpdateCamera(cam.ID, name, nil, nil)
	if s.writeError(err, w, 400) {
		return
	}

	cam.ID = newCam.CameraID()
	cam.Name = newCam.Name

	if newCam.Host != nil {
		cam.Host = *newCam.Host
	}

	s.writeJson(cam, w, 200)

}

func (s *Server) listCameras(w http.ResponseWriter, r *http.Request) {

	cams, err := s.data.ListCameras()

	if s.writeError(err, w, 0) {

		return
	}

	s.writeJson(cams, w, 0)

}

func (s *Server) listFiles(w http.ResponseWriter, r *http.Request) {

	cameraID := mux.Vars(r)["camera-id"]

	if cameraID == "" {
		s.writeError(errors.New("Camera ID required"), w, 400)
		return
	}

	var start, end *time.Time

	if st := r.URL.Query().Get("start"); st != "" {
		t, err := time.Parse(time.RFC3339, st)
		if err != nil && s.writeError(fmt.Errorf("Bad start time format: %w", err), w, 400) {
			return
		}
		start = &t
	}

	if et := r.URL.Query().Get("end"); et != "" {
		t, err := time.Parse(time.RFC3339, et)
		if err != nil && s.writeError(fmt.Errorf("Bad start time format: %w", err), w, 400) {
			return
		}
		end = &t
	} else {
		t := time.Now()
		end = &t
	}

	cams, err := s.data.ListFiles(cameraID, start, end, nil)

	if s.writeError(err, w, 0) {
		return
	}

	s.writeJson(cams, w, 0)

}

func (s *Server) getFileInfo(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["file-id"]

	id, err := strconv.Atoi(idStr)
	if s.writeError(err, w, 400) {
		return
	}

	fileInfo, err := s.data.GetFile(id)
	if s.writeError(err, w, 400) {
		return
	}

	s.writeJson(fileInfo, w, 200)
}

func (s *Server) getFile(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["file-id"]

	id, err := strconv.Atoi(idStr)
	if s.writeError(err, w, 400) {
		return
	}

	fileInfo, err := s.data.GetFile(id)
	if s.writeError(err, w, 400) {
		return
	}

	reader, err := s.files.GetFile(fileInfo.Path)
	if s.writeError(err, w, 400) {
		return
	}
	defer reader.Close()

	header := make([]byte, 0, 512)
	n, err := reader.Read(header)
	if s.writeError(err, w, 400) {
		return
	}

	contentType := http.DetectContentType(header)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Length))
	w.Header().Set("Content-Disposition", "attachment; filename="+path.Base(fileInfo.Path))

	w.WriteHeader(200)

	// write header bytes
	w.Write(header[0:n])
	_, err = io.Copy(w, reader)
	if err != nil {
		s.Logger.Error("Error writing file",
			zap.String("path", fileInfo.Path),
			zap.Int("file-id", fileInfo.ID),
			zap.Error(err))
	}

}

func (s *Server) Setup(public string) http.Handler {
	s.r = mux.NewRouter()

	// cameras
	s.r.Methods("POST").Path("/api/cameras").HandlerFunc(s.createCamera)
	s.r.Methods("GET").Path("/api/cameras").HandlerFunc(s.listCameras)
	s.r.Methods("GET").Path("/api/cameras/{id}").HandlerFunc(s.getCamera)
	s.r.Methods("PUT").Path("/api/cameras/{id}").HandlerFunc(s.updateCamera)

	// files
	s.r.Methods("GET").Path("/api/files/{camera-id}").HandlerFunc(s.listFiles)

	s.r.Methods("GET").Path("/api/files/{camera-id}/{file-id}").HandlerFunc(s.getFile)
	s.r.Methods("GET").Path("/api/files/{camera-id}/{file-id}/info").HandlerFunc(s.getFileInfo)

	// website
	s.r.Methods("GET").PathPrefix("/").Handler(
		http.FileServer(http.Dir("./web/frontend/build")),
	)

	return s.r
}
