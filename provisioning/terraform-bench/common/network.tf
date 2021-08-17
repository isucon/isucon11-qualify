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

### security group for jia-login ###

resource "aws_security_group" "isucon11q-jia-login" {
  name   = "isucon11q-jia-login"
  vpc_id = aws_vpc.isucon11q.id
  tags = {
    Name    = "isucon11q-jia-login"
  }
}

resource "aws_security_group_rule" "isucon11q-jia-login-http" {
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q-jia-login.id
}
