variable "project_name" { type = string }
variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "db_security_group_id" { type = string }
variable "redis_security_group_id" { type = string }
variable "random_suffix" { type = string }

variable "db_instance_class" { type = string }
variable "db_allocated_storage" { type = number }
variable "db_name" { type = string }
variable "db_username" { type = string }
variable "db_multi_az" { type = bool }

variable "redis_node_type" { type = string }
variable "redis_num_cache_nodes" { type = number }
