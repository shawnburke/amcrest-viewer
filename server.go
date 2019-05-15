/*----------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See LICENSE in the project root for license information.
 *---------------------------------------------------------------------------------------*/

package main

import (
	"fmt"
	"io"
	"path"
	"net/http"
	"html/template"
	"github.com/gorilla/mux"

)



var fileRoot = "./data"

type viewModel struct {
	Date string
	DayOfWeek string
	Files  [] fileViewModel
}

func newViewModel(fd FileDate) viewModel{
	vm := viewModel{
		Date:  fd.Date.Format("2006-01-02"),
		DayOfWeek:  fd.Date.Weekday().String(),
	}

	for _, f := range fd.Files {
		vm.Files = append(vm.Files,fileViewModel{
			FileItem: f,
		})
	}
	return vm
}

type fileViewModel struct {
	FileItem
}

func (fvm fileViewModel) Start() string {
	return fmt.Sprintf("%d.%d", fvm.StartTime.Hour(), int( fvm.StartTime.Minute()/60.0*100))
}

func (fvm fileViewModel) End() string {
	return  fmt.Sprintf("%d.%d", fvm.EndTime.Hour(), int( fvm.EndTime.Minute()/60.0*100))
}

func (fvm fileViewModel) Description() string {
	return fmt.Sprintf("%s-%s", fvm.StartTime.Format("15:04:05"),fvm.EndTime.Format("15:04:05"))
}

func index(w http.ResponseWriter, r *http.Request) {

	var mainPageTemplate = template.Must(template.ParseFiles("index.html"))

	files, err := FindFiles(fileRoot)
	if err != nil {
		io.WriteString(w, err.Error())
	}
	dates := make([]viewModel, len(files))
	for i, f := range files {
		dates[i] =newViewModel(f)
	}
	d :=  struct {
		Title string
		Dates []viewModel
	} {
		Title: "MP4 Files",
		Dates: dates,
	}

	mainPageTemplate.Execute(w,d )
}

func serve(w http.ResponseWriter, r *http.Request) {

	p := r.URL.Path[7:]

	
	Filename := path.Join(fileRoot, p)

	fmt.Println("Client requests: " + Filename)
	http.ServeFile(w, r, Filename)
}

func main() {
	portNumber := "9000"

	r := mux.NewRouter()
	
	r.PathPrefix("/files").HandlerFunc(serve)
	r.PathPrefix("/public/").Handler(
		http.StripPrefix("/public/", 
		http.FileServer(http.Dir("./public/"))))
	r.HandleFunc("/", index)
	
    http.Handle("/", r)
	
	fmt.Println("Server listening on port ", portNumber)
	http.ListenAndServe(":"+portNumber, nil)
}
