FROM golang:1.25-alpine AS build-env

WORKDIR /go/src/rotulador

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

ARG VERSION_LONG
ENV VERSION_LONG=$VERSION_LONG

ARG VERSION_GIT
ENV VERSION_GIT=$VERSION_GIT

RUN go build -v -o rotulador ./cmd/rotulador

FROM alpine:3.22

RUN apk add --no-cache ca-certificates iptables iproute2 ip6tables

COPY --from=build-env /go/src/rotulador/rotulador /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/rotulador" ]

