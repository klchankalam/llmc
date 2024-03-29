FROM golang:1.12.5
COPY . /go
WORKDIR /go/src/app
RUN go get -d -v ./...
RUN go get -d -v -t ../distancehelper ../requesthandler
RUN go test ../distancehelper ../requesthandler
RUN go install -v ./...
#&& RUN go get github.com/derekparker/delve/src/dlv
#&& RUN go build -i -v -gcflags "all=-N -l" ./...

CMD ["app"]