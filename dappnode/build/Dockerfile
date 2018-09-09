FROM golang:alpine as build
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh build-base

RUN mkdir -p /go/src/github.com/TomahawkEthBerlin/gotoma 
ADD . /go/src/github.com/TomahawkEthBerlin/gotoma/
WORKDIR /go/src/github.com/TomahawkEthBerlin/gotoma 
RUN go get
RUN go build -o /gotoma .

FROM golang:alpine
RUN mkdir -p /app 
COPY --from=build /gotoma /app/gotoma
WORKDIR /app 

ENTRYPOINT ["/app/gotoma","serve"]

