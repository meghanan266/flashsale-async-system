# SNS Topic for order processing events
resource "aws_sns_topic" "order_processing" {
  name = "${local.name_prefix}-order-processing-events"
  
  tags = local.common_tags
}

# SQS Queue for order processing
resource "aws_sqs_queue" "order_processing" {
  name                      = "${local.name_prefix}-order-processing-queue"
  delay_seconds             = 0
  max_message_size          = 262144  # 256 KB
  message_retention_seconds = 345600  # 4 days
  receive_wait_time_seconds = 20      # Long polling
  visibility_timeout_seconds = 30

  tags = local.common_tags
}

# SQS Queue Policy to allow SNS to send messages
resource "aws_sqs_queue_policy" "order_processing" {
  queue_url = aws_sqs_queue.order_processing.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action = "SQS:SendMessage"
        Resource = aws_sqs_queue.order_processing.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.order_processing.arn
          }
        }
      }
    ]
  })
}

# SNS Subscription to SQS
resource "aws_sns_topic_subscription" "order_queue" {
  topic_arn = aws_sns_topic.order_processing.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.order_processing.arn
}
