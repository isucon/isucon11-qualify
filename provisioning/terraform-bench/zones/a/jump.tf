
data "aws_security_group" "isucon11q-jump" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q-jump"]
  }
}

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

resource "aws_instance" "jump" {
  ami                    = data.aws_ami.jump.id
  instance_type          = "c5.large"
  key_name               = aws_key_pair.keypair.id
  subnet_id              = aws_subnet.isucon11q-zone-a.id
  private_ip             = "192.168.0.11"
  vpc_security_group_ids = [data.aws_security_group.isucon11q-jump.id]
  root_block_device {
    volume_size = 20
    volume_type = "gp2"
  }
  tags = {
    Name = "jump-zone-a"
  }
}
