FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/github.com/mivinci/shorturl/
COPY . .

RUN go get -d -v
RUN CGO_ENABLED=0 go build -o /usr/bin/shorturl

FROM alpine:3.12
WORKDIR /root
COPY --from=builder /usr/bin/shorturl /usr/bin/shorturl
COPY html /root/html
EXPOSE 5000
ENTRYPOINT [ "shorturl", "-domain", "https://l.xjj.pub"]
