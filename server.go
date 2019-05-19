/*----------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See LICENSE in the project root for license information.
 *---------------------------------------------------------------------------------------*/

package main

import (
	"fmt"
	"io"
	"path"
	"os"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"
	"encoding/json"
	"sort"

)

var fileRoot string

type viewModel struct {
	Date string
	DayOfWeek string
	Videos  [] fileViewModel
}

func newViewModel(fd *FileDate) viewModel{
	vm := viewModel{
		Date:  fd.Date.Format("2006-01-02"),
		DayOfWeek:  fd.Date.Weekday().String(),
	}

	for _, f := range fd.Videos {
		vm.Videos = append(vm.Videos,fileViewModel{
			CameraVideo: *f,
		})
	}
	return vm
}

type fileViewModel struct {
	CameraVideo
}


func (fvm fileViewModel) Start() string {
	return fmt.Sprintf("%d.%d", fvm.Time.Hour(), int( fvm.Time.Minute()/60.0*100))
}

func (fvm fileViewModel) End() string {
	return  fmt.Sprintf("%d.%d", fvm.Time.Hour(), int( fvm.CameraVideo.End().Minute()/60.0*100))
}

func (fvm fileViewModel) Description() string {
	return fmt.Sprintf("%s (%s)", fvm.Time.Format("15:04:05"),fvm.CameraVideo.Duration.String())
}

func index(w http.ResponseWriter, r *http.Request) {

	var mainPageTemplate = template.Must(template.ParseFiles("index.html"))

	files, err := FindFiles(fileRoot, true)
	if err != nil {
		io.WriteString(w, err.Error())
	}
	dates := make([]viewModel, 0, len(files))
	for _, f := range files {
		if len(f.Videos) > 0 {
			dates = append(dates, newViewModel(f))
		}
	}

	sort.Slice(files, func(i,j int)bool {
		return files[i].Date.Unix() > files[j].Date.Unix()
	})


	d :=  struct {
		Title string
		Dates []viewModel
		Json  string
	} {
		Title: "MP4 Files",
		Dates: dates,
	}

	jb, _ := json.MarshalIndent(d, "", "  ")

	d.Json = string(jb)

	mainPageTemplate.Execute(w,d )
}

func serve(w http.ResponseWriter, r *http.Request) {

	p := r.URL.Path[7:]

	
	Filename := path.Join(fileRoot, p)

	fmt.Println("Client requests: " + Filename)
	http.ServeFile(w, r, Filename)
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Error: files path required")
		os.Exit(1)
	}

	fileRoot = os.Args[1]



	portNumber := "9000"

	r := mux.NewRouter()
	
	r.PathPrefix("/files").HandlerFunc(serve)
	r.PathPrefix("/public/").Handler(
		http.StripPrefix("/public/", 
		http.FileServer(http.Dir("./public/"))))
	r.HandleFunc("/", index)
	
    http.Handle("/", r)
	
	fmt.Println("Server listening on port ", portNumber)
	err := http.ListenAndServe(":"+portNumber, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
