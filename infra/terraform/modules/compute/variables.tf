variable "project_name" { type = string }
variable "environment" { type = string }
variable "aws_region" { type = string }
variable "vpc_id" { type = string }
variable "public_subnet_ids" { type = list(string) }
variable "private_subnet_ids" { type = list(string) }
variable "alb_security_group_id" { type = string }
variable "ecs_security_group_id" { type = string }
variable "random_suffix" { type = string }

variable "backend_image" { type = string }
variable "frontend_image" { type = string }

variable "backend_cpu" { type = number }
variable "backend_memory" { type = number }
variable "frontend_cpu" { type = number }
variable "frontend_memory" { type = number }

variable "backend_desired_count" { type = number }
variable "frontend_desired_count" { type = number }
variable "backend_max_count" { type = number }
variable "frontend_max_count" { type = number }

variable "database_url" {
  type      = string
  sensitive = true
}

variable "redis_url" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}

variable "acm_certificate_arn" { type = string }
