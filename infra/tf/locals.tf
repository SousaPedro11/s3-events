locals {
  sns_topic_name = coalesce(var.sns_topic_name, "${var.project_name}-topic")
  lambda_name    = coalesce(var.lambda_name, "${var.project_name}-handler")
  lambda_source_hash = sha256(join("", [
    for f in fileset(var.lambda_source_dir, "**/*.go") : filesha256("${var.lambda_source_dir}/${f}")
  ]))
}
