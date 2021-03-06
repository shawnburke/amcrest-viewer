module github.com/shawnburke/amcrest-viewer

go 1.12

replace github.com/shawnburke/amcrest-viewer/web => ./web

replace github.com/Roverr/rtsp-stream => github.com/shawnburke/rtsp-stream v2.2.3-patch+incompatible

require (
	github.com/Roverr/hotstreak v1.1.0 // indirect
	github.com/Roverr/rtsp-stream v2.1.1+incompatible
	github.com/goftp/file-driver v0.0.0-20180502053751-5d604a0fc0c9
	github.com/goftp/server v0.0.0-20190712054601-1149070ae46b
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/icholy/digest v0.1.7
	github.com/jlaffaye/ftp v0.0.0-20200708175026-55bbb372b87e // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pkg/errors v0.9.1
	github.com/riltech/streamer v1.0.2 // indirect
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/config v1.4.0
	go.uber.org/fx v1.12.0
	go.uber.org/zap v1.10.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)
