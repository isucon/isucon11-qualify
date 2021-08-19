### data ###

data "aws_security_group" "isucon11q-jump" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q-jump"]
  }
}

data "aws_ami" "jump" {
  # Ubuntu 20.04 latest
  most_recent = true
  owners      = ["099720109477"] # Canonical
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

### resources ###

resource "aws_subnet" "isucon11q-jump" {
  vpc_id                  = data.aws_vpc.isucon11q.id
  cidr_block              = "192.168.0.0/25"
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
    Name = "jump-zone-a"
  }
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

### outputs ###

output "jump" {
  value = aws_eip.jump.public_ip
}
