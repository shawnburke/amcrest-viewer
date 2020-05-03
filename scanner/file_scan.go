package scanner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/shawnburke/amcrest-viewer/models"
)

const timeZone = "America/Los_Angeles"

var dateRegEx = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
var tsRegEx = regexp.MustCompile(`(\d{2}\.\d{2}\.\d{2})`)
var timeFormat = "2006-01-02T15.04.05"
var timeLoc *time.Location

func init() {
	tl, err := time.LoadLocation(timeZone)
	if err != nil {
		panic(err)
	}
	timeLoc = tl
}

func pathToTimestamps(p string) []time.Time {
	dateMatch := dateRegEx.FindString(p)
	if dateMatch == "" {
		return nil
	}
	timestampMatches := tsRegEx.FindAllString(p, -1)

	if timestampMatches == nil {
		return nil
	}

	start := fmt.Sprintf("%sT%s", dateMatch, timestampMatches[0])
	s, err := time.ParseInLocation(timeFormat, start, timeLoc)
	if err != nil {
		panic(err)
	}
	end := fmt.Sprintf("%sT%s", dateMatch, timestampMatches[1])
	e, err := time.ParseInLocation(timeFormat, end, timeLoc)
	if err != nil {
		panic(err)
	}

	return []time.Time{
		s,
		e,
	}
}

func jpgPathToTimestamp(p string) time.Time {

	p = strings.Replace(p, "/001/", "/xxx/", -1)

	ts, err := time.ParseInLocation("2006-01-02/xxx/jpg/15/04/05", p[0:27], timeLoc)
	if err != nil {
		panic(err)
	}
	return ts

}

func cachePath(root, date string) string {
	file := path.Join(root, fmt.Sprintf("%s-cache.json", date))
	return file
}

func FindFiles(root string, logger *zap.Logger, cache bool) ([]*models.FileDate, error) {

	if !path.IsAbs(root) {
		cwd, _ := os.Getwd()
		root = path.Clean(path.Join(cwd, root))
	}

	if _, err := os.Stat(root); err != nil {
		logger.Error("Bad root", zap.String("root", root), zap.Error(err))
		return nil, err
	}

	dates := []*models.FileDate{}

	items := []models.CameraItem{}

	filepath.Walk(root, func(p string, info os.FileInfo, err error) error {

		if info.IsDir() {
			logger.Info("Processing", zap.String("dir", p))
			dirname := path.Base((p))
			if cache && dateRegEx.MatchString(dirname) {
				file := cachePath(root, dirname)
				if _, err := os.Stat(file); err == nil {
					logger.Info("Loading cache file", zap.String("path", file))
					b, err := ioutil.ReadFile(file)
					if err != nil {
						panic(err)
					}
					loaded := &models.FileDate{}

					err = json.Unmarshal(b, loaded)

					dates = append(dates, loaded)
					return filepath.SkipDir
				}

			}
		}

		ext := path.Ext(p)

		rel, err := filepath.Rel(root, p)
		if err != nil {
			panic(err)
		}

		var item models.CameraItem

		switch ext {
		case ".mp4":

			ts := pathToTimestamps(p)

			item = &models.CameraVideo{
				CameraFile: models.CameraFile{
					Time: ts[0],
					Path: rel,
				},
				Duration: ts[1].Sub(ts[0]),
			}

			logger.Info("Found MP4", zap.Time("start-time", item.Timestamp()))

		case ".jpg":

			ts := jpgPathToTimestamp(rel)
			logger.Info("Found JPG", zap.Time("time", ts))

			item = &models.CameraStill{
				CameraFile: models.CameraFile{
					Time: ts,
					Path: rel,
				},
			}
		default:
			return nil
		}

		items = append(items, item)

		return nil
	})

	// sort
	sort.Slice(items, func(i, j int) bool {
		return items[i].Timestamp().Unix() < items[j].Timestamp().Unix()
	})

	// bucket into days, then into videos

	var curVideo *models.CameraVideo
	var curDate *models.FileDate

	for _, item := range items {

		date := item.Timestamp().Format("2006-01-02")

		if curDate == nil || curDate.DateString() != date {

			if curDate != nil && cache && time.Now().Sub(curDate.Date).Hours() > 24 {

				b, err := json.MarshalIndent(curDate, "  ", "")
				if err != nil {
					panic(err)
				}

				file := cachePath(root, curDate.DateString())
				err = ioutil.WriteFile(file, b, 666)
				if err != nil {
					panic(err)
				}
				logger.Info("Wrote cache", zap.Int("len", len(b)), zap.String("date", curDate.DateString()), zap.String("path", file))
			}
			curDate = &models.FileDate{
				Date: item.Timestamp(),
			}
			dates = append(dates, curDate)
			logger.Debug("Starting date bucket", zap.String("date", date))
		}

		if curVideo != nil && item.Timestamp().After(curVideo.End()) {
			logger.Debug("Ending video bucket",
				zap.Time("timestamp", item.Timestamp()),
				zap.String("path", curVideo.FilePath()),
			)
			curVideo = nil

		}

		switch i := item.(type) {
		case *models.CameraStill:
			if curVideo == nil {
				break
			}
			curVideo.Images = append(curVideo.Images, i)
			logger.Debug("Adding image",
				zap.Time("image-time", i.Timestamp()),
				zap.Time("video-bucket", curVideo.Timestamp()),
			)
		case *models.CameraVideo:
			curVideo = i
			curDate.Videos = append(curDate.Videos, i)
			logger.Debug("Adding video",
				zap.Time("timestamp", i.Timestamp()),
				zap.String("date-bucket", curDate.DateString()),
			)
		}
	}

	return dates, nil
}
