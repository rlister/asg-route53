package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"os"
	"strings"
)

// error handler
func check(e error) {
	if e != nil {
		panic(e.Error())
	}
}

// get current instance ID from metadata
func getInstanceId() *string {
	metadata := ec2metadata.New(session.New())
	id, err := metadata.GetMetadata("instance-id")
	check(err)
	return &id
}

// lookup the autoscaling group for an instance
func getAutoscalingGroup(instance_id *string) *string {
	svc := ec2.New(session.New())

	// get tags for the instance
	params := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("resource-id"),
				Values: []*string{
					aws.String(*instance_id),
				},
			},
		},
	}
	resp, err := svc.DescribeTags(params)
	check(err)

	// search for autoscaling group tag
	for _, tag := range resp.Tags {
		if *tag.Key == "aws:autoscaling:groupName" {
			return tag.Value
		}
	}

	// failure
	return nil
}

// return IDs of instances in autoscaling group
func getAutoscalingInstances(asg *string) []*string {
	svc := autoscaling.New(session.New())

	// get instances for autoscaling group
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(*asg),
		},
	}
	resp, err := svc.DescribeAutoScalingGroups(params)
	check(err)

	// get array of instance IDs
	instances := resp.AutoScalingGroups[0].Instances
	ids := make([]*string, len(instances))
	for i, instance := range instances {
		ids[i] = instance.InstanceId
	}
	return ids
}

// get IPs for given instance IDs
func getInstanceIpAddresses(ids []*string, public_ip bool) []*string {
	svc := ec2.New(session.New())

	params := &ec2.DescribeInstancesInput{
		InstanceIds: ids,
	}
	resp, err := svc.DescribeInstances(params)
	check(err)

	var ips []*string
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
		    if public_ip {
                        ips = append(ips, instance.PublicIpAddress)
		    } else {
			ips = append(ips, instance.PrivateIpAddress)
		    }
		}
	}
	return ips
}

// get id for named hosted zone
func getHostedZones(name string) *string {
	svc := route53.New(session.New())

	params := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(name),
	}
	resp, err := svc.ListHostedZonesByName(params)
	check(err)

	return resp.HostedZones[0].Id
}

// update record with given IPs
func changeRecord(zone *string, name *string, rectype *string, ips []*string) {
	svc := route53.New(session.New())

	// transform IPs into resource records
	rrecords := make([]*route53.ResourceRecord, len(ips))
	for i, ip := range ips {
		rrecords[i] = &route53.ResourceRecord{
			Value: aws.String(*ip),
		}
	}

	// update resource records
	params := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(*zone),
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(*name),
						Type:            aws.String(*rectype),
						TTL:             aws.Int64(180),
						ResourceRecords: rrecords,
					},
				},
			},
		},
	}
	resp, err := svc.ChangeResourceRecordSets(params)
	check(err)
	fmt.Println(resp)
}

// parse out zone (last two elements) from DNS record
func parseZone(record string) string {
	parts := strings.Split(record, ".")
	last2 := parts[len(parts)-2 : len(parts)]
	return strings.Join(last2, ".")
}

func main() {
	// cmdline flags
	public_ip := flag.Bool("public", false, "Use instance public id")
	asg := flag.String("asg", "", "autoscaling group name")
	rectype := flag.String("type", "SRV", "type of DNS records to create")
	priority := flag.Int("priority", 0, "priority for SRV records")
	weight := flag.Int("weight", 0, "weight for SRV records")
	port := flag.Int("port", 2380, "port for SRV records")
	flag.Parse()

	// DNS record to update is first cmdline arg
	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s DNS_RECORD\n", os.Args[0])
		os.Exit(1)
	}
	record := flag.Args()[0]

	// get autoscaling group
	if *asg == "" {
		asg = getAutoscalingGroup(getInstanceId())
	}

	// get instance IPs
	ids := getAutoscalingInstances(asg)
	ips := getInstanceIpAddresses(ids, *public_ip)

	// mangle SRV records into required format
	if *rectype == "SRV" {
		for i, ip := range ips {
			srv := fmt.Sprintf("%d %d %d %s", *priority, *weight, *port, *ip)
			ips[i] = &srv
		}
	}

	// update DNS
	zone_id := getHostedZones(parseZone(record))
	changeRecord(zone_id, &record, rectype, ips)
}
