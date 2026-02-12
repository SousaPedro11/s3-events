data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "sns_topic" {
  statement {
    sid     = "AllowS3Publish"
    effect  = "Allow"
    actions = ["sns:Publish"]

    principals {
      type        = "Service"
      identifiers = ["s3.amazonaws.com"]
    }

    resources = [aws_sns_topic.this.arn]

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = [aws_s3_bucket.this.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}
