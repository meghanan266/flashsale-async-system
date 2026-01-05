# Lambda Function
resource "aws_lambda_function" "order_processor" {
  filename         = "../lambda/function.zip"
  function_name    = "${local.name_prefix}-order-processor-lambda"
  role            = data.aws_iam_role.lab_role.arn
  handler         = "bootstrap"
  runtime         = "provided.al2"
  memory_size     = 512
  timeout         = 30

  environment {
    variables = {
      ENVIRONMENT = var.environment
    }
  }

  tags = local.common_tags
}

# SNS subscription to Lambda
resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.order_processing.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.order_processor.arn
}

# Permission for SNS to invoke Lambda
resource "aws_lambda_permission" "sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.order_processor.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.order_processing.arn
}

# CloudWatch Log Group for Lambda
resource "aws_cloudwatch_log_group" "lambda_processor" {
  name              = "/aws/lambda/${aws_lambda_function.order_processor.function_name}"
  retention_in_days = 7
  
  tags = local.common_tags
}