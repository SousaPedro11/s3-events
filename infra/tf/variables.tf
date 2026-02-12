variable "project_name" {
  type        = string
  description = "Base prefix for naming resources."
  default     = "s3-events"
}

variable "region" {
  type        = string
  description = "AWS region for the resources."
  default     = "us-east-1"
}

variable "bucket_name" {
  type        = string
  description = "S3 bucket name (must be globally unique)."
}

variable "sns_topic_name" {
  type        = string
  description = "SNS topic name."
  default     = null
}

variable "lambda_name" {
  type        = string
  description = "Lambda function name."
  default     = null
}

variable "lambda_source_dir" {
  type        = string
  description = "Directory with the Lambda Go code."
  default     = "lambda"
}

variable "lambda_reserved_concurrency" {
  type        = number
  description = "Lambda reserved concurrency limit (null to not set)."
  default     = null
}

variable "tags" {
  type        = map(string)
  description = "Common tags for the resources."
  default     = {}
}
