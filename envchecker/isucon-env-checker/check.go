package main

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (c *checker) checkAll() {
	c.checkInstances()
	c.checkVolumes()
	c.checkNetworkInterfaces()
	c.checkSecurityGroups()
}

var (
	checkNameByInstanceName = map[string]string{
		"isucon11-qualify-1": "qualify1",
		"isucon11-qualify-2": "qualify2",
		"isucon11-qualify-3": "qualify3",
	}
	privateAddressByInstanceName = map[string]string{
		"isucon11-qualify-1": "192.168.0.11",
		"isucon11-qualify-2": "192.168.0.12",
		"isucon11-qualify-3": "192.168.0.13",
	}
)

func (c *checker) checkInstances() {
	var count int
	for _, o := range c.DescribeInstances {
		for _, r := range o.Reservations {
			for _, i := range r.Instances {
				count++
				c.checkInstance(i)
			}
		}
	}
	if count > 3 {
		c.addFailure("%d個の EC2 インスタンスが検出されました (3個である必要があります)", count)
	}
}

func (c *checker) isAllowedAMI(id string) bool {
	if len(c.AllowedAMI) == 0 {
		return true
	}
	for _, s := range c.AllowedAMI {
		if id == s {
			return true
		}
	}
	return false
}

func (c *checker) checkInstance(i *ec2.Instance) {
	id := *i.InstanceId
	var nameTagFound bool
	for _, t := range i.Tags {
		if *t.Key != "Name" {
			continue
		}
		nameTagFound = true
		privateAddress, ok := privateAddressByInstanceName[*t.Value]
		if !ok {
			c.addFailure("%s の Name タグが %q です", id, *t.Value)
			break
		}
		if privateAddress != *i.PrivateIpAddress {
			c.addFailure("%s の Private IP が %s です", id, *i.PrivateIpAddress)
		}
	}
	if !nameTagFound {
		c.addFailure("%s に Name タグがありません", id)
	}
	if *i.InstanceType != "c5.large" {
		c.addFailure("%s のインスタンスタイプが %s です (t2.micro である必要があります)", id, *i.InstanceType)
	}
	if !c.isAllowedAMI(*i.ImageId) {
		c.addFailure("%s の AMI が %s です", id, *i.ImageId)
	}
	if c.ExpectedAZ != "" {
		azName := GetAZName(c.DescribeAvailabilityZones, c.ExpectedAZ)
		if *i.Placement.AvailabilityZone != azName {
			c.addFailure("%s のゾーンが %s です (%s (ID: %s) である必要があります)", id, *i.Placement.AvailabilityZone, azName, c.ExpectedAZ)
		}
	}
	if len(i.BlockDeviceMappings) != 1 {
		c.addFailure("%s に %d 個のブロックデバイスが検出されました (1個である必要があります)", id, len(i.BlockDeviceMappings))
	} else if i.BlockDeviceMappings[0].Ebs == nil {
		c.addFailure("%s のブロックデバイスが EBS ではありません", id)
	}
	if len(i.NetworkInterfaces) != 1 {
		c.addFailure("%s に %d 個のネットワークインターフェイスが検出されました (1個である必要があります)", id, len(i.NetworkInterfaces))
	}
}

func (c *checker) checkVolumes() {
	for _, o := range c.DescribeVolumes {
		for _, v := range o.Volumes {
			c.checkVolume(v)
		}
	}
}

func (c *checker) checkVolume(v *ec2.Volume) {
	id := *v.VolumeId
	if *v.Size != 20 {
		c.addFailure("%s のサイズが %d GB です (20 GB である必要があります)", id, *v.Size)
	}
	if *v.VolumeType != "gp3" {
		c.addFailure("%s のタイプが %s です (gp3 である必要があります)", id, *v.VolumeType)
	}
}

func (c *checker) checkNetworkInterfaces() {
	allowedInstances := make(map[string]struct{})
	for _, o := range c.DescribeInstances {
		for _, r := range o.Reservations {
			for _, i := range r.Instances {
				allowedInstances[*i.InstanceId] = struct{}{}
			}
		}
	}

	isAllowed := func(i *ec2.NetworkInterface) bool {
		if i.Attachment == nil || i.Attachment.InstanceId == nil {
			return false
		}
		_, ok := allowedInstances[*i.Attachment.InstanceId]
		return ok
	}

	for _, o := range c.DescribeNetworkInterfaces {
		for _, i := range o.NetworkInterfaces {
			if !isAllowed(i) {
				c.addFailure("不明なネットワークインターフェイス (%s) が VPC 内に見つかりました", *i.NetworkInterfaceId)
			}
		}
	}
}

func (c *checker) checkSecurityGroups() {
	for _, out := range c.DescribeSecurityGroups {
		for _, sg := range out.SecurityGroups {
			c.checkSecurityGroup(sg)
		}
	}

}

func (c *checker) checkSecurityGroup(sg *ec2.SecurityGroup) {
	id := *sg.GroupId

	var hasIngressSSH, hasIngressHTTPS, hasIngressInternal bool
	for _, p := range sg.IpPermissions {
		if c.isIngressSSH(p) {
			hasIngressSSH = true
			continue
		}
		if c.isIngressHTTPS(p) {
			hasIngressHTTPS = true
			continue
		}
		if c.isIngressInternal(p) {
			hasIngressInternal = true
			continue
		}
		c.addFailure("%s に不明なルールが見つかりました", id)
	}
	if !hasIngressSSH {
		c.addFailure("%s に SSH を許可するルールがありません", id)
	}
	if !hasIngressHTTPS {
		c.addFailure("%s に HTTPS を許可するルールがありません", id)
	}
	if !hasIngressInternal {
		c.addFailure("%s にインスタンス間通信を許可するルールがありません", id)
	}

	if len(sg.IpPermissionsEgress) != 1 {
		c.addFailure("%s のルールが不正です", id)
	}
	for _, p := range sg.IpPermissionsEgress {
		if !c.isEgressAll(p) {
			c.addFailure("%s に不正なルールが見つかりました", id)
		}
	}
}

func (c *checker) isIngressSSH(p *ec2.IpPermission) bool {
	return p.FromPort != nil && *p.FromPort == 22 &&
		p.ToPort != nil && *p.ToPort == 22 &&
		p.IpProtocol != nil && *p.IpProtocol == "tcp" &&
		len(p.IpRanges) == 1 && p.IpRanges[0].CidrIp != nil && *p.IpRanges[0].CidrIp == "0.0.0.0/0" &&
		p.Ipv6Ranges == nil &&
		p.PrefixListIds == nil &&
		p.UserIdGroupPairs == nil
}

func (c *checker) isIngressHTTPS(p *ec2.IpPermission) bool {
	return p.FromPort != nil && *p.FromPort == 443 &&
		p.ToPort != nil && *p.ToPort == 443 &&
		p.IpProtocol != nil && *p.IpProtocol == "tcp" &&
		len(p.IpRanges) == 1 && p.IpRanges[0].CidrIp != nil && *p.IpRanges[0].CidrIp == "0.0.0.0/0" &&
		p.Ipv6Ranges == nil &&
		p.PrefixListIds == nil &&
		p.UserIdGroupPairs == nil
}

func (c *checker) isIngressInternal(p *ec2.IpPermission) bool {
	return p.FromPort == nil &&
		p.ToPort == nil &&
		p.IpProtocol != nil && *p.IpProtocol == "-1" &&
		len(p.IpRanges) == 1 && p.IpRanges[0].CidrIp != nil && *p.IpRanges[0].CidrIp == "192.168.0.0/24" &&
		p.Ipv6Ranges == nil &&
		p.PrefixListIds == nil &&
		p.UserIdGroupPairs == nil
}

func (c *checker) isEgressAll(p *ec2.IpPermission) bool {
	return p.FromPort == nil &&
		p.ToPort == nil &&
		p.IpProtocol != nil && *p.IpProtocol == "-1" &&
		len(p.IpRanges) == 1 && p.IpRanges[0].CidrIp != nil && *p.IpRanges[0].CidrIp == "0.0.0.0/0" &&
		p.Ipv6Ranges == nil &&
		p.PrefixListIds == nil &&
		p.UserIdGroupPairs == nil
}
