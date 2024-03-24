locals {
  source_file = "../${path.module}/bootstrap"
  output_path_zip = "../${path.module}/bootstrap.zip"

}

data "archive_file" "lambda_zip" {
  type = "zip"

  source_file  = local.source_file
  output_path = local.output_path_zip
}

resource "aws_s3_object" "lambda_main" {
  bucket = var.bucket_main_id

  key    = "/pigeon/${formatdate("YYYYMMDD", timestamp())}/hash${formatdate("hhmmss", timestamp())}-bootstrap.zip"
  source = local.output_path_zip
  acl   = "private"
  etag = data.archive_file.lambda_zip.output_base64sha256
}