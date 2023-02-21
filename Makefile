
ROOT=$(pwd)
SERVER=backend/amcrest-server
WEB_ROOT=frontend
FRONTEND=$(WEB_ROOT)/build/index.html
CONFIG=dist/config/base.yaml
NPM_INSTALL=$(WEB_ROOT)/node_modules/.faux-npm-install

all: $(SERVER) flutter

dist_dir: dist
	mkdir -p dist

server: $(SERVER)

SERVER_STUB_PATH=backend/.gen/server

GOPATH ?= ~/go
$(GOPATH)/bin/oapi-codegen:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.10.1

SERVER_STUB_FILE=$(SERVER_STUB_PATH)/server.go
$(SERVER_STUB_FILE): openapi/amcrest-viewer.openapi.yml $(GOPATH)/bin/oapi-codegen
	echo "Generating OpenAPI: go"
	mkdir -p $(SERVER_STUB_PATH)
	$(GOPATH)/bin/oapi-codegen -package openapi_server -generate "types,chi-server" openapi/amcrest-viewer.openapi.yml >$(SERVER_STUB_FILE)


$(SERVER): dist_dir $(shell find backend -name '*.go') $(SERVER_STUB_FILE)
	echo "Building server Arch:$(GOARCH) Arm:$(GOARM)"
	cd backend && go build -o amcrest-server .

$(NPM_INSTALL): $(WEB_ROOT)/package-lock.json $(WEB_ROOT)/node_modules
	echo "Running NPM install"
	cd $(WEB_ROOT) && npm install
	touch $(NPM_INSTALL)

$(FRONTEND): $(NPM_INSTALL) $(shell find $(WEB_ROOT)/src)  $(shell find $(WEB_ROOT)/public)
	echo "Building frontend"
	cd $(WEB_ROOT) && npm run build

frontend: $(FRONTEND)

flutter: $(CLIENT_STUB_FILE)
	cd frontend-flutter && flutter pub get
	cd frontend-flutter && flutter build web

$(CONFIG): backend/config/base.yaml
	mkdir -p dist/config
	cp -R backend/config dist/
	sed -i 's/test_data/data/g' dist/config/base.yaml
	printf "web:\n  frontend: frontend\n" >>dist/config/base.yaml

dist: dist_dir $(CONFIG) $(SERVER) $(FRONTEND)
	mkdir -p dist/data/db
	mkdir -p dist/data/files
	cp $(SERVER) dist/amcrest-server
	cp -R $(WEB_ROOT)/build dist/frontend

docker:
	docker build -t amcrest-server:current .

clean: 
	rm -rf dist
	rm backend/amcrest-viewer
	rm -rf $(WEB_ROOT)/build


CLIENT_STUB_FILE=frontend-flutter/openapi/.gen/amcrest_viewer/lib/api/default_api.dart
$(CLIENT_STUB_FILE): openapi/amcrest-viewer.openapi.yml
	echo "Generating OpenAPI: dart"
	mkdir -p frontend-flutter/build/.openapi
	cp $< frontend-flutter/build/.openapi/openapi.yaml
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
		-i /local/$< \
		-g dart \
		-o /local/frontend-flutter/openapi/.gen/amcrest_viewer

openapi-gen: $(SERVER_STUB_FILE) $(CLIENT_STUB_FILE)


.PHONY=distdir dist clean npm-install server frontend all docker openapi-gen
