from golang:onbuild

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main .

#WORKDIR /root/go/src/github.com/TomahawkEthBerlin/gotoma
#COPY * /root/go/src/github.com/TomahawkEthBerlin/gotoma/


ENTRYPOINT /app/main
