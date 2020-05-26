set -e


docker build -t amcrest-viewer-build:current .



docker run --rm -v /home/shawn/docker/amcrest-viewer/src/amcrest-viewer:/usr/src/myapp -w /usr/src/myapp amcrest-viewer-build:current make dist 
