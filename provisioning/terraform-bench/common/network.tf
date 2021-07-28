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


resource "aws_security_group" "isucon11q" {
  name   = "isucon11q"
  vpc_id = aws_vpc.isucon11q.id
  tags = {
    Name    = "isucon11q"
  }
}

resource "aws_security_group_rule" "isucon11q-ssh" {
  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q.id
}
resource "aws_security_group_rule" "isucon11q-jiaapi" {
  type              = "ingress"
  from_port         = 5000
  to_port           = 5000
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q.id
}

resource "aws_security_group_rule" "isucon11q-egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11q.id
}
