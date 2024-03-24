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

resource "aws_iam_policy" "s3_get_object" {
  name        = "s3_get_object"
  description = "Allows access to the S3 object"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "s3:GetObject"
        ],
        Resource = [
          "arn:aws:s3:::${var.bucket_main_id}/${aws_s3_object.lambda_main.key}"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_exec_s3_get_object" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = aws_iam_policy.s3_get_object.arn
}