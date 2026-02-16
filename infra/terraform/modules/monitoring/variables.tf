variable "project_name" { type = string }
variable "environment" { type = string }
variable "aws_region" { type = string }
variable "random_suffix" { type = string }

variable "ecs_cluster_name" { type = string }
variable "backend_service_name" { type = string }
variable "frontend_service_name" { type = string }
variable "alb_arn_suffix" { type = string }
variable "rds_instance_id" { type = string }
variable "redis_cluster_id" { type = string }

variable "alert_email" { type = string }
variable "cpu_alarm_threshold" { type = number }
variable "memory_alarm_threshold" { type = number }
variable "error_rate_threshold" { type = number }
