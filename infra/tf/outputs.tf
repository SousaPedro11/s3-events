output "bucket_name" {
  description = "S3 bucket name."
  value       = aws_s3_bucket.this.bucket
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic."
  value       = aws_sns_topic.this.arn
}

output "lambda_arn" {
  description = "ARN of the Lambda function."
  value       = aws_lambda_function.this.arn
}
