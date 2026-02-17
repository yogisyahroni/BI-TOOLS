-- Add calculated_fields column to DashboardCard table
ALTER TABLE "DashboardCard"
ADD COLUMN IF NOT EXISTS "calculated_fields" JSONB;