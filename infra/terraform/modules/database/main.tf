# ============================================================================
# Database Module — RDS PostgreSQL + ElastiCache Redis
# ============================================================================
# PostgreSQL: Primary data store with automated backups, encryption at rest,
# and optional Multi-AZ for production high availability.
# Redis: Session cache, rate limiting, and real-time pub/sub.
# ============================================================================

# ===========================================================================
# RDS Subnet Group
# ===========================================================================
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-${var.environment}-db-subnet-${var.random_suffix}"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.project_name}-${var.environment}-db-subnet-group"
  }
}

# ===========================================================================
# RDS PostgreSQL — Random Password Generation
# ===========================================================================
resource "random_password" "db_password" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

# ===========================================================================
# Secrets Manager — Store DB Password
# ===========================================================================
resource "aws_secretsmanager_secret" "db_password" {
  name                    = "${var.project_name}-${var.environment}-db-password-${var.random_suffix}"
  description             = "RDS PostgreSQL master password"
  recovery_window_in_days = 7

  tags = {
    Name = "${var.project_name}-${var.environment}-db-password"
  }
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = random_password.db_password.result
}

# ===========================================================================
# RDS PostgreSQL Instance
# ===========================================================================
resource "aws_db_instance" "postgres" {
  identifier = "${var.project_name}-${var.environment}-pg-${var.random_suffix}"

  engine               = "postgres"
  engine_version       = "16.4"
  instance_class       = var.db_instance_class
  allocated_storage    = var.db_allocated_storage
  max_allocated_storage = var.db_allocated_storage * 2

  db_name  = var.db_name
  username = var.db_username
  password = random_password.db_password.result

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.db_security_group_id]

  # High Availability
  multi_az = var.db_multi_az

  # Storage
  storage_type      = "gp3"
  storage_encrypted = true

  # Backup & Maintenance
  backup_retention_period  = 7
  backup_window            = "03:00-04:00"
  maintenance_window       = "sun:05:00-sun:06:00"
  copy_tags_to_snapshot    = true
  delete_automated_backups = false

  # Monitoring
  performance_insights_enabled          = true
  performance_insights_retention_period = 7
  monitoring_interval                   = 60
  monitoring_role_arn                   = aws_iam_role.rds_monitoring.arn

  # Lifecycle
  skip_final_snapshot       = false
  final_snapshot_identifier = "${var.project_name}-${var.environment}-final-${var.random_suffix}"
  deletion_protection       = var.environment == "production"

  # Parameter group
  parameter_group_name = aws_db_parameter_group.postgres.name

  tags = {
    Name = "${var.project_name}-${var.environment}-postgres"
  }
}

# ===========================================================================
# RDS Parameter Group — Performance Tuning
# ===========================================================================
resource "aws_db_parameter_group" "postgres" {
  name_prefix = "${var.project_name}-${var.environment}-pg16-"
  family      = "postgres16"
  description = "Custom parameter group for InsightEngine PostgreSQL 16"

  parameter {
    name  = "log_min_duration_statement"
    value = "1000"
  }

  parameter {
    name  = "shared_preload_libraries"
    value = "pg_stat_statements"
  }

  parameter {
    name  = "pg_stat_statements.track"
    value = "all"
  }

  parameter {
    name  = "log_connections"
    value = "1"
  }

  parameter {
    name  = "log_disconnections"
    value = "1"
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name = "${var.project_name}-${var.environment}-pg-params"
  }
}

# ===========================================================================
# IAM Role for Enhanced Monitoring
# ===========================================================================
resource "aws_iam_role" "rds_monitoring" {
  name_prefix = "${var.project_name}-${var.environment}-rds-mon-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "monitoring.rds.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name = "${var.project_name}-${var.environment}-rds-monitoring-role"
  }
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  role       = aws_iam_role.rds_monitoring.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

# ===========================================================================
# ElastiCache Redis — Subnet Group
# ===========================================================================
resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.project_name}-${var.environment}-redis-subnet-${var.random_suffix}"
  subnet_ids = var.private_subnet_ids

  tags = {
    Name = "${var.project_name}-${var.environment}-redis-subnet-group"
  }
}

# ===========================================================================
# ElastiCache Redis Cluster
# ===========================================================================
resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "${var.project_name}-${var.environment}-redis-${var.random_suffix}"
  engine               = "redis"
  engine_version       = "7.1"
  node_type            = var.redis_node_type
  num_cache_nodes      = var.redis_num_cache_nodes
  port                 = 6379
  parameter_group_name = "default.redis7"

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = [var.redis_security_group_id]

  snapshot_retention_limit = 3
  snapshot_window          = "04:00-05:00"
  maintenance_window       = "sun:06:00-sun:07:00"

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true

  tags = {
    Name = "${var.project_name}-${var.environment}-redis"
  }
}
