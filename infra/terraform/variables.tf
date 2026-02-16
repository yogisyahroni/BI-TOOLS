# ============================================================================
# InsightEngine AI — Root Variables
# ============================================================================

# ===========================================================================
# General
# ===========================================================================
variable "project_name" {
  description = "Project name used for resource naming and tagging"
  type        = string
  default     = "insight-engine"
}

variable "environment" {
  description = "Deployment environment (production, staging, development)"
  type        = string
  validation {
    condition     = contains(["production", "staging", "development"], var.environment)
    error_message = "Environment must be one of: production, staging, development."
  }
}

variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "ap-southeast-1"
}

# ===========================================================================
# Networking
# ===========================================================================
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones to deploy into"
  type        = list(string)
  default     = ["ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"]
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets (ALB, NAT Gateway)"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets (ECS, RDS, Redis)"
  type        = list(string)
  default     = ["10.0.11.0/24", "10.0.12.0/24", "10.0.13.0/24"]
}

# ===========================================================================
# Database — PostgreSQL
# ===========================================================================
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "db_allocated_storage" {
  description = "Allocated storage in GB for the RDS instance"
  type        = number
  default     = 50
}

variable "db_name" {
  description = "Name of the PostgreSQL database"
  type        = string
  default     = "insightengine"
}

variable "db_username" {
  description = "Master username for the RDS instance"
  type        = string
  default     = "insightadmin"
  sensitive   = true
}

variable "db_multi_az" {
  description = "Enable Multi-AZ deployment for RDS"
  type        = bool
  default     = false
}

# ===========================================================================
# Database — Redis
# ===========================================================================
variable "redis_node_type" {
  description = "ElastiCache Redis node type"
  type        = string
  default     = "cache.t3.micro"
}

variable "redis_num_cache_nodes" {
  description = "Number of Redis cache nodes"
  type        = number
  default     = 1
}

# ===========================================================================
# Compute — ECS Fargate
# ===========================================================================
variable "backend_image" {
  description = "Docker image URI for the backend service"
  type        = string
}

variable "frontend_image" {
  description = "Docker image URI for the frontend service"
  type        = string
}

variable "backend_cpu" {
  description = "CPU units for the backend task (1024 = 1 vCPU)"
  type        = number
  default     = 512
}

variable "backend_memory" {
  description = "Memory in MB for the backend task"
  type        = number
  default     = 1024
}

variable "frontend_cpu" {
  description = "CPU units for the frontend task (1024 = 1 vCPU)"
  type        = number
  default     = 256
}

variable "frontend_memory" {
  description = "Memory in MB for the frontend task"
  type        = number
  default     = 512
}

variable "backend_desired_count" {
  description = "Desired number of backend service tasks"
  type        = number
  default     = 2
}

variable "frontend_desired_count" {
  description = "Desired number of frontend service tasks"
  type        = number
  default     = 2
}

variable "backend_max_count" {
  description = "Maximum number of backend service tasks for auto-scaling"
  type        = number
  default     = 6
}

variable "frontend_max_count" {
  description = "Maximum number of frontend service tasks for auto-scaling"
  type        = number
  default     = 4
}

variable "jwt_secret" {
  description = "JWT signing secret (injected as environment variable)"
  type        = string
  sensitive   = true
}

variable "acm_certificate_arn" {
  description = "ARN of ACM certificate for HTTPS on ALB"
  type        = string
  default     = ""
}

# ===========================================================================
# Monitoring
# ===========================================================================
variable "alert_email" {
  description = "Email address for CloudWatch alarm notifications"
  type        = string
}

variable "cpu_alarm_threshold" {
  description = "CPU utilization percentage threshold for alarms"
  type        = number
  default     = 80
}

variable "memory_alarm_threshold" {
  description = "Memory utilization percentage threshold for alarms"
  type        = number
  default     = 85
}

variable "error_rate_threshold" {
  description = "ALB 5xx error count threshold for alarms"
  type        = number
  default     = 50
}
