FROM resin/rpi-raspbian:latest

# RUN apt update && apt upgrade
COPY amcrest-viewer-pi .
EXPOSE 9000
CMD ./amcrest-viewer-pi /files 9000
