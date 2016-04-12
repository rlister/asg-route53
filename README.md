# asg-route53

## Installation

### Binaries for linux and OSX

Download a static binary from
https://github.com/rlister/asg-route53/releases.

### Docker

```
docker pull rlister/asg-route53:latest
```

Unless running under an IAM role, you will need to pass in your AWS
credentials to make AWS changes, for example:

```
docker run \
  -e AWS_REGION \
  -e AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY \
  rlister/asg-route53 my-security-group
```

### Build from source

Build your own using your favourite `go build` command, for example:

```
go build ./asg-route53.go
```

### Making a new release

Binaries in releases are built using
[goxc](https://github.com/laher/goxc):

```
goxc -t    # first use
goxc bump
goxc -bc="linux darwin"
```

## Docker image

I build a tiny docker image from scratch as follows:

```
version=0.0.1
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo asg-route53.go
docker build -t rlister/asg-route53:${version} .
docker tag -f rlister/asg-route53:${version} rlister/asg-route53:latest
docker push rlister/asg-route53
```