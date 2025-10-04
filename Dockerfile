# Running Pano in Docker is Experimental - not recommended for production use!

# Example of usage:
# docker build -t pano .
# docker run --name pano1 --entrypoint panotool pano --datadir=/var/pano genesis fake 1
# docker run --volumes-from pano1 -p 5050:5050 -p 5050:5050/udp -p 18545:18545 pano --fakenet 1/1 --http --http.addr=0.0.0.0

FROM golang:1.24 AS builder

RUN apt-get update && apt-get install -y git musl-dev make

WORKDIR /go/Pano
COPY . .

ARG GOPROXY
RUN go mod download
RUN make all


FROM golang:1.24

COPY --from=builder /go/Pano/build/panod /usr/local/bin/
COPY --from=builder /go/Pano/build/panotool /usr/local/bin/

EXPOSE 18545 18546 5050 5050/udp

VOLUME /var/pano

ENTRYPOINT ["panod", "--datadir=/var/pano"]
