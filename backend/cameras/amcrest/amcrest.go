package amcrest

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/shawnburke/amcrest-viewer/cameras/common"
	gcommon "github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const amcrestIngesterType = "amcrest"
const configKey = "cameras_types.amcrest"

// ErrorUnknownFile is standard
var ErrorUnknownFile = errors.New("UnknownFileType")

const badVideoPath = "BadVideoPath"

type amcrestConfig struct {
	ReturnSnapshot bool `yaml:"return_snapshot"`
}

type AmcrestResult struct {
	fx.Out
	Instance common.Type `group:"cameras"`
}

func New(
	logger *zap.Logger,
	localTime *time.Location,
	cfg config.Provider,
) (AmcrestResult, error) {

	amcrest := &amcrestCameraType{
		logger:  logger,
		local:   localTime,
		cameras: map[int]*amcrestApi{},
	}

	if val := cfg.Get(configKey); val.HasValue() {
		err := val.Populate(&amcrest.cfg)
		if err != nil {
			logger.Error("Error loading amcrest config", zap.Error(err))
		}
	}

	return AmcrestResult{
		Instance: amcrest,
	}, nil
}

type amcrestCameraType struct {
	logger  *zap.Logger
	local   *time.Location
	cfg     amcrestConfig
	cliPath *string
	cameras map[int]*amcrestApi
}

func (ac *amcrestCameraType) Name() string {
	return amcrestIngesterType
}

func (ac *amcrestCameraType) Capabilities() common.Capabilities {

	return common.Capabilities{
		LiveStream: true,
		Snapshot:   true,
	}
}

func (ac *amcrestCameraType) getApi(cam *entities.Camera) *amcrestApi {
	api, ok := ac.cameras[cam.ID]
	if !ok {
		api = &amcrestApi{
			Camera: cam,
			logger: ac.logger,
		}
		ac.cameras[cam.ID] = api
	}
	return api
}

func (ac *amcrestCameraType) ParseFilePath(cam *entities.Camera, p string) (*models.MediaFile, error) {

	mf, err := ac.pathToFile(p, cam.Timezone)

	switch path.Ext(p) {
	case ".mp4", ".jpg":
		break
	case ".idx":
		return nil, gcommon.ErrIngestDelete
	case ".mp4_", ".backup_":
		return nil, gcommon.ErrIngestIgnore
	}

	if err != nil {
		ac.logger.Error("Amcrest ingest unknown file", zap.Error(err), zap.String("path", p))
		return nil, err
	}
	mf.CameraID = cam.CameraID()
	return mf, nil
}

func (ac *amcrestCameraType) Snapshot(cam *entities.Camera) (io.ReadCloser, error) {

	if cam.Host == nil || cam.Username == nil || cam.Password == nil {
		return nil, fmt.Errorf("snapshot requires camera host, user, password")
	}

	tempPath := os.TempDir()
	tempPath = path.Join(tempPath, cam.CameraID())
	err := os.MkdirAll(tempPath, os.ModeDir)
	if err != nil {
		return nil, fmt.Errorf("can't create snapshot temp dir %q: %w", tempPath, err)
	}
	start := time.Now()

	tempPath = path.Join(tempPath, fmt.Sprintf("snapshot-%d.jpg", start.Unix()))
	var finish time.Time

	// trigger via API.
	aa := newAmcrestApi(*cam.Host, *cam.Username, *cam.Password, ac.logger)

	resp, err := aa.Execute("GET", "snapshot.cgi?channel=0")

	if err != nil {
		ac.logger.Error("Error getting snapshot from API", zap.Error(err))
		return nil, err
	}

	finish = time.Now()

	bytes, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		ac.logger.Warn("Error reading from snapshot api", zap.Error(err2))
		err = err2
	} else {
		if err2 = resp.Body.Close(); err2 != nil {
			ac.logger.Warn("Error closing body", zap.Error(err))
		}
		err = ioutil.WriteFile(tempPath, bytes, os.ModePerm)
	}

	ac.logger.Debug("Took snapshot",
		zap.String("camera", cam.CameraID()),
		zap.Duration("time", finish.Sub(start)),
	)

	f, err := os.Open(tempPath)

	if err != nil {
		return nil, err
	}

	res := &deleteOnClose{
		File:     f,
		fullPath: tempPath,
	}

	if !ac.cfg.ReturnSnapshot {
		res.Close()
		return nil, nil
	}

	return res, nil
}

func (ac *amcrestCameraType) RtspUri(cam *entities.Camera) (string, error) {

	if cam.Host == nil || cam.Username == nil || cam.Password == nil {
		return "", fmt.Errorf("rTSP requires camera host, user, password")
	}

	uri := fmt.Sprintf(
		"rtsp://%s:%s@%s:554/cam/realmonitor?channel=1&subtype=1",
		*cam.Username,
		*cam.Password,
		*cam.Host,
	)

	return uri, nil
}

type deleteOnClose struct {
	*os.File
	fullPath string
	logger   *zap.Logger
}

func (doc *deleteOnClose) Close() error {
	err := doc.File.Close()

	if delErr := os.Remove(doc.fullPath); delErr != nil {
		doc.logger.Error("Failed to cleanup snapshot file", zap.Error(delErr), zap.String("path", doc.fullPath))
	}
	return err
}

func (ac *amcrestCameraType) pathToFile(path string, tz string) (*models.MediaFile, error) {

	var loc *time.Location = ac.local
	if tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil {
			return nil, fmt.Errorf("couldn't load location %q: %w", tz, err)
		}
		loc = l
	}

	if strings.HasSuffix(path, ".mp4") && strings.Contains(path, "/dav/") {
		// "2019-05-09/001/dav/21/21.04.49-21.05.14[M][0@0][0].mp4"
		ts := pathToTimestamps(path, loc)
		if len(ts) != 2 {
			return nil, fmt.Errorf("%s: %v", badVideoPath, path)
		}
		d := ts[1].Sub(ts[0])
		return &models.MediaFile{
			Name:      path,
			Type:      models.MP4,
			Timestamp: ts[0],
			Duration:  &d,
		}, nil
	}

	if strings.HasSuffix(path, ".jpg") && strings.Contains(path, "/jpg/") {
		// 2019-05-09/001/jpg/06/07/27[M][0@0][0].jpg
		ts, err := jpgPathToTimestamp(path, loc)
		if err != nil {
			return nil, err
		}
		return &models.MediaFile{
			Name:      path,
			Type:      models.JPG,
			Timestamp: ts,
		}, nil
	}

	return nil, ErrorUnknownFile
}

var dateRegEx = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
var tsRegEx = regexp.MustCompile(`(\d{2}\.\d{2}\.\d{2})`)
var timeFormat = "2006-01-02T15.04.05"

func pathToTimestamps(p string, tz *time.Location) []time.Time {
	dateMatch := dateRegEx.FindString(p)
	if dateMatch == "" {
		return nil
	}
	timestampMatches := tsRegEx.FindAllString(p, -1)

	if timestampMatches == nil {
		return nil
	}

	// pick out the date and the time stamps and reformat
	// into a string like timeFormat above, then
	// use built in parsing to do the rest

	start := fmt.Sprintf("%sT%s", dateMatch, timestampMatches[0])
	s, err := time.ParseInLocation(timeFormat, start, tz)
	if err != nil {
		return nil
	}
	end := fmt.Sprintf("%sT%s", dateMatch, timestampMatches[1])
	e, err := time.ParseInLocation(timeFormat, end, tz)
	if err != nil {
		return nil
	}

	return []time.Time{
		s,
		e,
	}
}

// JPG Format
// 2019-05-09/001/jpg/06/07/27[M][0@0][0].jpg
// 2020-05-28/001/jpg/09/32/52[M][0@0][0].jpg

var counterRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2}(/\d{3}/)`)

func jpgPathToTimestamp(p string, tz *time.Location) (time.Time, error) {

	// convert to a format that built in parsing can manage by
	// removing the index number.
	match := counterRegex.FindStringSubmatch(p)

	if match == nil {
		return time.Time{}, ErrorUnknownFile
	}

	start := strings.Index(p, match[0])
	p = strings.Replace(p, match[1], "/xxx/", 1)

	return time.ParseInLocation("2006-01-02/xxx/jpg/15/04/05", p[start:start+27], tz)
}
