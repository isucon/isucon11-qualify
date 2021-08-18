package main

import (
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func NewAWSSession() (*session.Session, error) {
	baseSess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	cfg := aws.NewConfig().
		WithCredentials(ec2rolecreds.NewCredentials(baseSess)).
		WithRegion("ap-northeast-1")
	return session.NewSession(cfg)
}

func GetAZName(out *ec2.DescribeAvailabilityZonesOutput, id string) string {
	for _, az := range out.AvailabilityZones {
		if *az.ZoneId == id {
			return *az.ZoneName
		}
	}
	return ""
}

func GetPublicIP(svc *ec2metadata.EC2Metadata) (string, error) {
	return svc.GetMetadata("public-ipv4")
}

func GetVPC(svc *ec2metadata.EC2Metadata) (string, error) {
	s, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var unexpectedNames []string
	for _, i := range s {
		if i.Name == "eth0" {
			return svc.GetMetadata(fmt.Sprintf("network/interfaces/macs/%s/vpc-id", i.HardwareAddr.String()))
		}
		unexpectedNames = append(unexpectedNames, i.Name)
	}
	return "", fmt.Errorf("no expected network interface (%v)", unexpectedNames)
}

func DescribeInstances(svc *ec2.EC2, vpc string) ([]*ec2.DescribeInstancesOutput, error) {
	var s []*ec2.DescribeInstancesOutput

	err := svc.DescribeInstancesPages(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("network-interface.vpc-id"),
				Values: []*string{aws.String(vpc)},
			},
		},
	}, func(out *ec2.DescribeInstancesOutput, lastPage bool) bool {
		s = append(s, out)
		return true
	})
	return s, err
}

func DescribeVolumes(svc *ec2.EC2, instances []*ec2.DescribeInstancesOutput) ([]*ec2.DescribeVolumesOutput, error) {
	var s []*ec2.DescribeVolumesOutput

	var instanceIDs []*string
	for _, out := range instances {
		for _, res := range out.Reservations {
			for _, i := range res.Instances {
				instanceIDs = append(instanceIDs, i.InstanceId)
			}
		}
	}

	err := svc.DescribeVolumesPages(&ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: instanceIDs,
			},
		},
	}, func(out *ec2.DescribeVolumesOutput, lastPage bool) bool {
		s = append(s, out)
		return true
	})
	return s, err
}

func DescribeNetworkInterfaces(svc *ec2.EC2, vpc string) ([]*ec2.DescribeNetworkInterfacesOutput, error) {
	var s []*ec2.DescribeNetworkInterfacesOutput

	err := svc.DescribeNetworkInterfacesPages(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpc)},
			},
		},
	}, func(out *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		s = append(s, out)
		return true
	})
	return s, err
}

func DescribeSecurityGroups(svc *ec2.EC2, instances []*ec2.DescribeInstancesOutput) ([]*ec2.DescribeSecurityGroupsOutput, error) {
	var s []*ec2.DescribeSecurityGroupsOutput

	var ids []*string
	idsUnique := make(map[string]struct{})
	for _, o := range instances {
		for _, r := range o.Reservations {
			for _, i := range r.Instances {
				for _, sg := range i.SecurityGroups {
					if _, ok := idsUnique[*sg.GroupId]; !ok {
						ids = append(ids, sg.GroupId)
						idsUnique[*sg.GroupId] = struct{}{}
					}
				}
			}
		}
	}

	err := svc.DescribeSecurityGroupsPages(&ec2.DescribeSecurityGroupsInput{
		GroupIds: ids,
	}, func(out *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		s = append(s, out)
		return true
	})

	return s, err
}
