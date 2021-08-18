### vpc networking

resource "aws_vpc" "isucon11q" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "isucon11q"
  }
}

resource "aws_internet_gateway" "isucon11q" {
  vpc_id = aws_vpc.isucon11q.id
  tags = {
    Name = "isucon11q"
  }
}

resource "aws_route_table" "isucon11q" {
  vpc_id = aws_vpc.isucon11q.id
  route {
    gateway_id = aws_internet_gateway.isucon11q.id
    cidr_block = "0.0.0.0/0"
  }
  tags = {
    Name = "isucon11q"
  }
}

### security group for bench ###

resource "aws_security_group" "isucon11q-bench" {
  name   = "isucon11q-bench"
  vpc_id = aws_vpc.isucon11q.id
  tags = {
    Name = "isucon11q-bench"
  }
}

resource "aws_security_group_rule" "isucon11q-bench-jiaapi" {
  type              = "ingress"
  from_port         = 5000
  to_port           = 5000
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-bench.id
}

resource "aws_security_group_rule" "isucon11q-bench-internal" {
  type              = "ingress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["192.168.0.0/16"]
  security_group_id = aws_security_group.isucon11q-bench.id
}

resource "aws_security_group_rule" "isucon11q-bench-egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-bench.id
}

### security group for jump ###

resource "aws_security_group" "isucon11q-jump" {
  name   = "isucon11q-jump"
  vpc_id = aws_vpc.isucon11q.id
  tags = {
    Name = "isucon11q-jump"
  }
}

resource "aws_security_group_rule" "isucon11q-jump-ssh" {
  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-jump.id
}

resource "aws_security_group_rule" "isucon11q-jump-egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-jump.id
}

### security group for ecs ###

resource "aws_security_group" "isucon11q-ecs" {
  name   = "isucon11q-ecs"
  vpc_id = aws_vpc.isucon11q.id
  tags = {
    Name = "isucon11q-ecs"
  }
}

resource "aws_security_group_rule" "isucon11q-ecs-jiaapi-mock" {
  type              = "ingress"
  from_port         = 5000
  to_port           = 5000
  protocol          = "tcp"
  cidr_blocks       = ["192.168.128.0/25"]
  security_group_id = aws_security_group.isucon11q-ecs.id
}

resource "aws_security_group_rule" "isucon11q-ecs-egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-ecs.id
}

### security group for ecs-lb ###

resource "aws_security_group" "isucon11q-ecs-lb" {
  name   = "isucon11q-ecs-lb"
  vpc_id = aws_vpc.isucon11q.id
  tags = {
    Name = "isucon11q-ecs-lb"
  }
}

resource "aws_security_group_rule" "isucon11q-ecs-lb-https" {
  type      = "ingress"
  from_port = 443
  to_port   = 443
  protocol  = "tcp"
  #cidr_blocks       = ["54.64.248.104/32"]
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-ecs-lb.id
}

resource "aws_security_group_rule" "isucon11q-ecs-lb-egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-ecs-lb.id
}

### vpc peering

data "aws_vpc" "isucon11-portal" {
  filter {
    name   = "tag:Project"
    values = ["portal"]
  }
}
data "aws_route_table" "isucon11-portal" {
  filter {
    name   = "tag:Name"
    values = ["isucon11-private"]
  }
}
data "aws_security_group" "isucon11-portal" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q-ecs"]
  }
}

resource "aws_vpc_peering_connection" "isucon11q-and-isucon11-portal" {
  vpc_id      = aws_vpc.isucon11q.id
  peer_vpc_id = data.aws_vpc.isucon11-portal.id
  auto_accept = true
}

resource "aws_route" "isucon11q-to-isucon11-portal" {
  route_table_id            = aws_route_table.isucon11q.id
  destination_cidr_block    = data.aws_vpc.isucon11-portal.cidr_block
  vpc_peering_connection_id = aws_vpc_peering_connection.isucon11q-and-isucon11-portal.id
  depends_on                = [aws_vpc_peering_connection.isucon11q-and-isucon11-portal]
}

resource "aws_route" "isucon11-portal-to-isucon11q" {
  route_table_id            = data.aws_route_table.isucon11-portal.id
  destination_cidr_block    = aws_vpc.isucon11q.cidr_block
  vpc_peering_connection_id = aws_vpc_peering_connection.isucon11q-and-isucon11-portal.id
  depends_on                = [aws_vpc_peering_connection.isucon11q-and-isucon11-portal]
}

resource "aws_security_group_rule" "isucon11q-to-isucon11-portal-supervisor" {
  type              = "ingress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = [aws_vpc.isucon11q.cidr_block]
  security_group_id = data.aws_security_group.isucon11-portal.id
}

resource "aws_security_group_rule" "isucon11-portal-to-isucon11q-node-exporter" {
  type              = "ingress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = [data.aws_vpc.isucon11-portal.cidr_block]
  security_group_id = aws_security_group.isucon11q-bench.id
}
