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

data "aws_security_group" "isucon11q-ecs" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q-ecs"]
  }
}

data "aws_security_group" "isucon11q-ecs-lb" {
  filter {
    name   = "tag:Name"
    values = ["isucon11q-ecs-lb"]
  }
}

data "aws_ecs_cluster" "isucon11q-ecs" {
  cluster_name = "isuxportal-fargate"
}

data "aws_iam_role" "ecs" {
  name = "EcsTaskExecution"
}

data "aws_acm_certificate" "xi_isucon_dev" {
  domain      = "*.xi.isucon.dev"
  types       = ["AMAZON_ISSUED"]
  most_recent = true
}

### resources ###

resource "aws_subnet" "isucon11q-ecs-zone-a" {
  vpc_id                  = data.aws_vpc.isucon11q.id
  cidr_block              = "192.168.128.0/26"
  availability_zone       = "ap-northeast-1a"
  map_public_ip_on_launch = true
  tags = {
    Name = "isucon11q-ecs-a"
  }
}
resource "aws_subnet" "isucon11q-ecs-zone-c" {
  vpc_id                  = data.aws_vpc.isucon11q.id
  cidr_block              = "192.168.192.0/26"
  availability_zone       = "ap-northeast-1c"
  map_public_ip_on_launch = true
  tags = {
    Name = "isucon11q-ecs-c"
  }
}

resource "aws_route_table_association" "isucon11q-ecs-zone-a" {
  subnet_id      = aws_subnet.isucon11q-ecs-zone-a.id
  route_table_id = data.aws_route_table.isucon11q.id
}
resource "aws_route_table_association" "isucon11q-ecs-zone-c" {
  subnet_id      = aws_subnet.isucon11q-ecs-zone-c.id
  route_table_id = data.aws_route_table.isucon11q.id
}

### for ALB

resource "aws_lb" "isucon11q-ecs" {
  name               = "isucon11q-ecs"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [data.aws_security_group.isucon11q-ecs-lb.id]
  subnets            = [aws_subnet.isucon11q-ecs-zone-a.id, aws_subnet.isucon11q-ecs-zone-c.id]

  enable_deletion_protection = false
}

resource "aws_lb_listener" "isucon11q-ecs" {
  load_balancer_arn = aws_lb.isucon11q-ecs.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = data.aws_acm_certificate.xi_isucon_dev.arn

  default_action {
    type = "fixed-response"

    fixed_response {
      content_type = "text/plain"
      message_body = "not found"
      status_code  = "404"
    }
  }
}

resource "aws_lb_listener_rule" "isucon11q-ecs-jiaapi-mock" {
  listener_arn = aws_lb_listener.isucon11q-ecs.arn
  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.isucon11q-ecs-jiaapi-mock.arn
  }
  condition {
    path_pattern {
      values = ["/", "/api/auth"]
    }
  }
}

resource "aws_lb_target_group" "isucon11q-ecs-jiaapi-mock" {
  name   = "isucon11q-jiaapi-mock"
  vpc_id = data.aws_vpc.isucon11q.id

  port        = 5000
  protocol    = "HTTP"
  target_type = "ip"

  health_check {
    port = 5000
    path = "/"
  }
}

### for ECS

resource "aws_ecs_task_definition" "isucon11q-ecs-jiaapi-mock" {
  family                   = "isucon11q-jiaapi-mock"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  network_mode       = "awsvpc"
  task_role_arn      = data.aws_iam_role.ecs.arn
  execution_role_arn = data.aws_iam_role.ecs.arn

  container_definitions = jsonencode([
    {
      name      = "jiaapi-mock"
      image     = "245943874622.dkr.ecr.ap-northeast-1.amazonaws.com/jiaapi-mock:08.18.1"
      essential = true
      portMappings = [
        {
          containerPort = 5000
          hostPort      = 5000
        }
      ]
    }
  ])
}

resource "aws_ecs_service" "isucon11q-ecs-jiaapi-mock" {
  name = "isucon11q-jiaapi-mock"

  cluster         = data.aws_ecs_cluster.isucon11q-ecs.id
  launch_type     = "FARGATE"
  desired_count   = "2"
  task_definition = aws_ecs_task_definition.isucon11q-ecs-jiaapi-mock.arn

  network_configuration {
    subnets          = [aws_subnet.isucon11q-ecs-zone-a.id, aws_subnet.isucon11q-ecs-zone-c.id]
    security_groups  = [data.aws_security_group.isucon11q-ecs.id]
    assign_public_ip = true
  }
  load_balancer {
    target_group_arn = aws_lb_target_group.isucon11q-ecs-jiaapi-mock.arn
    container_name   = "jiaapi-mock"
    container_port   = 5000
  }
  depends_on = [aws_lb_listener_rule.isucon11q-ecs-jiaapi-mock]
}

