set -e


docker build -t amcrest-viewer-build:current .

if [ -n "$1" ]
then
   echo Using env vile
   envfile="--env-file $1"
fi

docker run --rm $envfile  -v /home/shawn/docker/amcrest-viewer/src/amcrest-viewer:/usr/src/myapp -w /usr/src/myapp amcrest-viewer-build:current make dist 
