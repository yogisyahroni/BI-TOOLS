/**
 * storyService — mirrors the pattern from pulse-service.ts which works correctly.
 * Token comes from NextAuth session (session.accessToken or session.user.token).
 */
import { getSession } from "next-auth/react";
import {
  Story,
  CreateStoryRequest,
  CreateManualStoryRequest,
  UpdateStoryRequest,
} from "@/types/story";

const API_BASE_URL = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api").replace(
  /\/$/,
  "",
);
const API_URL = API_BASE_URL.endsWith("/api") ? API_BASE_URL : `${API_BASE_URL}/api`;

const getAuthHeaders = async (): Promise<HeadersInit> => {
  const session = await getSession();
  // Token lives on the root of the session object (set by auth-options.ts callbacks)
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const token = (session as any)?.accessToken || (session as any)?.user?.token;

  if (!token) {
    console.error("storyService: No access token found in session");
  }

  return {
    "Content-Type": "application/json",
    ...(token && { Authorization: `Bearer ${token}` }),
  };
};

export const storyService = {
  async getStories(): Promise<Story[]> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories`, { headers });
    if (!response.ok) {
      if (response.status === 401) throw new Error("Unauthorized");
      throw new Error("Failed to fetch stories");
    }
    return response.json();
  },

  async getStory(id: string): Promise<Story> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories/${id}`, { headers });
    if (!response.ok) throw new Error("Failed to fetch story");
    return response.json();
  },

  /** AI-powered story creation (requires dashboard + prompt). */
  async createStory(data: CreateStoryRequest): Promise<Story> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories`, {
      method: "POST",
      headers,
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to create story");
    }
    return response.json();
  },

  /** Manual story creation — no AI required. Starts with a blank slide. */
  async createManualStory(data: CreateManualStoryRequest): Promise<Story> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories/manual`, {
      method: "POST",
      headers,
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to create story");
    }
    return response.json();
  },

  async updateStory(id: string, data: UpdateStoryRequest): Promise<Story> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories/${id}`, {
      method: "PUT",
      headers,
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error("Failed to update story");
    return response.json();
  },

  async deleteStory(id: string): Promise<void> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories/${id}`, {
      method: "DELETE",
      headers,
    });
    if (!response.ok) throw new Error("Failed to delete story");
  },

  async exportPPTX(id: string, title: string): Promise<void> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories/${id}/export`, {
      method: "GET",
      headers,
    });
    if (!response.ok) throw new Error("Failed to export PPTX");

    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${title.replace(/[^a-z0-9]/gi, "_").toLowerCase()}.pptx`;
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  },

  async togglePublicShare(
    id: string,
    is_public: boolean,
  ): Promise<{ is_public: boolean; share_token: string }> {
    const headers = await getAuthHeaders();
    const response = await fetch(`${API_URL}/stories/${id}/share`, {
      method: "PUT",
      headers,
      body: JSON.stringify({ is_public }),
    });
    if (!response.ok) throw new Error("Failed to toggle public share");
    return response.json();
  },
};
