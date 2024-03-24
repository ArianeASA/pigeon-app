resource "aws_lambda_function" "pigeon_lambda" {
    function_name = "pigeon_lambda"
    s3_bucket = var.bucket_main_id
    s3_key    = aws_s3_object.lambda_main.key
    s3_object_version = aws_s3_object.lambda_main.version_id

    runtime = "provided.al2"
    handler = "bootstrap"

    source_code_hash = aws_s3_object.lambda_main.content_base64

    environment {
        variables = {
            SMTP_HOST = var.smtp_host
            SMTP_PORT = var.smtp_port
            SMTP_USER = var.smtp_user
            SMTP_PASS = var.smtp_pass
            HEAD_METADATA = var.head_metadata
            REGION_INFRA = var.aws_region
            ROLE_URL = aws_iam_role.presign_role.arn
        }
    }
    role = aws_iam_role.lambda_exec.arn
}

resource "aws_lambda_permission" "allow_bucket" {
    statement_id = "AllowExecutionFromS3Bucket"
    action        = "lambda:InvokeFunction"
    function_name = aws_lambda_function.pigeon_lambda.function_name
    principal     = "s3.amazonaws.com"
    source_arn    =  aws_s3_bucket.pigeon_bucket.arn
}


resource "aws_iam_role" "lambda_exec" {
    name = "lambda_exec_role"

    assume_role_policy = jsonencode({
        Version = "2012-10-17"
        Statement = [
            {
                Action = "sts:AssumeRole"
                Effect = "Allow"
                Principal = {
                    Service = "lambda.amazonaws.com"
                }
            },
        ]
    })
}

resource "aws_iam_role_policy" "lambda_exec" {
    name = "lambda_exec_policy"
    role = aws_iam_role.lambda_exec.id

    policy = jsonencode({
        Version = "2012-10-17"
        Statement = [
            {
                Action = [
                    "logs:CreateLogGroup",
                    "logs:CreateLogStream",
                    "logs:PutLogEvents"
                ]
                Effect   = "Allow"
                Resource = "*"
            },
        ]
    })
}


resource "aws_iam_policy" "s3_access" {
    name        = "s3_access"
    description = "Allows access to the S3 bucket"

    policy = jsonencode({
        Version = "2012-10-17",
        Statement = [
            {
                Effect = "Allow",
                Action = [
                    "s3:GetObject",
                    "s3:ListBucket"
                ],
                Resource = [
                    "arn:aws:s3:::${aws_s3_bucket.pigeon_bucket.id}",
                    "arn:aws:s3:::${aws_s3_bucket.pigeon_bucket.id}/*"
                ]
            }
        ]
    })
}

resource "aws_iam_role_policy_attachment" "lambda_exec_s3_access" {
    role       = aws_iam_role.lambda_exec.name
    policy_arn = aws_iam_policy.s3_access.arn
}