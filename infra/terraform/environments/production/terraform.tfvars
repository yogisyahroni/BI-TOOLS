# ============================================================================
# Production Environment — terraform.tfvars
# ============================================================================
# Usage: terraform plan -var-file=environments/production/terraform.tfvars
# ============================================================================

environment = "production"
aws_region  = "ap-southeast-1"

# Networking
vpc_cidr             = "10.0.0.0/16"
availability_zones   = ["ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c"]
public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
private_subnet_cidrs = ["10.0.11.0/24", "10.0.12.0/24", "10.0.13.0/24"]

# Database — PostgreSQL
db_instance_class    = "db.r6g.large"
db_allocated_storage = 100
db_name              = "insightengine"
db_username          = "insightadmin"
db_multi_az          = true

# Database — Redis
redis_node_type       = "cache.r6g.large"
redis_num_cache_nodes = 2

# Compute — Backend (Go)
backend_image         = "ACCOUNT_ID.dkr.ecr.ap-southeast-1.amazonaws.com/insight-engine-backend:latest"
backend_cpu           = 1024
backend_memory        = 2048
backend_desired_count = 3
backend_max_count     = 10

# Compute — Frontend (Next.js)
frontend_image         = "ACCOUNT_ID.dkr.ecr.ap-southeast-1.amazonaws.com/insight-engine-frontend:latest"
frontend_cpu           = 512
frontend_memory        = 1024
frontend_desired_count = 3
frontend_max_count     = 8

# Secrets (override via environment variables or -var flag, NEVER commit values)
# jwt_secret = "OVERRIDE_VIA_ENV_TF_VAR_jwt_secret"

# ACM Certificate
acm_certificate_arn = "arn:aws:acm:ap-southeast-1:ACCOUNT_ID:certificate/CERT_ID"

# Monitoring
alert_email            = "platform-alerts@insightengine.io"
cpu_alarm_threshold    = 75
memory_alarm_threshold = 80
error_rate_threshold   = 25
