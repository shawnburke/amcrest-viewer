
SERVER=dist/amcrest-viewer
FRONTEND=dist/frontend/index.html
WEB_ROOT=web/frontend
CONFIG=dist/config/base.yaml
ROOT=$(pwd)


NPM_INSTALL=$(WEB_ROOT)/node_modules/.faux-npm-install
WEBPACK=$(WEB_ROOT)/build/index.html
CONFIG=dist/config/base.yaml


dist: $(CONFIG) $(SERVER) $(FRONTEND)
	mkdir -p dist/data/db
	mkdir -p dist/data/files

$(CONFIG):
	mkdir -p dist/config
	cp -R config dist


$(SERVER): $(shell find . -name '*.go') $(CONFIG)
	go build -o dist/amcrest-viewer .

$(NPM_INSTALL): $(WEB_ROOT)/package-lock.json 
	cd $(WEB_ROOT) && npm install
	touch $(NPM_INSTALL)

$(WEBPACK): $(NPM_INSTALL) $(shell find $(WEB_ROOT)/src)  $(shell find $(WEB_ROOT)/public)
	cd $(WEB_ROOT) && npm run build
	
$(FRONTEND): $(WEBPACK)
	mkdir -p dist/frontend
	cp -R $(WEB_ROOT)/build/. dist/frontend

clean: 
	rm -rf dist
	rm amcrest-viewer
	rm -rf $(WEB_ROOT)/build


.PHONY=distdir dist clean npm-install