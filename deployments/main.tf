
provider "aws" {
  region = var.aws_region

  secret_key                 = var.secret_key
  access_key                 = var.access_key
  skip_credentials_validation = true
  skip_requesting_account_id = true
  skip_metadata_api_check    = true
}