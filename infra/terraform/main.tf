# ============================================================================
# InsightEngine AI — Root Terraform Configuration
# ============================================================================
# This is the root module that orchestrates all infrastructure modules.
# Each module is independently versioned and can be deployed separately.
#
# Usage:
#   terraform init
#   terraform plan -var-file=environments/production/terraform.tfvars
#   terraform apply -var-file=environments/production/terraform.tfvars
# ============================================================================

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  backend "s3" {
    # Configure via -backend-config or environment-specific .tfbackend files
    # bucket         = "insight-engine-terraform-state"
    # key            = "infrastructure/terraform.tfstate"
    # region         = "ap-southeast-1"
    # dynamodb_table = "terraform-locks"
    # encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "InsightEngine"
      Environment = var.environment
      ManagedBy   = "Terraform"
      Team        = "Platform"
    }
  }
}

# ===========================================================================
# Random suffix for globally unique resource names
# ===========================================================================
resource "random_id" "suffix" {
  byte_length = 4
}

# ===========================================================================
# Module: Networking (VPC, Subnets, Security Groups)
# ===========================================================================
module "networking" {
  source = "./modules/networking"

  project_name        = var.project_name
  environment         = var.environment
  vpc_cidr            = var.vpc_cidr
  availability_zones  = var.availability_zones
  public_subnet_cidrs = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  random_suffix       = random_id.suffix.hex
}

# ===========================================================================
# Module: Database (RDS PostgreSQL + ElastiCache Redis)
# ===========================================================================
module "database" {
  source = "./modules/database"

  project_name         = var.project_name
  environment          = var.environment
  vpc_id               = module.networking.vpc_id
  private_subnet_ids   = module.networking.private_subnet_ids
  db_security_group_id = module.networking.db_security_group_id
  random_suffix        = random_id.suffix.hex

  # PostgreSQL
  db_instance_class    = var.db_instance_class
  db_allocated_storage = var.db_allocated_storage
  db_name              = var.db_name
  db_username          = var.db_username
  db_multi_az          = var.db_multi_az

  # Redis
  redis_node_type      = var.redis_node_type
  redis_num_cache_nodes = var.redis_num_cache_nodes
  redis_security_group_id = module.networking.redis_security_group_id
}

# ===========================================================================
# Module: Compute (ECS Fargate — Backend + Frontend)
# ===========================================================================
module "compute" {
  source = "./modules/compute"

  project_name       = var.project_name
  environment        = var.environment
  aws_region         = var.aws_region
  vpc_id             = module.networking.vpc_id
  public_subnet_ids  = module.networking.public_subnet_ids
  private_subnet_ids = module.networking.private_subnet_ids
  alb_security_group_id = module.networking.alb_security_group_id
  ecs_security_group_id = module.networking.ecs_security_group_id
  random_suffix      = random_id.suffix.hex

  # Container images
  backend_image  = var.backend_image
  frontend_image = var.frontend_image

  # Task sizing
  backend_cpu    = var.backend_cpu
  backend_memory = var.backend_memory
  frontend_cpu   = var.frontend_cpu
  frontend_memory = var.frontend_memory

  # Scaling
  backend_desired_count  = var.backend_desired_count
  frontend_desired_count = var.frontend_desired_count
  backend_max_count      = var.backend_max_count
  frontend_max_count     = var.frontend_max_count

  # Environment injection
  database_url = module.database.connection_string
  redis_url    = module.database.redis_connection_string
  jwt_secret   = var.jwt_secret

  # ACM certificate for HTTPS
  acm_certificate_arn = var.acm_certificate_arn
}

# ===========================================================================
# Module: Monitoring (CloudWatch Alarms, Dashboards, SNS)
# ===========================================================================
module "monitoring" {
  source = "./modules/monitoring"

  project_name       = var.project_name
  environment        = var.environment
  aws_region         = var.aws_region
  random_suffix      = random_id.suffix.hex

  # Resources to monitor
  ecs_cluster_name       = module.compute.ecs_cluster_name
  backend_service_name   = module.compute.backend_service_name
  frontend_service_name  = module.compute.frontend_service_name
  alb_arn_suffix         = module.compute.alb_arn_suffix
  rds_instance_id        = module.database.rds_instance_id
  redis_cluster_id       = module.database.redis_cluster_id

  # Alert configuration
  alert_email = var.alert_email

  # Thresholds
  cpu_alarm_threshold    = var.cpu_alarm_threshold
  memory_alarm_threshold = var.memory_alarm_threshold
  error_rate_threshold   = var.error_rate_threshold
}
