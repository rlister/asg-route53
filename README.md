# asg-route53

This was written primarily for CoreOS Etcd instances created as part
of an AWS Autoscaling Group, using the method described at
https://github.com/MonsantoCo/etcd-aws-cluster.

Rather than adding the `etcd-aws-cluster` unit to every `proxy` client
of your `etcd` cluster, `asg-route53` will add a `SRV` record to DNS
for all the `etcd` hosts, so they can be discovered by proxy clients,
as described at
https://coreos.com/etcd/docs/latest/clustering.html#discovery.

## How does it work?

- Looks up instance metadata and queries `aws-sdk` for autoscaling group.
- Looks up all instances for this autoscaling group.
- Updates given DNS record in Route53 to contain all ASG instances.

## Usage

```
docker pull rlister/asg-route53:latest
docker run rlister/asg-route53:latest _etcd-server._tcp.example.com
```

## IAM policy required

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
                "route53:ChangeResourceRecordSets"
            ],
            "Resource": "arn:aws:route53:::hostedzone/$zone_id",
            "Effect": "Allow"
        },
        {
            "Action": [
                "route53:GetChange"
            ],
            "Resource": "arn:aws:route53:::change/*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "route53:ListHostedZonesByName"
            ],
            "Resource": "*",
            "Effect": "Allow"
        },
        {
            "Action": [
                "autoscaling:DescribeAutoScalingGroups",
                "ec2:DescribeInstances",
                "ec2:DescribeTags"
            ],
            "Resource": "*",
            "Effect": "Allow"
        }
    ]
}
```

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