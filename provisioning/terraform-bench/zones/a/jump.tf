### data ###

data "aws_internet_gateway" "isucon11q" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q"]
  }
}

data "aws_security_group" "isucon11q-jump" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q-jump"]
  }
}
#data "aws_security_group" "isucon11q-jia-login" {
#  filter {
#    name   = "tag:Name"
#    values = ["isucon11q-jia-login"]
#  }
#}

data "aws_ami" "jump" {
  # Ubuntu 20.04 latest
  most_recent = true
  owners = ["099720109477"] # Canonical
  filter {
      name   = "name"
      values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
  filter {
      name   = "virtualization-type"
      values = ["hvm"]
  }
}
#data "aws_ami" "jia-login" {
#  owners = ["self"]
#  filter {
#    name = "name"
#    values = ["isucon11q-amd64-jia-login-*"]
#  }
#  filter {
#    name = "tag:GitTag"
#    values = [var.git_tag]
#  }
#}

### resources ###

resource "aws_subnet" "isucon11q-jump" {
  vpc_id                  = data.aws_vpc.isucon11q.id
  cidr_block              = "192.168.0.0/24"
  availability_zone       = "ap-northeast-1a"
  map_public_ip_on_launch = true
  tags = {
    Name = "isucon11q-jump"
  }
}

resource "aws_route_table_association" "isucon11q-jump" {
  subnet_id      = aws_subnet.isucon11q-jump.id
  route_table_id = data.aws_route_table.isucon11q.id
}

resource "aws_eip" "jump" {
  vpc                       = true
  instance                  = aws_instance.jump.id
  associate_with_private_ip = "192.168.0.11"
  tags = {
    Name    = "jump-zone-a"
  }
  #depends_on                = [data.aws_internet_gateway.isucon11q]
}
resource "aws_instance" "jump" {
  ami                    = data.aws_ami.jump.id
  instance_type          = "c5.large"
  key_name               = aws_key_pair.keypair.id
  subnet_id              = aws_subnet.isucon11q-jump.id
  private_ip             = "192.168.0.11"
  vpc_security_group_ids = [data.aws_security_group.isucon11q-jump.id]
  root_block_device {
    volume_size = 20
    volume_type = "gp2"
  }
  tags = {
    Name = "jump-zone-a"
    Kind = "jump"
  }
}

#resource "aws_eip" "jia-login" {
#  vpc                       = true
#  instance                  = aws_instance.jia-login.id
#  associate_with_private_ip = "192.168.0.12"
#  tags = {
#    Name    = "jia-login-zone-a"
#  }
#  #depends_on                = [aws_internet_gateway.isucon11-qualify-dev-20210813]
#}
#resource "aws_instance" "jia-login" {
#  ami                    = data.aws_ami.jia-login.id
#  instance_type          = "c5.large"
#  key_name               = aws_key_pair.keypair.id
#  subnet_id              = aws_subnet.isucon11q-jump.id
#  private_ip             = "192.168.0.12"
#  vpc_security_group_ids = [data.aws_security_group.isucon11q-jia-login.id]
#  root_block_device {
#    volume_size = 20
#    volume_type = "gp2"
#  }
#  tags = {
#    Name = "jia-login-zone-a"
#    Kind = "jia-login"
#  }
#}

### outputs ###

output "jump" {
  value = aws_eip.jump.public_ip
}
#output "jia-login" {
#  value = aws_eip.jia-login.public_ip
#}
