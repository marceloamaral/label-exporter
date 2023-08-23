FROM golang:1.20

WORKDIR /go/src/github.com/marceloamaral/label-exporter

COPY . .

RUN make build

RUN cp bin/label-exporter /bin/label-exporter

ENTRYPOINT ["/bin/label-exporter"]