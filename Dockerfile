FROM golang:1.8
MAINTAINER Ric Lister <rlister@gmail.com>

ADD certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /go/src/asg-route53
COPY . .

RUN go get
RUN go build -o /usr/bin/asg-route53

ENTRYPOINT [ "asg-route53" ]