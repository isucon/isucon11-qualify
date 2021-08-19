### variables & locals ###

variable "git_tag" {
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

data "aws_security_group" "isucon11q-bench" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q-bench"]
  }
}

data "aws_ami" "bench" {
  owners = ["self"]
  filter {
    name = "name"
    values = ["isucon11q-amd64-bench-*"]
  }
  filter {
    name = "tag:GitTag"
    values = [var.git_tag]
  }
}

data "template_file" "isuxportal_supervisor_env" {
  for_each = toset(local.team_ids)

  template = file("../../.base/generate-isuxportal-supervisor-env.sh")
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
    content      = data.template_file.isuxportal_supervisor_env[each.key].rendered
  }
}

### resources ###

resource "aws_subnet" "isucon11q-zone-c" {
  vpc_id                  = data.aws_vpc.isucon11q.id
  cidr_block              = "192.168.2.0/24"
  availability_zone       = "ap-northeast-1c"
  map_public_ip_on_launch = true
  tags = {
    Name = "isucon11q-zone-c"
  }
}

resource "aws_route_table_association" "isucon11q" {
  subnet_id      = aws_subnet.isucon11q-zone-c.id
  route_table_id = data.aws_route_table.isucon11q.id
}

resource "aws_key_pair" "keypair" {
  key_name   = "isucon11q-zone-c"
  public_key = file("../../.base/pubkey.pem")
}

resource "aws_instance" "bench" {
  for_each = toset(local.team_ids)

  ami                    = data.aws_ami.bench.id
  instance_type          = "c5.large"
  key_name               = aws_key_pair.keypair.id
  subnet_id              = aws_subnet.isucon11q-zone-c.id
  private_ip             = "192.168.2.${index(local.team_ids, each.key) + 4}"
  vpc_security_group_ids = [data.aws_security_group.isucon11q-bench.id]
  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    throughput  = 125
  }
  user_data = data.template_cloudinit_config.config[each.key].rendered
  tags = {
    Name = format("bench-%s", each.value)
    Kind = "bench"
  }
}
