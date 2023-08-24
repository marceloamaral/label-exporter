FROM golang:1.20 AS builder

WORKDIR /go/src/github.com/marceloamaral/label-exporter

COPY . .

RUN make build

RUN cp bin/label-exporter /go/bin/label-exporter

# create a smaller image
FROM scratch

COPY --from=builder /go/bin/label-exporter /go/bin/label-exporter

LABEL name=kabel-exporter

ENTRYPOINT ["/go/bin/label-exporter"]