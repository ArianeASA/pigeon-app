resource "aws_s3_bucket" "pigeon-bucket" {
    bucket = "pigeon-bucket"
}

resource "aws_s3_bucket_ownership_controls" "lambda_bucket_controls" {
    bucket = aws_s3_bucket.pigeon-bucket.id
    rule {
        object_ownership = "BucketOwnerPreferred"
    }
}

resource "aws_s3_bucket_acl" "lambda_bucket_acl" {
    depends_on = [aws_s3_bucket_ownership_controls.lambda_bucket_controls]

    bucket = aws_s3_bucket.pigeon-bucket.id
    acl    = "private"
}
resource "aws_s3_bucket_notification" "bucket_notification" {
    bucket = aws_s3_bucket.pigeon-bucket.id

    lambda_function {
        lambda_function_arn = aws_lambda_function.pigeon_lambda.arn
        events              = ["s3:ObjectCreated:*"]
        filter_prefix       = "test/"
    }
}