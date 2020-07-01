package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"github.com/shawnburke/amcrest-viewer/storage/models"
)

const defaultFrontendPath = "./web/frontend/build"

type HttpParams struct {
	fx.In
	Args   *common.Params
	Logger *zap.Logger
	Data   data.Repository
	Files  file.Manager
	Config config.Provider
}

func New(p HttpParams) HttpServer {

	server := &Server{
		FileRoot: p.Args.DataDir,
		Logger:   p.Logger,
		args:     p.Args,
		data:     p.Data,
		files:    p.Files,
	}

	frontendPath := ""

	err := p.Config.Get("web.frontend").Populate(&frontendPath)
	if err != nil {
		panic(err)
	}

	if frontendPath == "" {
		frontendPath = defaultFrontendPath
	}
	r := server.Setup(frontendPath)

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

		if errors.Is(err, os.ErrNotExist) {
			status = 404
		}
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

type getCameraResult struct {
	*entities.Camera
	LatestSnapshot *entities.File `json:"latest_snapshot,omitempty"`
}

func (s *Server) getCamera(w http.ResponseWriter, r *http.Request) {

	strID := mux.Vars(r)["camera-id"]

	cam, err := s.data.GetCamera(strID)

	if s.writeError(err, w, 0) {
		return
	}

	res := &getCameraResult{
		Camera: cam,
	}

	if ls := r.URL.Query().Get("latest_snapshot"); ls == "true" || ls == "1" {
		f, err := s.data.GetLatestFile(strID, 0)

		if err != nil {
			if s.writeError(err, w, 0) {
				return
			}
		}
		updated := s.updateFilePaths(strID, f)
		res.LatestSnapshot = updated[0]
	}
	s.writeJson(res, w, 200)
}

func (s *Server) getCameraStats(w http.ResponseWriter, r *http.Request) {

	strID := mux.Vars(r)["camera-id"]

	start, end, err := s.getTimeRange(r)

	if s.writeError(err, w, 400) {
		return
	}

	cam, err := s.data.GetCameraStats(strID, start, end, "")

	if s.writeError(err, w, 0) {

		return
	}

	s.writeJson(cam, w, 200)

}

func (s *Server) updateCamera(w http.ResponseWriter, r *http.Request) {

	strID := mux.Vars(r)["camera-id"]

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

	newCam, err := s.data.UpdateCamera(strID, name, nil, nil)
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

func (s *Server) updateCameraCreds(w http.ResponseWriter, r *http.Request) {
	strID := mux.Vars(r)["camera-id"]

	creds := struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	bytes, err := ioutil.ReadAll(r.Body)

	if s.writeError(err, w, 500) {
		return
	}

	err = json.Unmarshal(bytes, &creds)

	if s.writeError(err, w, 400) {
		return
	}

	err = s.data.UpdateCameraCreds(strID, creds.Host, creds.Username, creds.Password)
	if s.writeError(err, w, 400) {
		return
	}

	w.WriteHeader(200)

}

func (s *Server) listCameras(w http.ResponseWriter, r *http.Request) {

	cams, err := s.data.ListCameras()

	if s.writeError(err, w, 0) {

		return
	}

	res := make([]*getCameraResult, len(cams))

	for i, cam := range cams {

		r1 := &getCameraResult{Camera: cam}

		if ls := r.URL.Query().Get("latest_snapshot"); ls == "true" || ls == "1" {

			f, err := s.data.GetLatestFile(cam.CameraID(), 0)

			if err != nil {
				if s.writeError(err, w, 0) {
					return
				}
			}
			updated := s.updateFilePaths(cam.CameraID(), f)
			r1.LatestSnapshot = updated[0]
		}
		res[i] = r1
	}

	s.writeJson(res, w, 0)

}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02",
}

func (s *Server) parseTime(t string) (time.Time, error) {
	for _, tf := range timeFormats {
		t, err := time.Parse(tf, t)
		if err == nil {
			return t, nil
		}
	}

	// check unix time ( unix ms )
	val, err := strconv.ParseInt(t, 10, 64)
	if err == nil {
		return time.Unix(val/1000, 0), nil
	}

	return time.Time{}, fmt.Errorf("Could not parse time %q as either YYYY-MM-DD or YY-MM-DDTHH:MM:SSZ", t)
}

func (s *Server) updateFilePaths(cam string, files ...*entities.File) []*entities.File {
	newFiles := make([]*entities.File, len(files))
	for i, f := range files {
		nf := *f
		ext := ".jpg"
		if f.Type == 1 {
			ext = ".mp4"
		}
		nf.Path = fmt.Sprintf("/api/cameras/%s/files/%d%s", cam, f.ID, ext)
		newFiles[i] = &nf
	}
	return newFiles
}

func (s *Server) getTimeRange(r *http.Request) (*time.Time, *time.Time, error) {
	var start, end *time.Time

	if st := r.URL.Query().Get("start"); st != "" {
		t, err := s.parseTime(st)
		if err != nil {
			return nil, nil, fmt.Errorf("Bad start time format: %w", err)
		}
		start = &t
	}

	if et := r.URL.Query().Get("end"); et != "" {
		t, err := s.parseTime(et)
		if err != nil {
			return nil, nil, fmt.Errorf("Bad end time format: %w", err)
		}
		end = &t
	} else {
		t := time.Now()
		end = &t
	}

	return start, end, nil
}

func (s *Server) listFiles(w http.ResponseWriter, r *http.Request) {

	cameraID := mux.Vars(r)["camera-id"]

	if cameraID == "" {
		s.writeError(errors.New("Camera ID required"), w, 400)
		return
	}

	start, end, err := s.getTimeRange(r)

	if s.writeError(err, w, 400) {
		return
	}

	lff := &data.ListFilesFilter{
		Start: start,
		End:   end,
	}

	sort := r.URL.Query().Get("sort")
	lff.Descending = sort == "desc"

	files, err := s.data.ListFiles(cameraID, lff)

	if s.writeError(err, w, 0) {
		return
	}

	files = s.updateFilePaths(cameraID, files...)

	s.writeJson(files, w, 0)

}

func (s *Server) getFileInfo(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["file-id"]

	if ext := path.Ext(idStr); ext != "" {
		idStr = idStr[0 : len(idStr)-len(ext)]
	}

	id, err := strconv.Atoi(idStr)
	if s.writeError(err, w, 400) {
		return
	}

	fileInfo, err := s.data.GetFile(id)
	if s.writeError(err, w, 400) {
		return
	}

	camid := mux.Vars(r)["camera-id"]

	fi := s.updateFilePaths(camid, fileInfo)

	s.writeJson(fi[0], w, 200)
}

const mimeTextPlain = "text/plain"

func (s *Server) getFile(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["file-id"]

	if ext := path.Ext(idStr); ext != "" {
		idStr = idStr[0 : len(idStr)-len(ext)]
	}

	id, err := strconv.Atoi(idStr)
	if s.writeError(err, w, 400) {
		return
	}

	fileInfo, err := s.data.GetFile(id)
	if s.writeError(err, w, 400) {
		return
	}

	stream := r.URL.Query().Get("stream") != ""

	if stream {
		reader, err := s.files.GetFile(fileInfo.Path)
		if s.writeError(err, w, 400) {
			return
		}
		defer reader.Close()

		contentType := getContentType(fileInfo)

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Length))

		if dl := r.URL.Query().Get("download"); dl != "" {
			w.Header().Set("Content-Disposition", "attachment; filename="+path.Base(fileInfo.Path))
		}

		w.WriteHeader(200)

		// write header bytes
		_, err = io.Copy(w, reader)
		if err != nil {
			s.Logger.Error("Error writing file",
				zap.String("path", fileInfo.Path),
				zap.Int("file-id", fileInfo.ID),
				zap.Error(err))
		}
		return
	}

	p, err := s.files.GetFilePath(fileInfo.Path)
	if s.writeError(err, w, 400) {
		return
	}
	http.ServeFile(w, r, p)

}

func (s *Server) Setup(frontendPath string) http.Handler {
	s.r = mux.NewRouter()

	// cameras
	s.r.Methods("POST").Path("/api/cameras").HandlerFunc(s.createCamera)
	s.r.Methods("GET").Path("/api/cameras").HandlerFunc(s.listCameras)
	s.r.Methods("GET").Path("/api/cameras/{camera-id}").HandlerFunc(s.getCamera)
	s.r.Methods("GET").Path("/api/cameras/{camera-id}/stats").HandlerFunc(s.getCameraStats)

	s.r.Methods("PUT").Path("/api/cameras/{camera-id}").HandlerFunc(s.updateCamera)
	s.r.Methods("PUT").Path("/api/cameras/{camera-id}/creds").HandlerFunc(s.updateCameraCreds)
	// files
	s.r.Methods("GET").Path("/api/cameras/{camera-id}/files").HandlerFunc(s.listFiles)

	s.r.Methods("GET").Path("/api/cameras/{camera-id}/files/{file-id}").HandlerFunc(s.getFile)
	s.r.Methods("GET").Path("/api/cameras/{camera-id}/files/{file-id}/info").HandlerFunc(s.getFileInfo)

	s.Logger.Info("web server path", zap.String("path", frontendPath))
	// website
	s.r.Methods("GET").PathPrefix("/").Handler(
		http.FileServer(http.Dir(frontendPath)),
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
