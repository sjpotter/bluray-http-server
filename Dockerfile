FROM golang:buster
RUN apt-get update && apt-get -y install libbluray-dev
WORKDIR /go/src
COPY . /go/src/
RUN go build ./cmd/bluray-server

FROM ubuntu:19.10
RUN apt-get update && apt-get -y install libbluray2
COPY --from=0 /go/src/bluray-server /
CMD /bluray-server -port 8090
