import { getSession } from "next-auth/react";

const API_BASE_URL = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api").replace(
  /\/$/,
  "",
);
const API_URL = API_BASE_URL.endsWith("/api") ? API_BASE_URL : `${API_BASE_URL}/api`;

export interface Pulse {
  id: string;
  dashboard_id: string;
  query_id?: string;
  name: string;
  schedule_interval: string;
  channel: "slack" | "teams" | "email";
  destination?: string;
  is_active: boolean;
  last_run?: string;
  next_run?: string;
}

export interface CreatePulseRequest {
  dashboard_id: string;
  query_id?: string;
  name: string;
  schedule_interval: string;
  channel: "slack" | "teams" | "email";
  destination?: string;
}

export interface UpdatePulseRequest {
  name?: string;
  schedule_interval?: string;
  channel?: "slack" | "teams" | "email";
  destination?: string;
  is_active?: boolean;
}

const getAuthHeaders = async () => {
  const session = await getSession();
  // Debug logs can be helpful, keeping them for now but cleaned up
  // console.log('DEBUG: PulseService session:', session);

  // Correctly accessing the token from the root of the session object based on auth-options.ts
  const token = (session as any)?.accessToken || (session as any)?.user?.token;

  if (!token) {
    console.error("PulseService: No access token found in session");
    // Optionally throw an error here if authentication is mandatory
    // throw new Error('Unauthorized');
  }

  return {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };
};

export const pulseService = {
  async getPulses(): Promise<Pulse[]> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/pulses`, {
      headers: headers,
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to fetch pulses");
    }
    return response.json();
  },

  async createPulse(pulse: CreatePulseRequest): Promise<Pulse> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/pulses`, {
      method: "POST",
      headers: headers,
      body: JSON.stringify(pulse),
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to create pulse");
    }
    return response.json();
  },

  async updatePulse(id: string, pulse: UpdatePulseRequest): Promise<Pulse> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/pulses/${id}`, {
      method: "PUT",
      headers: headers,
      body: JSON.stringify(pulse),
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to update pulse");
    }
    return response.json();
  },

  async deletePulse(id: string): Promise<void> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/pulses/${id}`, {
      method: "DELETE",
      headers: headers,
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to delete pulse");
    }
  },

  async triggerPulse(id: string): Promise<void> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/pulses/${id}/trigger`, {
      method: "POST",
      headers: headers,
    });
    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to trigger pulse");
    }
  },
};
