export interface Pulse {
  id: string;
  name: string;
  dashboardId: string;
  schedule: string; // Cron expression
  config: PulseConfig;
  isActive: boolean;
  webhookUrl: string; // For slack/teams webhook
  userId: string;
  createdAt: string;
  updatedAt: string;
  lastRunAt?: string;
  nextRunAt?: string;
}

export interface PulseConfig {
  width: number;
  height: number;
  format: "png" | "pdf";
  includeFilters?: boolean;
}

export interface CreatePulseRequest {
  name: string;
  dashboardId: string;
  schedule: string;
  config: PulseConfig;
  isActive: boolean;
  webhookUrl: string;
}

export interface UpdatePulseRequest {
  name?: string;
  dashboardId?: string;
  schedule?: string;
  config?: PulseConfig;
  isActive?: boolean;
  webhookUrl?: string;
}
