FROM golang:1.19-alpine as gobuild
RUN apk update && apk add --no-cache build-base
WORKDIR /app
COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download
COPY backend/ backend/
COPY openapi openapi
COPY Makefile .
RUN SERVER=/app/amcrest-server make server

FROM node:16-alpine as nodebuild

WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
COPY frontend/yarn.lock ./
RUN yarn install

COPY frontend/ .
RUN yarn build

FROM ghcr.io/cirruslabs/flutter:3.10.5 as flutterbuild
WORKDIR /app
COPY frontend/flutter/pubspec.yaml frontend/flutter/pubspec.lock frontend/flutter/
COPY openapi openapi
COPY frontend/flutter/ frontend/flutter/
COPY Makefile .
RUN make flutter



FROM alpine:3.14
RUN apk update
RUN apk add  tzdata ffmpeg


WORKDIR /app
COPY --from=gobuild /app/amcrest-server /app/amcrest-server
COPY --from=nodebuild /app/build /app/frontend
COPY --from=flutterbuild /app/frontend/flutter/build/web /app/frontend/flutter
COPY backend/config /app/config


RUN mkdir -p /app/data/db
RUN mkdir -p /app/data/files

ENV DB_PATH=/app/db/cam.db
ENV PORT=9000

EXPOSE 9000
EXPOSE 2121

CMD ["/app/amcrest-server", "--data-dir", "/app/data", "--frontend-dir", "/app/frontend"]
