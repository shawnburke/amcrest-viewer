package amcrest


import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"io"
	"time"

	gcommon "github.com/shawnburke/amcrest-viewer/common"
	"github.com/shawnburke/amcrest-viewer/storage/entities"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"github.com/shawnburke/amcrest-viewer/cameras/common"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const amcrestIngesterType = "amcrest"

// ErrorUnknownFile is standard
var ErrorUnknownFile = errors.New("UnknownFileType")

const badVideoPath = "BadVideoPath"


type AmcrestResult struct {
	fx.Out
	Instance common.Type `group:"ingester"`
}

func New(logger *zap.Logger, localTime *time.Location) (AmcrestResult, error) {

	amcrest := &amcrestCameraType{
		logger: logger,
		local: localTime,
	}

	return AmcrestResult{
		Instance: amcrest,
	}, nil
}


type amcrestCameraType struct {
	logger *zap.Logger
	local *time.Location
}

func (ac *amcrestCameraType) Name() string {
	return  amcrestIngesterType
}

func (ac *amcrestCameraType) Capabilities() common.Capabilities {
	return common.Capabilities{
		Snapshot: true,
	}
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
	return nil, nil
}


func (ac *amcrestCameraType) pathToFile(path string, tz string) (*models.MediaFile, error) {

	var loc *time.Location = ac.local
	if tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil {
			return nil, fmt.Errorf("Couldn't load location %q: %w", tz, err)
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
