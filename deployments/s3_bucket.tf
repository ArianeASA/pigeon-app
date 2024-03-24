resource "aws_s3_bucket" "pigeon_bucket" {
    bucket = "pigeon-bucket-hacka-app"
}

resource "aws_s3_bucket_ownership_controls" "lambda_bucket_controls" {
    bucket = aws_s3_bucket.pigeon_bucket.id
    rule {
        object_ownership = "BucketOwnerPreferred"
    }
}

resource "aws_s3_object" "lambda_object_filter" {
    bucket = aws_s3_bucket.pigeon_bucket.id
    key    = "test/"
    acl   = "private"
}

resource "aws_iam_policy" "s3_get_object_pigeon" {
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
                    "arn:aws:s3:::${aws_s3_bucket.pigeon_bucket.id}/${aws_s3_object.lambda_object_filter.key}"
                ]
            }
        ]
    })
}

resource "aws_iam_role_policy_attachment" "lambda_object_filter_s3_get_object" {
    role       = aws_iam_role.lambda_exec.name
    policy_arn = aws_iam_policy.s3_get_object_pigeon.arn
}

resource "aws_s3_bucket_acl" "lambda_bucket_acl" {
    depends_on = [aws_s3_bucket_ownership_controls.lambda_bucket_controls]

    bucket = aws_s3_bucket.pigeon_bucket.id
    acl    = "private"
}
resource "aws_s3_bucket_notification" "bucket_notification" {
    bucket = aws_s3_bucket.pigeon_bucket.id

    lambda_function {
        lambda_function_arn = aws_lambda_function.pigeon_lambda.arn
        events              = ["s3:ObjectCreated:*"]
        filter_prefix       = "test/"
    }
}