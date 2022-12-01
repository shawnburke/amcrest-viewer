
ROOT=$(pwd)
SERVER=backend/amcrest-server
WEB_ROOT=frontend
FRONTEND=$(WEB_ROOT)/build/index.html
CONFIG=dist/config/base.yaml
NPM_INSTALL=$(WEB_ROOT)/node_modules/.faux-npm-install

all: $(SERVER) $(FRONTEND)

dist_dir: dist
	mkdir -p dist

server: $(SERVER)

$(SERVER): dist_dir $(shell find backend -name '*.go') 
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


.PHONY=distdir dist clean npm-install server frontend all docker
