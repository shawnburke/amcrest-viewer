package ftp

import ( 
	
	"github.com/shawnburke/amcrest-viewer/common"

	ftps	"github.com/shawnburke/amcrest-viewer/ftp-server"
	"fmt"
)


type ftpFileSystem struct {
	server *ftps.Server
}

func New(args *common.Params) interface{} {
	fmt.Println("Created FTP server")
	return ftpFileSystem{}
} 