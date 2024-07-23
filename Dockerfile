FROM golang:1.22 as build
WORKDIR /usr/src/app
COPY go.mod go.sum makefile ./
RUN make install

COPY src ./src
RUN make build

FROM debian:12-slim
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y ca-certificates dumb-init && \
    mkdir /app
WORKDIR /app
COPY --from=build /usr/src/app/bin/osm-download-workflow .
ENTRYPOINT ["/usr/bin/dumb-init", "--"]