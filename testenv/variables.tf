variable "region" {
  description = "AWS region for the test environment"
  type        = string
  default     = "ap-southeast-2"
}

variable "project_tag" {
  description = "Value of the project tag (used by reachr's scope selector and for teardown)"
  type        = string
  default     = "reachr-testenv"
}

variable "vpc_cidr" {
  description = "CIDR block for the test VPC"
  type        = string
  default     = "10.20.0.0/16"
}

variable "instance_type" {
  description = "App-tier EC2 instance type (t2.micro is free-tier in ap-southeast-2)"
  type        = string
  default     = "t2.micro"
}

variable "app_port" {
  description = "Port the app tier listens on behind the ALB"
  type        = number
  default     = 8080
}
