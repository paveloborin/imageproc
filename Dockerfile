FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/paveloborin/imageproc
COPY . .

RUN apk update \
    && apk add make \
    && make build-server \
    && mv bin/server /exe

FROM jjanzic/docker-python3-opencv:latest
COPY --from=builder /exe /
COPY ./scripts/* /scripts/

ENTRYPOINT ["/exe"]