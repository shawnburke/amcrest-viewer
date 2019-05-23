/*----------------------------------------------------------------------------------------
 * Copyright (c) Microsoft Corporation. All rights reserved.
 * Licensed under the MIT License. See LICENSE in the project root for license information.
 *---------------------------------------------------------------------------------------*/

 package main

 import (
	 "fmt"
	 "net/http"
	 "os"
	
	 "go.uber.org/zap"

	 "github.com/shawnburke/amcrest-viewer/web"
	
 )
 
 
 var logger, _ = zap.NewDevelopment()
 
 func main() {
 
	 if len(os.Args) < 2 {
		 fmt.Println("Error: files path required")
		 os.Exit(1)
	 }
 
	 fileRoot := os.Args[1]
 
	 portNumber := "9000"
 
	 if len(os.Args) > 2 {
		 portNumber = os.Args[2]
	 }
	 
 
	 server := &web.Server{
		 FileRoot: fileRoot,
		 Logger:   logger,
	 }
 
	 r := server.Setup("./public/")
 
	 http.Handle("/", r)
 
	 logger.Info("Server listening", zap.String("port", portNumber))
	 err := http.ListenAndServe(":"+portNumber, nil)
	 if err != nil {
		 fmt.Println(err)
		 os.Exit(1)
	 }
 }
 