resource "aws_vpc" "isucon11q" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name    = "isucon11q"
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
    Name    = "isucon11q-bench"
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
    Name    = "isucon11q-jump"
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
    Name    = "isucon11q-ecs"
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
    Name    = "isucon11q-ecs-lb"
  }
}

resource "aws_security_group_rule" "isucon11q-ecs-lb-https" {
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
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
