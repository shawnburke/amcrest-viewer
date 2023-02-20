module github.com/shawnburke/amcrest-viewer

go 1.19

replace github.com/shawnburke/amcrest-viewer/web => ./web

replace github.com/Roverr/rtsp-stream => github.com/shawnburke/rtsp-stream v2.2.3-patch+incompatible

require (
	github.com/Roverr/rtsp-stream v2.1.1+incompatible
	github.com/go-chi/chi/v5 v5.0.8
	github.com/goftp/file-driver v0.0.0-20180502053751-5d604a0fc0c9
	github.com/goftp/server v0.0.0-20190712054601-1149070ae46b
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/gorilla/mux v1.7.4
	github.com/icholy/digest v0.1.7
	github.com/jmoiron/sqlx v1.2.0
	github.com/julienschmidt/httprouter v1.2.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.6.1
	go.uber.org/config v1.4.0
	go.uber.org/fx v1.12.0
	go.uber.org/zap v1.10.0
)

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Roverr/hotstreak v1.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jlaffaye/ftp v0.0.0-20200708175026-55bbb372b87e // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/natefinch/lumberjack v2.0.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/riltech/streamer v1.0.2 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	go.uber.org/atomic v1.5.0 // indirect
	go.uber.org/dig v1.9.0 // indirect
	go.uber.org/multierr v1.4.0 // indirect
	go.uber.org/tools v0.0.0-20190618225709-2cfd321de3ee // indirect
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367 // indirect
	golang.org/x/sys v0.0.0-20200212091648-12a6c2dcc1e4 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20200213224642-88e652f7a869 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
	honnef.co/go/tools v0.0.1-2019.2.3 // indirect
)
