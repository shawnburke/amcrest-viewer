package ingest

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/shawnburke/amcrest-viewer/ftp"
	"github.com/shawnburke/amcrest-viewer/storage/models"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const amcrestIngesterType = "amcrest"

// ErrorUnknownFile is standard
var ErrorUnknownFile = errors.New("UnknownFileType")

const badVideoPath = "BadVideoPath"

// todo: move this into its own package and load with ingest manager
// but solve Ingester circular dep
type AmcrestParams struct {
	fx.In
	TZ     *time.Location `optional:"true"`
	Logger *zap.Logger
}

type AmcrestResult struct {
	fx.Out
	Ingester Ingester `group:"ingester"`
}

func Amcrest(p AmcrestParams) (AmcrestResult, error) {

	amcrest := &amcrestIngester{
		tz:     p.Timezone(),
		logger: p.Logger,
	}

	return AmcrestResult{
		Ingester: amcrest,
	}, nil
}

func (p AmcrestParams) Timezone() *time.Location {
	if p.TZ != nil {
		return p.TZ
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	return loc
}

type amcrestIngester struct {
	tz     *time.Location
	logger *zap.Logger
}

func (ai *amcrestIngester) Name() string {
	return amcrestIngesterType
}
func (ai *amcrestIngester) IngestFile(f *ftp.File) *models.MediaFile {
	mf, err := ai.pathToFile(f.FullName)
	if err != nil {
		return nil
	}
	mf.CameraID = f.User
	return mf
}

func (ai *amcrestIngester) pathToFile(path string) (*models.MediaFile, error) {

	if strings.HasSuffix(path, ".mp4") && strings.Contains(path, "/dav/") {
		// "2019-05-09/001/dav/21/21.04.49-21.05.14[M][0@0][0].mp4"
		ts := pathToTimestamps(path, ai.tz)
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
		ts, err := jpgPathToTimestamp(path, ai.tz)
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

	start := fmt.Sprintf("%sT%s", dateMatch, timestampMatches[0])
	s, err := time.ParseInLocation(timeFormat, start, tz)
	if err != nil {
		panic(err)
	}
	end := fmt.Sprintf("%sT%s", dateMatch, timestampMatches[1])
	e, err := time.ParseInLocation(timeFormat, end, tz)
	if err != nil {
		panic(err)
	}

	return []time.Time{
		s,
		e,
	}
}

func jpgPathToTimestamp(p string, tz *time.Location) (time.Time, error) {

	// todo replace with regex
	p = strings.Replace(p, "/001/", "/xxx/", -1)

	return time.ParseInLocation("2006-01-02/xxx/jpg/15/04/05", p[0:27], tz)
}
