
ROOT=$(pwd)
SERVER ?= build/server/av-server
WEB_ROOT=frontend/js
FRONTEND=$(WEB_ROOT)/build/index.html
CONFIG=dist/config/base.yaml
NPM_INSTALL=$(WEB_ROOT)/node_modules/.faux-npm-install

all: $(SERVER) flutter



##
## OPEN API STUFF
##


SERVER_STUB_FILE=$(SERVER_STUB_PATH)/server.go
$(SERVER_STUB_FILE): openapi/amcrest-viewer.openapi.yaml $(GOPATH)/bin/oapi-codegen
	echo "Generating OpenAPI: go"
	mkdir -p $(SERVER_STUB_PATH)
	$(GOPATH)/bin/oapi-codegen -package openapi_server -generate "types,chi-server" openapi/amcrest-viewer.openapi.yaml >$(SERVER_STUB_FILE)


CLIENT_STUB_FILE_FLUTTER=frontend/flutter/openapi/.gen/amcrest_viewer/lib/api/default_api.dart
CLIENT_STUB_FILE_JS=frontend/js/src/.gen/src/index.js


$(CLIENT_STUB_FILE_FLUTTER): openapi/amcrest-viewer.openapi.yaml
	echo "Generating OpenAPI: dart"
	mkdir -p frontend/flutter/build/.openapi
	cp $< frontend/flutter/build/.openapi/openapi.yaml
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
		-i /local/$< \
		-g dart \
		-o /local/frontend/flutter/openapi/.gen/amcrest_viewer
	
$(CLIENT_STUB_FILE_JS): openapi/amcrest-viewer.openapi.yaml
	echo "Generating OpenAPI: JS"
	mkdir -p frontend/src/.gen
	cp $< frontend/build/openapi.yaml
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
		-i /local/frontend/build/openapi.yaml \
		-g javascript \
		-o /local/frontend/src/.gen/
	
openapi-client-js: $(CLIENT_STUB_FILE_JS)
openapi-client-flutter: $(CLIENT_STUB_FILE_FLUTTER)


openapi-gen: $(SERVER_STUB_FILE) openapi-client-js openapi-client-flutter

server: $(SERVER)
SERVER_ARM64=build/server/av-server-aarch64


$(SERVER_ARM64): 
	SERVER=$(SERVER_ARM64) GOOS=linux GOARCH=arm64 $(MAKE) server

server-arm64: $(SERVER_ARM64)

SERVER_STUB_PATH=backend/.gen/server

GOPATH ?= ~/go
$(GOPATH)/bin/oapi-codegen:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.10.1


$(SERVER): $(shell find backend -name '*.go') $(SERVER_STUB_FILE)
	echo "Building server Arch:$(GOARCH) OS:$(GOOS)"
	mkdir -p build
	cd backend && go build -o .amcrest-server-build .
	rm -rf $(SERVER)
	mkdir -p $$(dirname $(SERVER))
	mv backend/.amcrest-server-build $(SERVER)

$(NPM_INSTALL): $(WEB_ROOT)/package-lock.json $(WEB_ROOT)/node_modules
	echo "Running NPM install"
	cd $(WEB_ROOT) && npm install
	touch $(NPM_INSTALL)

$(FRONTEND): $(NPM_INSTALL) $(shell find $(WEB_ROOT)/src)  $(shell find $(WEB_ROOT)/public) openapi-client-js
	echo "Building frontend"
	cd $(WEB_ROOT) && npm run build
	cp -R $(WEB_ROOT)/build/ build/js

frontend-js: $(FRONTEND)

test-backend: $(SERVER)
	cd backend && go test ./...

flutter-tar: flutter-web.tar.gz

flutter-web.tar.gz: flutter-web
	tar -czf flutter-web.tar.gz frontend/flutter/build/web

$(CONFIG): backend/config/base.yaml
	mkdir -p dist/config
	cp -R backend/config dist/
	sed -i 's/test_data/data/g' dist/config/base.yaml
	printf "web:\n  frontend: frontend\n" >>dist/config/base.yaml

dist: $(CONFIG) $(SERVER) $(FRONTEND)
	mkdir -p dist/data/db
	mkdir -p dist/data/files
	cp $(SERVER) dist/amcrest-server
	cp -R $(WEB_ROOT)/build dist/frontend

clean: 
	rm -rf build




FLUTTER_TARGET=~/flutter/bin/flutter
FLUTTER_VERSION=3.7.5

$(FLUTTER_TARGET):
	wget -O /tmp/flutter.tar.gz https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_$(FLUTTER_VERSION)-stable.tar.xz
	# expand to an child directory to avoid accidental pollution of homedir
	mkdir -p ~/flutter-sdk
	tar -xf /tmp/flutter.tar.gz -C ~/flutter-sdk
	mv ~/flutter-sdk /flutter
	~/flutter/bin/flutter precache
	echo "SET PATH:"
	echo "export PATH=$$PATH:~/flutter/bin"

flutter-install: $(FLUTTER_TARGET)
flutter-webserver: flutter-install
	cd frontend/flutter && flutter run -d web-server --web-hostname 0.0.0.0


FLUTTER_WEB=build/flutter/web/main.dart.js
flutter-deps: $(find frontend/flutter/lib -name "*.dart") $(CLIENT_STUB_FILE_FLUTTER)
flutter-linux: flutter-deps
	@echo "Building flutter"
	cd frontend/flutter && flutter build linux

$(FLUTTER_WEB): $(find frontend/flutter/lib -name "*.dart") $(CLIENT_STUB_FILE_FLUTTER)
	@echo "Fetching deps"
	cd frontend/flutter && flutter pub get
	@echo "Building flutter web"
	cd frontend/flutter && flutter build web
	mkdir -p build/flutter
	cp -R frontend/flutter/build/web/ build/flutter/

flutter: flutter-linux flutter-web
flutter-web: $(FLUTTER_WEB)

docker: server server-arm64 flutter-web
	docker build -t amcrest-server:current -f Dockerfile_build .

.PHONY=distdir dist clean npm-install server frontend all docker flutter-install flutter-webserver openapi-gen test-backend
