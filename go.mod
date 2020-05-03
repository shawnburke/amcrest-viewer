module github.com/shawnburke/amcrest-viewer

go 1.12

replace github.com/shawnburke/amcrest-viewer/web => ./web

require (
	github.com/goftp/file-driver v0.0.0-20180502053751-5d604a0fc0c9
	github.com/goftp/server v0.0.0-20190712054601-1149070ae46b // indirect
	github.com/gorilla/mux v1.7.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.8.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	go.uber.org/fx v1.12.0
	go.uber.org/zap v1.10.0
)
