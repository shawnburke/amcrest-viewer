package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/models"
)

func New(args *common.Params, logger *zap.Logger, data data.Repository) HttpServer {

	server := &Server{
		FileRoot: args.DataDir,
		Logger:   logger,
		args:     args,
		data:     data,
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

	data data.Repository
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

// func (s *Server) index(w http.ResponseWriter, r *http.Request) {

// 	var mainPageTemplate = template.Must(template.ParseFiles("index.html"))

// 	files, err := scanner.FindFiles(s.FileRoot, s.Logger, true)
// 	if err != nil {
// 		io.WriteString(w, err.Error())
// 	}

// 	sort.Slice(files, func(i, j int) bool {
// 		return files[i].Date.Unix() > files[j].Date.Unix()
// 	})

// 	dates := make([]viewModel, 0, len(files))
// 	for _, f := range files {
// 		if len(f.Videos) > 0 {
// 			dates = append(dates, newViewModel(f))
// 		}
// 	}

// 	d := struct {
// 		Title string
// 		Dates []viewModel
// 		Json  string
// 	}{
// 		Title: "MP4 Files",
// 		Dates: dates,
// 	}

// 	if r.URL.Query().Get("debug") != "" {
// 		jb, _ := json.MarshalIndent(d, "", "  ")

// 		d.Json = string(jb)
// 	}

// 	err = mainPageTemplate.Execute(w, d)
// 	if err != nil {
// 		s.Logger.Error("Error rendering", zap.Error(err))
// 	}
// }

// func (s *Server) serve(w http.ResponseWriter, r *http.Request) {

// 	p := r.URL.Path[7:]

// 	fn := path.Join(s.FileRoot, p)

// 	s.Logger.Info("Request", zap.String("filename", fn),
// 		zap.String("ip", r.RemoteAddr),
// 		zap.String("user-agent", r.UserAgent()),
// 	)
// 	http.ServeFile(w, r, fn)
// }

// func (s *Server) health(w http.ResponseWriter, r *http.Request) {
// 	// make sure the path exists
// 	if _, err := os.Stat(s.FileRoot); err != nil {
// 		w.WriteHeader(500)
// 		w.Write([]byte("Bad file path"))
// 		return
// 	}

// 	w.Write([]byte("OK"))
// }

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
		Message: "Error accessing DB",
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

func (s *Server) listCameras(w http.ResponseWriter, r *http.Request) {

	cams, err := s.data.ListCameras()

	if s.writeError(err, w, 0) {

		return
	}

	s.writeJson(cams, w, 0)

}

func (s *Server) Setup(public string) http.Handler {
	s.r = mux.NewRouter()

	// assets
	s.r.PathPrefix("/public/").Handler(
		http.StripPrefix("/public/",
			http.FileServer(http.Dir(public))))

	// s.r.PathPrefix("/files").HandlerFunc(s.serve)

	// s.r.HandleFunc("/health", s.health)
	// s.r.HandleFunc("/", s.index)

	s.r.Methods("POST").Path("/cameras").HandlerFunc(s.createCamera)
	s.r.Methods("GET").Path("/cameras").HandlerFunc(s.listCameras)
	s.r.Methods("GET").Path("/cameras/{id}").HandlerFunc(s.getCamera)
	return s.r
}
