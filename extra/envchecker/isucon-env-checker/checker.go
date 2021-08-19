package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cenkalti/backoff/v4"
)

type CheckConfig struct {
	AMI []string `json:"ami_id"`
	AZ  string   `json:"az_id"`
}

type checker struct {
	AllowedAMI []string
	ExpectedAZ string

	InstanceIP    string
	InstanceID    string
	InstanceVPCID string

	Name string

	DescribeInstances         []*ec2.DescribeInstancesOutput
	DescribeVolumes           []*ec2.DescribeVolumesOutput
	DescribeNetworkInterfaces []*ec2.DescribeNetworkInterfacesOutput
	DescribeSecurityGroups    []*ec2.DescribeSecurityGroupsOutput

	DescribeAvailabilityZones *ec2.DescribeAvailabilityZonesOutput

	failures []string

	adminLog    *bytes.Buffer
	adminLogger *log.Logger
}

type Result struct {
	Name         string
	Passed       bool
	IPAddress    string
	Message      string
	AdminMessage string
	RawData      string
}

func Check(cfg CheckConfig) Result {
	buf := new(bytes.Buffer)
	logger := log.New(buf, "", log.LstdFlags)
	c := &checker{
		AllowedAMI: cfg.AMI,
		ExpectedAZ: cfg.AZ,

		adminLog:    buf,
		adminLogger: logger,
	}
	if err := c.loadAWS(); err != nil {
		c.adminLogger.Printf("loading AWS data: %+v", err)
		raw, _ := json.Marshal(c)
		return Result{
			Name:         c.name(),
			Passed:       false,
			Message:      "AWS との通信でエラーが発生しました",
			AdminMessage: c.adminLog.String(),
			RawData:      string(raw),
		}
	}

	c.checkAll()

	raw, _ := json.Marshal(c)
	return Result{
		Name:         c.name(),
		Passed:       len(c.failures) == 0,
		IPAddress:    c.InstanceIP,
		Message:      c.message(),
		AdminMessage: c.adminLog.String(),
		RawData:      string(raw),
	}
}

func (c *checker) loadAWS() error {
	sess, err := NewAWSSession()
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	ec2md := ec2metadata.New(sess)
	ec2client := ec2.New(sess)

	err = backoff.Retry(func() error {
		c.InstanceIP, err = GetPublicIP(ec2md)
		return err
	}, newBackoff())
	if err != nil {
		return fmt.Errorf("GetPublicIP: %w", err)
	}
	c.InstanceID, err = GetInstanceID(ec2md)
	if err != nil {
		return fmt.Errorf("GetInstanceID: %w", err)
	}
	c.InstanceVPCID, err = GetVPC(ec2md)
	if err != nil {
		return fmt.Errorf("GetVPC: %w", err)
	}

	c.DescribeInstances, err = DescribeInstances(ec2client, c.InstanceVPCID)
	if err != nil {
		return fmt.Errorf("DescribeInstances: %w", err)
	}
	c.DescribeVolumes, err = DescribeVolumes(ec2client, c.DescribeInstances)
	if err != nil {
		return fmt.Errorf("DescribeVolumes: %w", err)
	}
	c.DescribeNetworkInterfaces, err = DescribeNetworkInterfaces(ec2client, c.InstanceVPCID)
	if err != nil {
		return fmt.Errorf("DescribeNetworkInterfaces: %w", err)
	}
	c.DescribeSecurityGroups, err = DescribeSecurityGroups(ec2client, c.DescribeInstances)
	if err != nil {
		return fmt.Errorf("DescribeSecurityGroups: %w", err)
	}
	c.DescribeAvailabilityZones, err = ec2client.DescribeAvailabilityZones(nil)
	if err != nil {
		return fmt.Errorf("DescribeAvailabilityZones: %w", err)
	}
	return nil
}

func (c *checker) addFailure(format string, a ...interface{}) {
	c.failures = append(c.failures, fmt.Sprintf(format, a...))
}

func (c *checker) message() string {
	if len(c.failures) == 0 {
		return "全てのチェックをパスしました"
	}
	return strconv.Itoa(len(c.failures)) + "個の問題があります\n" + strings.Join(c.failures, "\n")
}

func (c *checker) name() string {
	for _, o := range c.DescribeInstances {
		for _, r := range o.Reservations {
			for _, i := range r.Instances {
				id := *i.InstanceId
				if id != c.InstanceID {
					continue
				}
				for _, t := range i.Tags {
					if *t.Key != "Name" {
						continue
					}
					name := *t.Value
					if checkName, ok := checkNameByInstanceName[name]; ok {
						return checkName
					} else {
						return "qualify-unknown"
					}
				}
			}
		}
	}
	return "qualify-unknown"
}
