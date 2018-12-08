FROM golang:latest
WORKDIR $GOPATH/src/socket
MAINTAINER George "george@betterde.com"
ADD . $GOPATH/src/socket
RUN export GO111MODULE=on && go mod download && go build -o socket main.go
EXPOSE 6064
ENTRYPOINT  ["./socket start -d"]