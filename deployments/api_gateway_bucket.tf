resource "aws_api_gateway_rest_api" "api" {
    name        = "APIGatewayProxy"
    description = "API Gateway acting as a proxy for S3"
}

resource "aws_api_gateway_resource" "resource" {
    rest_api_id = aws_api_gateway_rest_api.api.id
    parent_id   = aws_api_gateway_rest_api.api.root_resource_id
    path_part   = "{proxy+}"
}

resource "aws_api_gateway_method" "method" {
    rest_api_id   = aws_api_gateway_rest_api.api.id
    resource_id   = aws_api_gateway_resource.resource.id
    http_method   = "GET"
    authorization = "NONE"
}

resource "aws_api_gateway_integration" "integration" {
    rest_api_id = aws_api_gateway_rest_api.api.id
    resource_id = aws_api_gateway_resource.resource.id
    http_method = aws_api_gateway_method.method.http_method

    integration_http_method = "GET"
    type                    = "AWS"
    uri                     = "arn:aws:apigateway:${var.aws_region}:s3:path/${aws_s3_bucket.pigeon_bucket.id}/{proxy}"
}

resource "aws_api_gateway_deployment" "deployment" {
    depends_on  = [aws_api_gateway_integration.integration]
    rest_api_id = aws_api_gateway_rest_api.api.id
    stage_name  = "prod"
}