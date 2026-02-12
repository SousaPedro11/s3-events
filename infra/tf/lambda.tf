resource "null_resource" "build_lambda" {
  triggers = {
    source_hash = local.lambda_source_hash
  }

  provisioner "local-exec" {
    working_dir = path.module
    command     = <<EOT
    set -euxo pipefail
    # source and module
    SRC="${var.lambda_source_dir}"
    MODULE_TF="${path.module}"

    # ensure build dirs exist
    mkdir -p "$SRC/build"
    mkdir -p "$MODULE_TF/build"

    # ensure MODULE_DIR is absolute (handles when ${path.module} is relative like '.')
    MODULE_DIR=$(cd "$MODULE_TF" && pwd)

    # build
    cd "$SRC"
    env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/bootstrap .
    chmod +x build/bootstrap

    # zip bootstrap into the module build folder
    cd build
    zip -r "$MODULE_DIR/build/${local.lambda_name}.zip" bootstrap
    EOT
  }
}


resource "aws_lambda_function" "this" {
  function_name                  = local.lambda_name
  role                           = aws_iam_role.lambda.arn
  handler                        = "bootstrap"
  runtime                        = "provided.al2"
  filename                       = "${path.module}/build/${local.lambda_name}.zip"
  timeout                        = 10
  memory_size                    = 128
  reserved_concurrent_executions = var.lambda_reserved_concurrency
  tags                           = var.tags
  depends_on                     = [null_resource.build_lambda]
}

resource "aws_lambda_permission" "allow_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this.arn
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.this.arn
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${local.lambda_name}"
  retention_in_days = 7
  tags              = var.tags
}
