import { create } from "zustand";
import { persist } from "zustand/middleware";
import { workspaceApi } from "@/lib/api/workspaces";
import { hasPermission, type Permission, type Role } from "@/lib/rbac/permissions";
import type { Workspace as ApiWorkspace } from "@/lib/types/batch3";

export interface Workspace extends ApiWorkspace {
  slug: string;
  plan: "FREE" | "PRO" | "ENTERPRISE";
  role: Role;
}

const mapApiToWorkspace = (apiW: ApiWorkspace): Workspace => ({
  ...apiW,
  slug: apiW.name.toLowerCase().replace(/\s+/g, "-"),
  plan: "FREE",
  role: "OWNER",
});

interface WorkspaceState {
  workspace: Workspace | null;
  isLoading: boolean;
  setWorkspace: (workspace: Workspace | null) => void;
  hasPermission: (permission: Permission) => boolean;
  initialize: (isAuthenticated: boolean) => Promise<void>;
}

export const useWorkspaceStore = create<WorkspaceState>()(
  persist(
    (set, get) => ({
      workspace: null,
      isLoading: true,

      setWorkspace: (workspace) => set({ workspace }),

      hasPermission: (permission: Permission) => {
        const { workspace } = get();
        if (!workspace) return false;
        return hasPermission(workspace.role, permission);
      },

      initialize: async (isAuthenticated: boolean) => {
        if (!isAuthenticated) {
          set({ isLoading: false });
          return;
        }

        try {
          // If we already have a persisted workspace, just stop loading
          if (get().workspace) {
            set({ isLoading: false });
            return;
          }

          // Otherwise fetch from API
          const workspaces = await workspaceApi.list();
          if (workspaces && workspaces.length > 0) {
            set({ workspace: mapApiToWorkspace(workspaces[0]) });
          } else {
            console.log("[WorkspaceStore] No workspaces found, creating default...");
            try {
              const newWorkspace = await workspaceApi.create({
                name: "My Workspace",
                description: "Default workspace",
              });
              set({ workspace: mapApiToWorkspace(newWorkspace) });
            } catch (createError) {
              console.error("[WorkspaceStore] Failed to create default workspace:", createError);
            }
          }
        } catch (error) {
          console.error("[WorkspaceStore] Failed to initialize workspace:", error);
        } finally {
          set({ isLoading: false });
        }
      },
    }),
    {
      name: "activeWorkspace", // unique name
      partialize: (state) => ({ workspace: state.workspace }), // only persist workspace
    },
  ),
);
