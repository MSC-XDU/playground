FROM golang:1.7

ENV SRC_DIR $GOPATH/src/github.com/MSC-XDU/playground

ENV PLAYGROUND_TEMP_DIR /tmp
VOLUME /tmp

COPY . $SRC_DIR

WORKDIR $SRC_DIR

EXPOSE 80

CMD go run main.go