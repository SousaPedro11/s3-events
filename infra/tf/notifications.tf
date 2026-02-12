resource "aws_s3_bucket_notification" "this" {
  bucket = aws_s3_bucket.this.id

  topic {
    topic_arn = aws_sns_topic.this.arn
    events    = ["s3:ObjectCreated:*", "s3:ObjectRemoved:*", "s3:LifecycleExpiration:*"]
  }

  depends_on = [aws_sns_topic_policy.this]
}
