/*----------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See LICENSE in the project root for license information.
 *---------------------------------------------------------------------------------------*/

package web

import (
	"fmt"

	"github.com/shawnburke/amcrest-viewer/models"
)

type viewModel struct {
	Date      string
	DayOfWeek string
	Videos    []fileViewModel
}

func newViewModel(fd *models.FileDate) viewModel {
	vm := viewModel{
		Date:      fd.Date.Format("2006-01-02"),
		DayOfWeek: fd.Date.Weekday().String(),
	}

	for _, f := range fd.Videos {

		vm.Videos = append(vm.Videos, fileViewModel{

			CameraVideo: *f,
		})
	}
	return vm
}

type fileViewModel struct {
	models.CameraVideo
}

func (fvm fileViewModel) Start() string {
	return fmt.Sprintf("%d.%d", fvm.Time.Hour(), int(fvm.Time.Minute()/60.0*100))
}

func (fvm fileViewModel) End() string {
	return fmt.Sprintf("%d.%d", fvm.Time.Hour(), int(fvm.CameraVideo.End().Minute()/60.0*100))
}

func (fvm fileViewModel) Description() string {
	return fmt.Sprintf("%s (%s)", fvm.Time.Format("03:04:05 PM"), fvm.CameraVideo.Duration.String())
}

func (fvm fileViewModel) Thumbs() []*models.CameraStill {
	return fvm.thumbs(3)
}

func (fvm fileViewModel) thumbs(count int) []*models.CameraStill {
	imgCount := len(fvm.Images)

	if imgCount <= count {
		ret := make([]*models.CameraStill, imgCount)
		copy(ret, fvm.Images)
		return ret
	}

	skip := imgCount / count

	thumbs := make([]*models.CameraStill, count)
	pos := 0
	for index := 0; index < count; index++ {
		if pos >= imgCount {
			thumbs[index] = fvm.Images[imgCount-1]
			break
		}

		thumbs[index] = fvm.Images[pos]
		pos += skip
	}
	return thumbs
}
