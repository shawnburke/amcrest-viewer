
SERVER=dist/amcrest-viewer
FRONTEND=dist/frontend/index.html
WEB_ROOT=web/frontend
CONFIG=dist/config/base.yaml
ROOT=$(pwd)


NPM_INSTALL=$(WEB_ROOT)/node_modules/.faux-npm-install
WEBPACK=$(WEB_ROOT)/build/index.html
CONFIG=dist/config/base.yaml


$(CONFIG):
	mkdir -p dist/config
	cp -R config dist
	sed -i 's/test_data/data/g' dist/config/base.yaml
	printf "web:\n  frontend: frontend\n" >>dist/config/base.yaml

$(SERVER): $(shell find . -name '*.go') 
	echo "Building server Arch:$(GOARCH) Arm:$(GOARM)"
	go build -o $(SERVER) .

$(NPM_INSTALL): $(WEB_ROOT)/package-lock.json 
	echo "Running NPM install"
	cd $(WEB_ROOT) && npm install
	touch $(NPM_INSTALL)

$(WEBPACK): $(NPM_INSTALL) $(shell find $(WEB_ROOT)/src)  $(shell find $(WEB_ROOT)/public)
	echo "Building frontend"
	cd $(WEB_ROOT) && npm run build
	
$(FRONTEND): $(WEBPACK)
	echo "Copy frontend to dist"
	mkdir -p dist/frontend
	cp -R $(WEB_ROOT)/build/. dist/frontend


dist: $(CONFIG) $(SERVER) $(FRONTEND)
	mkdir -p dist/data/db
	mkdir -p dist/data/files


clean: 
	rm -rf dist
	rm amcrest-viewer
	rm -rf $(WEB_ROOT)/build


.PHONY=distdir dist clean npm-install
