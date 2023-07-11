
ROOT=$(pwd)
SERVER ?= build/server/av-server
CONFIG=dist/config/base.yaml
WEB_ROOT=frontend/js
FLUTTER_ROOT=frontend/flutter

all: $(SERVER) frontend-flutter frontend-js

GOPATH ?= ~/go

##
## OPEN API STUFF
##
$(GOPATH)/bin/oapi-codegen:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.10.1

SERVER_STUB_PATH=backend/.gen/server
SERVER_STUB_FILE=$(SERVER_STUB_PATH)/server.go
$(SERVER_STUB_FILE): openapi/amcrest-viewer.openapi.yaml $(GOPATH)/bin/oapi-codegen
	echo "Generating OpenAPI: go"
	mkdir -p $(SERVER_STUB_PATH)
	$(GOPATH)/bin/oapi-codegen -package openapi_server -generate "types,chi-server" openapi/amcrest-viewer.openapi.yaml >$(SERVER_STUB_FILE)


CLIENT_STUB_FILE_FLUTTER=$(FLUTTER_ROOT)/openapi/.gen/amcrest_viewer/lib/api/default_api.dart
CLIENT_STUB_FILE_JS=$(WEB_ROOT)/.gen/src/index.js


$(CLIENT_STUB_FILE_FLUTTER): openapi/amcrest-viewer.openapi.yaml
	echo "Generating OpenAPI: dart"
	mkdir -p $(FLUTTER_ROOT)/build/.openapi
	cp $< $(FLUTTER_ROOT)/build/.openapi/openapi.yaml
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
		-i /local/$< \
		-g dart \
		-o /local/$(FLUTTER_ROOT)/openapi/.gen/amcrest_viewer
	
$(CLIENT_STUB_FILE_JS): openapi/amcrest-viewer.openapi.yaml
	echo "Generating OpenAPI: JS"
	mkdir -p $(WEB_ROOT)/build
	cp $< $(WEB_ROOT)/build/openapi.yaml
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
		-i /local/$(WEB_ROOT)/build/openapi.yaml \
		-g javascript \
		-o /local/$(WEB_ROOT)/.gen/
	
openapi-client-js: $(CLIENT_STUB_FILE_JS)
openapi-client-flutter: $(CLIENT_STUB_FILE_FLUTTER)
openapi-backend: $(SERVER_STUB_FILE) 
openapi-gen: openapi-backend openapi-client-js openapi-client-flutter


#
# Server stuff
#

server: $(SERVER)
SERVER_ARM64=build/server/av-server-aarch64

$(SERVER_ARM64): 
	SERVER=$(SERVER_ARM64) GOOS=linux GOARCH=arm64 $(MAKE) server

server-arm64: $(SERVER_ARM64)

$(SERVER): $(shell find backend -name '*.go') $(SERVER_STUB_FILE)
	echo "Building server Arch:$(GOARCH) OS:$(GOOS)"
	mkdir -p build
	cd backend && go build -o .amcrest-server-build .
	rm -rf $(SERVER)
	mkdir -p $$(dirname $(SERVER))
	mv backend/.amcrest-server-build $(SERVER)

#
# JS-frontend stuff
#

FRONTEND=$(WEB_ROOT)/build/index.html
NPM_INSTALL=$(WEB_ROOT)/node_modules/.faux-npm-install


$(NPM_INSTALL): $(WEB_ROOT)/package-lock.json
	echo "Running NPM install"
	cd $(WEB_ROOT) && npm install
	touch $(NPM_INSTALL)

$(FRONTEND): $(NPM_INSTALL) $(shell find $(WEB_ROOT)/src)  $(shell find $(WEB_ROOT)/public) openapi-client-js
	echo "Building frontend"
	cd $(WEB_ROOT)/.gen && npm install && npm link
	cd $(WEB_ROOT) && npm link .gen
	cd $(WEB_ROOT) && PUBLIC_URL=js npm run build
	mkdir -p build
	cp -R $(WEB_ROOT)/build/ build/js

frontend-js: $(FRONTEND)

test-backend: $(SERVER)
	cd backend && go test ./...

#
# Flutter stuff
#

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
	cd $(FLUTTER_ROOT) && flutter run -d web-server --web-hostname 0.0.0.0


FLUTTER_WEB=build/flutter/web/main.dart.js
flutter-deps: $(find $(FLUTTER_ROOT)/lib -name "*.dart") $(CLIENT_STUB_FILE_FLUTTER)
flutter-linux: flutter-deps
	@echo "Building flutter"
	cd $(FLUTTER_ROOT) && flutter build linux

$(FLUTTER_WEB): $(find $(FLUTTER_ROOT)/lib -name "*.dart") $(CLIENT_STUB_FILE_FLUTTER)
	@echo "Fetching deps"
	cd $(FLUTTER_ROOT) && flutter pub get
	@echo "Building flutter web"
	cd $(FLUTTER_ROOT) && flutter build web
	mkdir -p build
	cp -R $(FLUTTER_ROOT)/build/web/ build/flutter

flutter: flutter-linux frontend-flutter
frontend-flutter: $(FLUTTER_WEB)

flutter-tar: flutter-web.tar.gz

flutter-web.tar.gz: frontend-flutter
	tar -czf flutter-web.tar.gz $(FLUTTER_ROOT)/build/web

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
	rm -rf $(WEB_ROOT)/build
	rm -rf $(FLUTTER_ROOT)/build
	rm -rf backend/.gen



docker: server server-arm64 frontend-flutter frontend-js
	docker build -t amcrest-server:current -f Dockerfile_build .

.PHONY=distdir dist clean npm-install server frontend all docker flutter-install flutter-webserver openapi-gen test-backend
