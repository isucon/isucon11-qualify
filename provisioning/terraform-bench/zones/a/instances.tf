### variables & locals ###

variable "ami_id" {
  type = string
}
variable "isuxportal_supervisor_endpoint_url" {
  type = string
}
variable "isuxportal_supervisor_token" {
  type = string
}

locals {
  team_ids = jsondecode(file("./teams.json"))
}

### data ###

data "aws_vpc" "isucon11q" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q"]
  }
}

data "aws_route_table" "isucon11q" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q"]
  }
}

data "aws_internet_gateway" "isucon11q" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q"]
  }
}

data "aws_security_group" "isucon11q" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q"]
  }
}

# TODO 効かない
data "template_file" "isuxportal_supervisor_env" {
  for_each = toset(local.team_ids)

  template = file("../../base/generate-isuxportal-supervisor-env.sh")
  vars = {
    isuxportal_supervisor_endpoint_url = var.isuxportal_supervisor_endpoint_url
    isuxportal_supervisor_token        = var.isuxportal_supervisor_token
    isuxportal_supervisor_team_id      = each.value
  }
}

data "template_cloudinit_config" "config" {
  for_each = toset(local.team_ids)

  gzip          = true
  base64_encode = true
  part {
    content_type = "text/x-shellscript"
    content      = data.template_file.isuxportal_supervisor_env[each.value].rendered
  }
}

### resources ###

resource "aws_subnet" "isucon11q-zone-a" {
  vpc_id                  = data.aws_vpc.isucon11q.id
  cidr_block              = "192.168.0.0/20"
  availability_zone       = "ap-northeast-1a"
  map_public_ip_on_launch = true
  tags = {
    Name = "isucon11q-zone-a"
  }
}

resource "aws_route_table_association" "isucon11q" {
  subnet_id      = aws_subnet.isucon11q-zone-a.id
  route_table_id = data.aws_route_table.isucon11q.id
}

resource "aws_eip" "bench" {
  for_each = toset(local.team_ids)

  vpc                       = true
  instance                  = aws_instance.bench[each.value].id
  associate_with_private_ip = "192.168.1.${index(local.team_ids, each.value) + 1}"
  depends_on                = [data.aws_internet_gateway.isucon11q]
}

resource "aws_key_pair" "bench" {
  key_name   = "isucon11q-zone-a"
  public_key = file("../../base/pubkey.pem")
}

resource "aws_instance" "bench" {
  for_each = toset(local.team_ids)

  ami                    = var.ami_id
  instance_type          = "c5.large"
  key_name               = aws_key_pair.bench.id
  subnet_id              = aws_subnet.isucon11q-zone-a.id
  private_ip             = "192.168.1.${index(local.team_ids, each.value) + 1}"
  vpc_security_group_ids = [data.aws_security_group.isucon11q.id]
  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    throughput  = 125
  }
  user_data = data.template_cloudinit_config.config[each.value].rendered
  tags = {
    Name = format("bench-%s", each.value)
  }
}

