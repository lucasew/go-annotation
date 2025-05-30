FROM golang:1.24-alpine AS build-env

WORKDIR /go/src/go-annotation

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

ARG VERSION_LONG
ENV VERSION_LONG=$VERSION_LONG

ARG VERSION_GIT
ENV VERSION_GIT=$VERSION_GIT

RUN go build -v -o go-annotation ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates iptables iproute2 ip6tables

COPY --from=build-env /go/src/go-annotation/go-annotation /usr/local/bin

ENTRYPOINT [ "/usr/local/bin/go-annotation" ]

