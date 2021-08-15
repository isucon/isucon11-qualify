### variables & locals ###

variable "defaultGitTag" {
  type    = string
  default = "08.05.0"
}

locals {
  teams = {
    "qualify-dev" : {
      gitTag   = var.defaultGitTag
      baseAddr = 10 # 11, 12, 13
    }
    "sorah" : {
      gitTag   = var.defaultGitTag
      baseAddr = 20
    }
    "toki" : {
      gitTag   = var.defaultGitTag
      baseAddr = 30
    }
    #"temma" : {
    #  gitTag   = var.defaultGitTag
    #  baseAddr = 40
    #}
    "takonomura" : {
      gitTag   = var.defaultGitTag
      baseAddr = 50
    }
    #"eagletmt" : {
    #  gitTag   = var.defaultGitTag
    #  baseAddr = 60
    #}
    #"hosshii" : {
    #  gitTag   = var.defaultGitTag
    #  baseAddr = 70
    #}
    #"sapphi_red" : {
    #  gitTag   = var.defaultGitTag
    #  baseAddr = 80
    #}
    #"karino" : {
    #  gitTag   = var.defaultGitTag
    #  baseAddr = 90
    #}
  }
}

### data ###

data "aws_ami" "bench" {
  owners = ["self"]
  filter {
    name   = "name"
    values = ["isucon11q-amd64-bench-*"]
  }
  filter {
    name   = "tag:GitTag"
    values = [var.defaultGitTag]
  }
}
data "aws_ami" "monitoring" {
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
data "aws_ami" "contestant" {
  for_each = local.teams

  owners = ["self"]
  filter {
    name   = "name"
    values = ["isucon11q-amd64-contestant-*"]
  }
  filter {
    name   = "tag:GitTag"
    values = [each.value.gitTag]
  }
}
data "template_file" "set-pubkey" {
  template = file("./set-pubkey-for-dev.sh")
}
data "template_cloudinit_config" "config" {
  gzip          = true
  base64_encode = true
  part {
    content_type = "text/x-shellscript"
    content      = data.template_file.set-pubkey.rendered
  }
}
data "template_file" "set-pubkey-for-monitoring" {
  template = file("./set-pubkey-for-monitoring.sh")
}
data "template_cloudinit_config" "config-for-monitoring" {
  gzip          = true
  base64_encode = true
  part {
    content_type = "text/x-shellscript"
    content      = data.template_file.set-pubkey-for-monitoring.rendered
  }
}


### resources ###

resource "aws_vpc" "isucon11-qualify-dev" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name    = "isucon11-qualify-dev"
  }
}

resource "aws_subnet" "isucon11-qualify-dev" {
  vpc_id                  = aws_vpc.isucon11-qualify-dev.id
  cidr_block              = "192.168.0.0/24"
  availability_zone       = "ap-northeast-1a"
  map_public_ip_on_launch = true
  tags = {
    Name    = "isucon11-qualify-dev"
  }
}

resource "aws_internet_gateway" "isucon11-qualify-dev" {
  vpc_id = aws_vpc.isucon11-qualify-dev.id
  tags = {
    Name    = "isucon11-qualify-dev"
  }
}

resource "aws_route_table" "isucon11-qualify-dev" {
  vpc_id = aws_vpc.isucon11-qualify-dev.id
  route {
    gateway_id = aws_internet_gateway.isucon11-qualify-dev.id
    cidr_block = "0.0.0.0/0"
  }
  tags = {
    Name    = "isucon11-qualify-dev"
  }
}

resource "aws_route_table_association" "isucon11-qualify-dev" {
  subnet_id      = aws_subnet.isucon11-qualify-dev.id
  route_table_id = aws_route_table.isucon11-qualify-dev.id
}

resource "aws_security_group" "isucon11-qualify-dev" {
  name   = "isucon11-qualify-dev"
  vpc_id = aws_vpc.isucon11-qualify-dev.id
  tags = {
    Name    = "isucon11-qualify-dev"
  }
}

resource "aws_security_group_rule" "isucon11-qualify-dev-ssh" {
  type              = "ingress"
  from_port         = 22
  to_port           = 22
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11-qualify-dev.id
}

resource "aws_security_group_rule" "isucon11-qualify-dev-internal" {
  type              = "ingress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["192.168.0.0/24"]
  security_group_id = aws_security_group.isucon11-qualify-dev.id
}

resource "aws_security_group_rule" "isucon11-qualify-dev-egress" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.isucon11-qualify-dev.id
}

### Bench VM (2 台: 競技者がポータルから実行する用と予選チーム用)

resource "aws_eip" "bench" {
  vpc                       = true
  instance                  = aws_instance.bench.id
  associate_with_private_ip = "192.168.0.10"
  depends_on                = [aws_internet_gateway.isucon11-qualify-dev]
}
resource "aws_instance" "bench" {
  ami                    = data.aws_ami.bench.id
  instance_type          = "c5.large"
  subnet_id              = aws_subnet.isucon11-qualify-dev.id
  private_ip             = "192.168.0.10"
  vpc_security_group_ids = [aws_security_group.isucon11-qualify-dev.id]
  root_block_device {
    volume_size = 20
  }
  user_data = data.template_cloudinit_config.config.rendered
  tags = {
    Name    = "isucon11-qualify-dev-bench"
  }
}
output "bench_public_ip" {
  description = "bench"
  value       = aws_eip.bench.public_ip
}

resource "aws_instance" "bench-dev" {
  ami                    = data.aws_ami.bench.id
  instance_type          = "c5.large"
  subnet_id              = aws_subnet.isucon11-qualify-dev.id
  private_ip             = "192.168.0.8"
  vpc_security_group_ids = [aws_security_group.isucon11-qualify-dev.id]
  root_block_device {
    volume_size = 20
  }
  user_data = data.template_cloudinit_config.config.rendered
  tags = {
    Name    = "isucon11-qualify-dev-bench-dev"
  }
}
output "bench-dev_public_ip" {
  description = "monitoring"
  value       = aws_instance.bench-dev.public_ip
}


### Monitoring VM (1 台)
resource "aws_instance" "monitoring" {
  ami                    = data.aws_ami.monitoring.id
  instance_type          = "c5.large"
  subnet_id              = aws_subnet.isucon11-qualify-dev.id
  private_ip             = "192.168.0.9"
  vpc_security_group_ids = [aws_security_group.isucon11-qualify-dev.id]
  root_block_device {
    volume_size = 20
  }
  user_data = data.template_cloudinit_config.config-for-monitoring.rendered
  tags = {
    Name    = "isucon11-qualify-dev-monitoring"
  }
}
output "monitoring_public_ip" {
  description = "monitoring"
  value       = aws_instance.monitoring.public_ip
}


### Contestant VM (3 * N 台)

resource "aws_eip" "contestant-01" {
  for_each = local.teams

  vpc                       = true
  instance                  = aws_instance.contestant-01[each.key].id
  associate_with_private_ip = "192.168.0.${each.value.baseAddr + 1}"
  tags = {
    Name    = "isucon11-qualify-dev-${each.key}-01"
    TeamID  = each.key
  }
  depends_on                = [aws_internet_gateway.isucon11-qualify-dev]
}
resource "aws_instance" "contestant-01" {
  for_each = local.teams

  ami                    = data.aws_ami.contestant[each.key].id
  instance_type          = "c5.large"
  subnet_id              = aws_subnet.isucon11-qualify-dev.id
  private_ip             = "192.168.0.${each.value.baseAddr + 1}"
  vpc_security_group_ids = [aws_security_group.isucon11-qualify-dev.id]
  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    throughput  = 125
  }
  user_data = data.template_cloudinit_config.config.rendered
  tags = {
    Name    = "isucon11-qualify-dev-${each.key}-01"
    TeamID  = each.key
  }
}
output "contestant-01_public_ip" {
  description = "isucon11-qualify-dev"
  value = {
    for eip in aws_eip.contestant-01 :
    eip.tags.Name => eip.public_ip
  }
}

resource "aws_eip" "contestant-02" {
  for_each = local.teams

  vpc                       = true
  instance                  = aws_instance.contestant-02[each.key].id
  associate_with_private_ip = "192.168.0.${each.value.baseAddr + 2}"
  tags = {
    Name    = "isucon11-qualify-dev-${each.key}-02"
    TeamID  = each.key
  }
  depends_on                = [aws_internet_gateway.isucon11-qualify-dev]
}
resource "aws_instance" "contestant-02" {
  for_each = local.teams

  ami                    = data.aws_ami.contestant[each.key].id
  instance_type          = "c5.large"
  subnet_id              = aws_subnet.isucon11-qualify-dev.id
  private_ip             = "192.168.0.${each.value.baseAddr + 2}"
  vpc_security_group_ids = [aws_security_group.isucon11-qualify-dev.id]
  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    throughput  = 125
  }
  user_data = data.template_cloudinit_config.config.rendered
  tags = {
    Name    = "isucon11-qualify-dev-${each.key}-02"
    TeamID  = each.key
  }
}
output "contestant-02_public_ip" {
  description = "isucon11-qualify-dev"
  value = {
    for eip in aws_eip.contestant-02 :
    eip.tags.Name => eip.public_ip
  }
}

resource "aws_eip" "contestant-03" {
  for_each = local.teams

  vpc                       = true
  instance                  = aws_instance.contestant-03[each.key].id
  associate_with_private_ip = "192.168.0.${each.value.baseAddr + 3}"
  tags = {
    Name    = "isucon11-qualify-dev-${each.key}-03"
    TeamID  = each.key
  }
  depends_on                = [aws_internet_gateway.isucon11-qualify-dev]
}
resource "aws_instance" "contestant-03" {
  for_each = local.teams

  ami                    = data.aws_ami.contestant[each.key].id
  instance_type          = "c5.large"
  subnet_id              = aws_subnet.isucon11-qualify-dev.id
  private_ip             = "192.168.0.${each.value.baseAddr + 3}"
  vpc_security_group_ids = [aws_security_group.isucon11-qualify-dev.id]
  root_block_device {
    volume_size = 20
    volume_type = "gp3"
    throughput  = 125
  }
  user_data = data.template_cloudinit_config.config.rendered
  tags = {
    Name    = "isucon11-qualify-dev-${each.key}-03"
    TeamID  = each.key
  }
}
output "contestant-03_public_ip" {
  description = "isucon11-qualify-dev"
  value = {
    for eip in aws_eip.contestant-03 :
    eip.tags.Name => eip.public_ip
  }
}
