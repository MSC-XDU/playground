FROM golang:1.7

COPY . $GOPATH/src/app

WORKDIR $GOPATH/src/app

RUN go get -d -v
RUN rm -rf $GOPATH/src/github.com/docker/docker/vendor/github.com/docker/go-units

CMD go run main.go