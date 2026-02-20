import type {
  SchedulerJob,
  CreateSchedulerJobInput,
  UpdateSchedulerJobInput,
} from "@/lib/types/notifications";
import { fetchWithAuth } from "@/lib/utils";

// Use strict backend URL matching pulse-service.ts logic
const API_BASE_URL = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api").replace(
  /\/$/,
  "",
);
const API_BASE = API_BASE_URL.endsWith("/api") ? API_BASE_URL : `${API_BASE_URL}/api`;

export const listJobs = async (): Promise<SchedulerJob[]> => {
  // Direct call to backend
  const res = await fetchWithAuth(`${API_BASE}/scheduler/jobs`);
  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.error || `Failed to fetch jobs: ${res.statusText}`);
  }
  const data = await res.json();
  return data.jobs || [];
};

export const createJob = async (job: CreateSchedulerJobInput): Promise<SchedulerJob> => {
  const res = await fetchWithAuth(`${API_BASE}/scheduler/jobs`, {
    method: "POST",
    body: JSON.stringify(job),
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.error || "Failed to create job");
  }
  return res.json();
};

export const updateJob = async (
  id: string,
  job: UpdateSchedulerJobInput,
): Promise<SchedulerJob> => {
  const res = await fetchWithAuth(`${API_BASE}/scheduler/jobs/${id}`, {
    method: "PUT",
    body: JSON.stringify(job),
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.error || "Failed to update job");
  }
  return res.json();
};

export const deleteJob = async (id: string): Promise<void> => {
  const res = await fetchWithAuth(`${API_BASE}/scheduler/jobs/${id}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.error || "Failed to delete job");
  }
};

export const triggerJob = async (id: string): Promise<void> => {
  const res = await fetchWithAuth(`${API_BASE}/scheduler/jobs/${id}/trigger`, {
    method: "POST",
  });
  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.error || "Failed to trigger job");
  }
};

export const getJobHistory = async (id: string): Promise<any[]> => {
  const res = await fetchWithAuth(`${API_BASE}/scheduler/jobs/${id}/history`);
  if (!res.ok) {
    const error = await res.json().catch(() => ({}));
    throw new Error(error.error || "Failed to fetch job history");
  }
  return res.json();
};

// Export as object for backward compatibility
export const schedulerApi = {
  listJobs,
  createJob,
  updateJob,
  deleteJob,
  triggerJob,
  getJobHistory,
};
