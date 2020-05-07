package ingest

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type MediaFileType int

var ErrorUnknownFile = errors.New("UnknownFileType")

const badVideoPath = "BadVideoPath"

const (
	Unknown MediaFileType = 0
	MP4     MediaFileType = 1
	JPG     MediaFileType = 2
)

type MediaFile struct {
	Type      MediaFileType
	Path      string
	Timestamp time.Time
	Duration  *time.Duration
}

type Ingester interface {
	OnNewFile(path string) (*MediaFile, error)
}

func New(tz *time.Location) (Ingester, error) {
	return &amcrestIngester{
		tz: tz,
	}, nil
}

type amcrestIngester struct {
	tz *time.Location
}

func (ai *amcrestIngester) OnNewFile(path string) (*MediaFile, error) {

	if strings.Contains(path, "/dav/") {
		// "2019-05-09/001/dav/21/21.04.49-21.05.14[M][0@0][0].mp4"
		ts := pathToTimestamps(path, ai.tz)
		if len(ts) != 2 {
			return nil, fmt.Errorf("%s: %v", badVideoPath, path)
		}
		d := ts[1].Sub(ts[0])
		return &MediaFile{
			Type:      MP4,
			Timestamp: ts[0],
			Duration:  &d,
		}, nil
	}

	if strings.Contains(path, "/jpg/") {
		// 2019-05-09/001/jpg/06/07/27[M][0@0][0].jpg
		ts, err := jpgPathToTimestamp(path, ai.tz)
		if err != nil {
			return nil, err
		}
		return &MediaFile{
			Type:      JPG,
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