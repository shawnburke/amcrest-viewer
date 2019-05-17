package main

import (
	"time"
	"os"
	"path/filepath"
	"path"
	"regexp"
	"fmt"
	"strings"
	"sort"

	"go.uber.org/zap"
)

const timeZone = "America/Los_Angeles"

var logger, _ = zap.NewDevelopment();

type FileDate struct {
	Date  time.Time
	Files    []*FileItem
}

type FileItem struct {
	StartTime time.Time
	EndTime   time.Time
	Path      string
	Images    []ImageItem
	Thumb     ImageItem
}

type ImageItem struct {
	Time time.Time
	Path string
}

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

	return []time.Time {
		s,
		e,
	}
}

func jpgPathToTimestamp(p string) time.Time {

	p = strings.Replace(p, "/001/", "/xxx/", -1)

	ts, err := time.Parse("2006-01-02/xxx/jpg/15/04/05", p[0:27])
	if err != nil {
		panic(err)
	}
	return ts;
	
}

func FindFiles(root string) ([]FileDate, error) {

	if !path.IsAbs(root) {
		cwd, _ := os.Getwd()
		root = path.Clean(path.Join(cwd, root))
	}

	if _, err := os.Stat(root); err != nil {
		logger.Error("Bad root", zap.String("root", root), zap.Error(err))
		return nil, err
	}

	items := []*FileItem{}
	jpgs := []ImageItem{}

	filepath.Walk(root, func(p string, info os.FileInfo, err error) error{

		if info.IsDir() {
			logger.Info("Processing", zap.String("dir", p))
		}

		ext := path.Ext(p)

		rel, err :=  filepath.Rel(root, p)
		if err != nil {
			panic(err)
		}

		switch ext {
		case ".mp4":

			
		ts := pathToTimestamps(p)
	
	
		items = append(items, &FileItem{
			StartTime: ts[0],
			EndTime: ts[1],
			Path: rel,
		})

		logger.Info("Found MP4", zap.String("start-time", ts[0].String()))
	

		case ".jpg":

			ts := jpgPathToTimestamp(rel)
			jpgs = append(jpgs, ImageItem{
				Time: ts,
				Path: rel,
			})
		}
	

		return nil
	})


	sort.Slice(items, func(i, j int) bool {
		return items[i].StartTime.Unix() < items[j].StartTime.Unix()
	})

	sort.Slice(jpgs, func(i, j int) bool {
		return jpgs[i].Time.Unix() < jpgs[i].Time.Unix()
	})

	dates := []FileDate{}

	var curDate string
	var imgPos = 0
	for i, item := range items {

		date := item.StartTime.Format("2006-01-02")
		
		if curDate != date {
			dates = append(dates, FileDate{
				Date: item.StartTime,
			})
			curDate = date
		}

		d := &dates[len(dates)-1]

		for ; imgPos < len(jpgs); imgPos++ {
			img := jpgs[imgPos]
			if item.StartTime.Before(img.Time) &&
			   item.EndTime.After(img.Time) {
				   item.Images = append(item.Images, img)
				   continue
			   }

			if item.EndTime.Before(img.Time) {
				imgPos--
				break
			}
		}

		if len(item.Images) > 0 {
			items[i].Thumb = item.Images[len(item.Images)/2]
		}
		d.Files = append(d.Files, item)	
	}

	return dates, nil
}