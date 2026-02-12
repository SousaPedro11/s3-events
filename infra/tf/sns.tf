resource "aws_sns_topic" "this" {
  name = local.sns_topic_name
  tags = var.tags
}

resource "aws_sns_topic_policy" "this" {
  arn    = aws_sns_topic.this.arn
  policy = data.aws_iam_policy_document.sns_topic.json
}

resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.this.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.this.arn

  depends_on = [aws_lambda_permission.allow_sns]
}
