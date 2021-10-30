FROM golang:1.16.3 AS builder

ENV GOPRIVATE=github.com/blinkops

WORKDIR /go/src/github.com/blinkops/blink-ssh

COPY go.mod go.sum ./
RUN go mod download
COPY .. .

RUN go build -tags netgo -ldflags '-w -extldflags "-static"' -o /go/bin/blink-ssh ./plugin.go

FROM alpine:3.14.0 AS plugin

WORKDIR /blink-http
COPY --from=builder /go/bin/blink-ssh .
COPY config.yaml plugin.yaml ./
COPY actions actions/


# Expose the gRPC port.
EXPOSE 1337

RUN chmod a+x blink-ssh

ENTRYPOINT ./blink-ssh
