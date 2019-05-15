package main

import (
	"time"
	"os"
	"path/filepath"
	"path"
	"regexp"
	"fmt"
	"sort"
)


type FileDate struct {
	Date  time.Time
	//DayOfWeek string
	Files    []FileItem
}

type FileItem struct {
	StartTime time.Time
	EndTime   time.Time
	// Start   string
	// End string
	Path      string
}

var dateRegEx = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
var tsRegEx = regexp.MustCompile(`(\d{2}\.\d{2}\.\d{2})`)
var timeFormat = "2006-01-02T15.04.05"
var timeLoc *time.Location

func init() {
	tl, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
	timeLoc = tl
}

func pathToTimestampes(p string) []time.Time {
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

func FindFiles(root string) ([]FileDate, error) {

	if !path.IsAbs(root) {
		cwd, _ := os.Getwd()
		root = path.Clean(path.Join(cwd, root))
	}

	items := []FileItem{}

	filepath.Walk(root, func(p string, info os.FileInfo, err error) error{
		ext := path.Ext(p)

		if ext != ".mp4" {
			return nil
		}

	
		ts := pathToTimestampes(p)
	
		p, err =  filepath.Rel(root, p)
		if err != nil {
			panic(err)
		}
		items = append(items, FileItem{
			StartTime: ts[0],
			EndTime: ts[1],
			// Start: fmt.Sprintf("%d.%d", ts[0].Hour(), int(ts[0].Minute()/60.0*100)),
			// End: fmt.Sprintf("%d.%d", ts[1].Hour(), int(ts[1].Minute()/60.0*100)),
			Path: p,
		})
		return nil
	})


	sort.Slice(items, func(i, j int) bool {
		return items[i].StartTime.Unix() < items[j].StartTime.Unix()
	})

	dates := []FileDate{}

	var curDate string
	for _, item := range items {

		date := item.StartTime.Format("2006-01-02")
		
		if curDate != date {
		//	y, m, d := item.StartTime.Date()
			dates = append(dates, FileDate{
				Date: item.StartTime,
				//DayOfWeek: item.StartTime.Weekday().String(),
			})
			curDate = date
		}

		d := &dates[len(dates)-1]
		d.Files = append(d.Files, item)	
	}

	return dates, nil
}