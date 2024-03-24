
variable "aws_region" {
  description = "AWS region for all resources."

  type    = string
  default = "us-east-2"
}

variable "account_id" {
  type    = string
}

variable "access_key" {
  type    = string
}

variable "secret_key" {
  type    = string
}

variable "bucket_main_id" {
  type    = string
}

variable "smtp_host" {
  default = ""
}
variable "smtp_port" {
  default = ""
}
variable "smtp_user" {
  default = ""
}
variable "smtp_pass" {
  default = ""
}
variable "head_metadata" {
  default = ""
}