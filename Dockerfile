FROM golang:1.12.5
WORKDIR /go/src/llmc
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
#&& RUN go get github.com/derekparker/delve/cmd/dlv
#&& RUN go build -i -v -gcflags "all=-N -l" ./...

#ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.5.0/wait /wait
#RUN chmod +x /wait

CMD ["llmc"]