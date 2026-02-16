# ============================================================================
# Staging Environment — terraform.tfvars
# ============================================================================
# Usage: terraform plan -var-file=environments/staging/terraform.tfvars
# ============================================================================

environment = "staging"
aws_region  = "ap-southeast-1"

# Networking
vpc_cidr             = "10.1.0.0/16"
availability_zones   = ["ap-southeast-1a", "ap-southeast-1b"]
public_subnet_cidrs  = ["10.1.1.0/24", "10.1.2.0/24"]
private_subnet_cidrs = ["10.1.11.0/24", "10.1.12.0/24"]

# Database — PostgreSQL (smaller for cost)
db_instance_class    = "db.t3.medium"
db_allocated_storage = 20
db_name              = "insightengine_staging"
db_username          = "insightadmin"
db_multi_az          = false

# Database — Redis (single node)
redis_node_type       = "cache.t3.micro"
redis_num_cache_nodes = 1

# Compute — Backend
backend_image         = "ACCOUNT_ID.dkr.ecr.ap-southeast-1.amazonaws.com/insight-engine-backend:staging"
backend_cpu           = 512
backend_memory        = 1024
backend_desired_count = 1
backend_max_count     = 3

# Compute — Frontend
frontend_image         = "ACCOUNT_ID.dkr.ecr.ap-southeast-1.amazonaws.com/insight-engine-frontend:staging"
frontend_cpu           = 256
frontend_memory        = 512
frontend_desired_count = 1
frontend_max_count     = 2

# Secrets (override via environment variables)
# jwt_secret = "OVERRIDE_VIA_ENV_TF_VAR_jwt_secret"

# ACM Certificate
acm_certificate_arn = "arn:aws:acm:ap-southeast-1:ACCOUNT_ID:certificate/STAGING_CERT_ID"

# Monitoring
alert_email            = "platform-staging@insightengine.io"
cpu_alarm_threshold    = 85
memory_alarm_threshold = 90
error_rate_threshold   = 100
