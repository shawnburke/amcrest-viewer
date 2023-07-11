package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	openapi_server "github.com/shawnburke/amcrest-viewer/.gen/server"
	"github.com/shawnburke/amcrest-viewer/cameras"
	"github.com/shawnburke/amcrest-viewer/storage"
	"github.com/shawnburke/amcrest-viewer/storage/data"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/file"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"go.uber.org/zap"
)

type handlers struct {
	s *Server

	data  data.Repository
	files file.Manager
	gc    storage.GCManager
	rtsp  cameras.RtspServer
}

var _ openapi_server.ServerInterface = &handlers{}

func (h *handlers) createCamera(w http.ResponseWriter, r *http.Request) {

	cam := &models.Camera{}

	bytes, err := io.ReadAll(r.Body)

	if h.s.writeError(err, w, 500) {
		return
	}

	err = json.Unmarshal(bytes, cam)

	if h.s.writeError(err, w, 400) {
		return
	}

	var host *string
	if cam.Host != "" {
		host = &cam.Host
	}
	camEntity, err := h.data.AddCamera(cam.Name, cam.Type, host)

	if h.s.writeError(err, w, 400) {
		return
	}

	h.s.writeJson(cam, w, 201, "Location", "cameras/"+camEntity.CameraID())

}

type getCameraResult struct {
	*entities.Camera
	LatestSnapshot *entities.File `json:"latest_snapshot,omitempty"`
}

func newCameraResult(cam *entities.Camera) *getCameraResult {
	cr := &getCameraResult{
		Camera: cam,
	}
	cr.CameraCreds.Password = nil
	return cr
}

const streamMarker = "/live/"

func (h *handlers) getCameraLiveStream(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(mux.Vars(r)["camera-id"])

	if err != nil {
		h.s.writeError(err, w, 400)
		return

	}
	redirect := r.URL.Query().Get("redirect") == "true"

	params := openapi_server.GetCameraLiveStreamParams{
		Redirect: &redirect,
	}

	h.GetCameraLiveStream(w, r, id, params)

}

func (h *handlers) handleCameraLiveStream(w http.ResponseWriter, r *http.Request) {

	// get the RTSP support for this camera

	strID := mux.Vars(r)["camera-id"]

	p := r.URL.Path

	markerIndex := strings.Index(p, streamMarker)

	if markerIndex == -1 {
		w.WriteHeader(400)
		return
	}

	p = p[markerIndex+len(streamMarker):]

	err := h.rtsp.Handle(strID, p, w, r)

	if h.s.writeError(err, w, 500) {
		return
	}
}

type updateCamera struct {
	models.Camera
	Password string `json:"Password,omitempty"`
	Username string `json:"Username,omitempty"`
}

func str(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h *handlers) updateCamera(w http.ResponseWriter, r *http.Request) {

	strID := mux.Vars(r)["camera-id"]

	cam := &updateCamera{}

	bytes, err := io.ReadAll(r.Body)

	if h.s.writeError(err, w, 500) {
		return
	}

	err = json.Unmarshal(bytes, cam)

	if h.s.writeError(err, w, 400) {
		return
	}

	existingCam, err := h.data.GetCamera(strID)

	if h.s.writeError(err, w, 0) {
		return
	}

	var host *string = existingCam.Host
	if cam.Host != "" {
		host = &cam.Host
	}

	newCam, err := h.data.UpdateCamera(strID, &existingCam.Name, host, existingCam.Enabled)
	if h.s.writeError(err, w, 400) {
		return
	}

	var pwd *string = newCam.Host
	if cam.Password != "" {
		pwd = &cam.Password
	}

	var user *string = newCam.Username
	if cam.Username != "" {
		user = &cam.Username
	}

	err = h.data.UpdateCameraCreds(strID, str(host), str(user), str(pwd))

	if h.s.writeError(err, w, 400) {
		return
	}

	c2 := &models.Camera{
		Host:    *newCam.Host,
		Enabled: *newCam.Enabled,
		ID:      newCam.CameraID(),
		Name:    newCam.Name,
		Type:    newCam.Type,
	}

	h.s.writeJson(c2, w, 200)

}

func (h *handlers) updateCameraCreds(w http.ResponseWriter, r *http.Request) {
	strID := mux.Vars(r)["camera-id"]

	creds := struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	bytes, err := io.ReadAll(r.Body)

	if h.s.writeError(err, w, 500) {
		return
	}

	err = json.Unmarshal(bytes, &creds)

	if h.s.writeError(err, w, 400) {
		return
	}

	err = h.data.UpdateCameraCreds(strID, creds.Host, creds.Username, creds.Password)
	if h.s.writeError(err, w, 400) {
		return
	}

	w.WriteHeader(200)

}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02",
}

func (h *handlers) parseTime(t string) (time.Time, error) {
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

	return time.Time{}, fmt.Errorf("could not parse time %q as either YYYY-MM-DD or YY-MM-DDTHH:MM:SSZ", t)
}

func (h *handlers) updateFilePaths(cam string, files ...*entities.File) []*entities.File {
	newFiles := make([]*entities.File, len(files))
	for i, f := range files {
		if f == nil {
			continue
		}
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

func (h *handlers) getFileInfo(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["file-id"]

	if ext := path.Ext(idStr); ext != "" {
		idStr = idStr[0 : len(idStr)-len(ext)]
	}

	id, err := strconv.Atoi(idStr)
	if h.s.writeError(err, w, 400) {
		return
	}

	fileInfo, err := h.data.GetFile(id)
	if h.s.writeError(err, w, 400) {
		return
	}

	camid := mux.Vars(r)["camera-id"]

	fi := h.updateFilePaths(camid, fileInfo)

	h.s.writeJson(fi[0], w, 200)
}

func (h *handlers) adminTriggerGC(w http.ResponseWriter, r *http.Request) {
	if h.gc != nil {
		err := h.gc.Cleanup()
		if h.s.writeError(err, w, 500) {
			return
		}
		w.WriteHeader(200)
		return
	}
	w.Write([]byte("GC not available"))
	w.WriteHeader(404)
}

var resizeCounter = int32(0)
var resizeLogThreshold = int32(25)

func (h *handlers) resizeImage(raw []byte, maxSize int) ([]byte, error) {

	start := time.Now()

	defer func() {
		delta := time.Since(start)

		if atomic.AddInt32(&resizeCounter, 1)%resizeLogThreshold == 0 {
			h.s.Logger.Info("resize", zap.Duration("resize", delta))
		}
	}()

	reader := bytes.NewReader(raw)

	data, _, err := image.Decode(reader)

	if err != nil {
		return raw, err
	}

	// already the right size.
	if data.Bounds().Size().X <= maxSize {
		return raw, nil
	}

	newImage := resize.Resize(uint(maxSize), 0, data, resize.NearestNeighbor)

	if newImage == nil {
		return raw, nil
	}

	// TODO: consider pooling buffers
	buffer := &bytes.Buffer{}
	buffer.Grow(len(raw))
	err = jpeg.Encode(buffer, newImage, nil)

	if err == nil {
		return buffer.Bytes(), nil
	}

	return nil, err
}

func (h *handlers) getFile(w http.ResponseWriter, r *http.Request) {

	idStr := mux.Vars(r)["file-id"]

	if ext := path.Ext(idStr); ext != "" {
		idStr = idStr[0 : len(idStr)-len(ext)]
	}

	id, err := strconv.Atoi(idStr)
	if h.s.writeError(err, w, 400) {
		return
	}

	fileInfo, err := h.data.GetFile(id)
	if h.s.writeError(err, w, 400) {
		return
	}

	stream := r.URL.Query().Get("stream") != ""
	maxWidthStr := r.URL.Query().Get("max_width")
	if stream || maxWidthStr != "" {
		reader, err := h.files.GetFile(fileInfo.Path)
		if h.s.writeError(err, w, 400) {
			return
		}

		defer reader.Close()

		contents, err := io.ReadAll(reader)

		if h.s.writeError(err, w, 500) {
			return
		}

		if fileInfo.Type == entities.FileTypeJpg && maxWidthStr != "" {

			if maxWidth, err := strconv.Atoi(maxWidthStr); err == nil {

				contents2, err := h.resizeImage(contents, maxWidth)

				if err != nil {
					h.s.Logger.Warn("Failed to resize image", zap.Error(err), zap.Int("file-id", fileInfo.ID))
				}
				if contents2 != nil {
					contents = contents2
				}
			}
		}

		contentType := getContentType(fileInfo)

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(contents)))

		if dl := r.URL.Query().Get("download"); dl != "" {
			w.Header().Set("Content-Disposition", "attachment; filename="+path.Base(fileInfo.Path))
		}

		w.WriteHeader(200)

		// write header bytes
		_, err = w.Write(contents)
		if err != nil {
			h.s.Logger.Error("Error writing file",
				zap.String("path", fileInfo.Path),
				zap.Int("file-id", fileInfo.ID),
				zap.Error(err))
		}
		return
	}

	p, err := h.files.GetFilePath(fileInfo.Path)
	if h.s.writeError(err, w, 400) {
		return
	}

	http.ServeFile(w, r, p)

}

// OpenAPI wrappers

func (h *handlers) GetCameras(w http.ResponseWriter, r *http.Request, params openapi_server.GetCamerasParams) {

	cams, err := h.data.ListCameras()

	if h.s.writeError(err, w, 0) {

		return
	}

	res := make([]*getCameraResult, len(cams))

	for i, cam := range cams {

		r1 := newCameraResult(cam)

		f, err := h.data.GetLatestFile(cam.CameraID(), 0)

		if err != nil {
			if h.s.writeError(err, w, 0) {
				return
			}
		}
		updated := h.updateFilePaths(cam.CameraID(), f)
		r1.LatestSnapshot = updated[0]

		res[i] = r1
	}

	h.s.writeJson(res, w, 0)

}

func (h *handlers) GetCamera(w http.ResponseWriter, r *http.Request, id int) {

	cam, err := h.data.GetCamera(strconv.FormatInt(int64(id), 10))

	if h.s.writeError(err, w, 0) {

		return
	}

	res := newCameraResult(cam)

	f, err := h.data.GetLatestFile(cam.CameraID(), 0)

	if err != nil {
		if h.s.writeError(err, w, 0) {
			return
		}
	}
	updated := h.updateFilePaths(cam.CameraID(), f)
	res.LatestSnapshot = updated[0]

	h.s.writeJson(res, w, 0)

}

func (h *handlers) GetCameraLiveStream(w http.ResponseWriter, r *http.Request, id int, params openapi_server.GetCameraLiveStreamParams) {
	// get the RTSP support for this camera
	strID := strconv.Itoa(id)
	rtspPath, err := h.rtsp.StreamPath(strID)

	if h.s.writeError(err, w, 500) {
		return
	}

	rtspPath = fmt.Sprintf("/api/cameras/%s%s%s", strID, streamMarker, rtspPath)

	// make sure we didn't get any double slashes
	rtspPath = strings.Replace(rtspPath, "//", "/", -1)

	w.Header().Add("Cache-Control", "no-store")

	if params.Redirect != nil && *params.Redirect {
		// redirect to the RTSP path
		//
		http.Redirect(w, r, rtspPath, http.StatusMovedPermanently)
		return
	}

	res := struct {
		URI string `json:"uri"`
	}{
		URI: rtspPath,
	}

	h.s.writeJson(res, w, 200)
}

func (h *handlers) GetCameraFiles(w http.ResponseWriter, r *http.Request, id string, params openapi_server.GetCameraFilesParams) {

	lff := &data.ListFilesFilter{
		Start: params.Start,
		End:   params.End,
	}

	lff.Descending = params.Sort != nil && *params.Sort == "desc"

	files, err := h.data.ListFiles(id, lff)

	if h.s.writeError(err, w, 0) {
		return
	}

	files = h.updateFilePaths(id, files...)

	h.s.writeJson(files, w, 0)

}

func (h *handlers) GetCameraStats(w http.ResponseWriter, r *http.Request, id int, params openapi_server.GetCameraStatsParams) {

	if params.End == nil {
		now := time.Now()
		params.End = &now
	}

	cam, err := h.data.GetCameraStats(strconv.Itoa(id), params.Start, params.End, "")

	if h.s.writeError(err, w, 0) {

		return
	}

	h.s.writeJson(cam, w, 200)
}
