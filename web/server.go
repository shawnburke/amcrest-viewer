package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"time"

	"html/template"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/scanner"
)

func New(args *common.Params, logger *zap.Logger) HttpServer {

	server := &Server{
		FileRoot: args.DataDir,
		Logger:   logger,
		args:     args,
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

func (s *Server) index(w http.ResponseWriter, r *http.Request) {

	var mainPageTemplate = template.Must(template.ParseFiles("index.html"))

	files, err := scanner.FindFiles(s.FileRoot, s.Logger, true)
	if err != nil {
		io.WriteString(w, err.Error())
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Date.Unix() > files[j].Date.Unix()
	})

	dates := make([]viewModel, 0, len(files))
	for _, f := range files {
		if len(f.Videos) > 0 {
			dates = append(dates, newViewModel(f))
		}
	}

	d := struct {
		Title string
		Dates []viewModel
		Json  string
	}{
		Title: "MP4 Files",
		Dates: dates,
	}

	if r.URL.Query().Get("debug") != "" {
		jb, _ := json.MarshalIndent(d, "", "  ")

		d.Json = string(jb)
	}

	err = mainPageTemplate.Execute(w, d)
	if err != nil {
		s.Logger.Error("Error rendering", zap.Error(err))
	}
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {

	p := r.URL.Path[7:]

	fn := path.Join(s.FileRoot, p)

	s.Logger.Info("Request", zap.String("filename", fn),
		zap.String("ip", r.RemoteAddr),
		zap.String("user-agent", r.UserAgent()),
	)
	http.ServeFile(w, r, fn)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	// make sure the path exists
	if _, err := os.Stat(s.FileRoot); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("Bad file path"))
		return
	}

	w.Write([]byte("OK"))
}

func (s *Server) Setup(public string) http.Handler {
	s.r = mux.NewRouter()

	s.r.PathPrefix("/files").HandlerFunc(s.serve)
	s.r.PathPrefix("/public/").Handler(
		http.StripPrefix("/public/",
			http.FileServer(http.Dir(public))))
	s.r.HandleFunc("/health", s.health)
	s.r.HandleFunc("/", s.index)
	return s.r
}