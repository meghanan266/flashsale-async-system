# ECR Repository
resource "aws_ecr_repository" "order_app" {
  name                 = "${local.name_prefix}-order-app"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.common_tags
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "${local.name_prefix}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = local.common_tags
}

# Use the existing LabRole instead of creating new roles
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Security Group for ECS Tasks
resource "aws_security_group" "ecs_tasks" {
  name_prefix = "${local.name_prefix}-ecs-tasks-sg"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.common_tags, {
    Name = "${local.name_prefix}-ecs-tasks-sg"
  })
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "order_receiver" {
  name              = "/ecs/${local.name_prefix}/order-receiver"
  retention_in_days = 7

  tags = local.common_tags
}

# ECS Task Definition for Order Receiver
# ECS Task Definition for Order Receiver
resource "aws_ecs_task_definition" "order_receiver" {
  family                   = "${local.name_prefix}-order-receiver"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = data.aws_iam_role.lab_role.arn
  task_role_arn           = data.aws_iam_role.lab_role.arn

  container_definitions = jsonencode([
    {
      name  = "order-receiver"
      image = "${aws_ecr_repository.order_app.repository_url}:latest"
      
      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "PORT"
          value = "8080"
        },
        {
          name  = "SNS_TOPIC_ARN"
          value = aws_sns_topic.order_processing.arn
        },
        {
          name  = "SQS_QUEUE_URL"
          value = aws_sqs_queue.order_processing.url
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.order_receiver.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }

      essential = true
    }
  ])

  tags = local.common_tags
}

# ECS Service for Order Receiver
resource "aws_ecs_service" "order_receiver" {
  name            = "${local.name_prefix}-order-receiver"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.order_receiver.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.order_receiver.arn
    container_name   = "order-receiver"
    container_port   = 8080
  }

  depends_on = [aws_lb_listener.main]

  tags = local.common_tags
}

# CloudWatch Log Group for Order Processor
resource "aws_cloudwatch_log_group" "order_processor" {
  name              = "/ecs/${local.name_prefix}/order-processor"
  retention_in_days = 7

  tags = local.common_tags
}

# ECR Repository for Order Processor
resource "aws_ecr_repository" "order_processor" {
  name                 = "${local.name_prefix}-order-processor"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = local.common_tags
}

# ECS Task Definition for Order Processor
resource "aws_ecs_task_definition" "order_processor" {
  family                   = "${local.name_prefix}-order-processor"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = data.aws_iam_role.lab_role.arn
  task_role_arn           = data.aws_iam_role.lab_role.arn

  container_definitions = jsonencode([
    {
      name  = "order-processor"
      image = "${aws_ecr_repository.order_processor.repository_url}:latest"
      
      environment = [
        {
          name  = "SQS_QUEUE_URL"
          value = aws_sqs_queue.order_processing.url
        },
        {
          name  = "WORKER_COUNT"
          value = "20"  # Start with 5 workers
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.order_processor.name
          "awslogs-region"        = var.region
          "awslogs-stream-prefix" = "ecs"
        }
      }

      essential = true
    }
  ])

  tags = local.common_tags
}

# ECS Service for Order Processor
resource "aws_ecs_service" "order_processor" {
  name            = "${local.name_prefix}-order-processor"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.order_processor.arn
  desired_count   = 1  # 1 task with 5 worker goroutines
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  tags = local.common_tags
}