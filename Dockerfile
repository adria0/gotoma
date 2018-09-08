FROM golang:alpine
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh build-base

RUN mkdir -p /go/src/github.com/TomahawkEthBerlin/gotoma 
ADD . /go/src/github.com/TomahawkEthBerlin/gotoma/
WORKDIR /go/src/github.com/TomahawkEthBerlin/gotoma 
RUN go get
RUN go build -o /gotoma .
ENTRYPOINT ["/gotoma","serve"]