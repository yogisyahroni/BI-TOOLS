# ============================================================================
# InsightEngine AI — Root Outputs
# ============================================================================

output "vpc_id" {
  description = "ID of the created VPC"
  value       = module.networking.vpc_id
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.compute.alb_dns_name
}

output "alb_zone_id" {
  description = "Zone ID of the ALB (for Route53 alias records)"
  value       = module.compute.alb_zone_id
}

output "rds_endpoint" {
  description = "RDS PostgreSQL endpoint"
  value       = module.database.rds_endpoint
  sensitive   = true
}

output "redis_endpoint" {
  description = "ElastiCache Redis endpoint"
  value       = module.database.redis_endpoint
  sensitive   = true
}

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = module.compute.ecs_cluster_name
}

output "backend_service_name" {
  description = "Name of the backend ECS service"
  value       = module.compute.backend_service_name
}

output "frontend_service_name" {
  description = "Name of the frontend ECS service"
  value       = module.compute.frontend_service_name
}

output "cloudwatch_dashboard_url" {
  description = "URL to the CloudWatch monitoring dashboard"
  value       = module.monitoring.dashboard_url
}
