'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { hasPermission, type Permission, type Role } from '@/lib/rbac/permissions';
import { workspaceApi } from '@/lib/api/workspaces';

export interface Workspace {
    id: string;
    name: string;
    slug: string;
    plan: 'FREE' | 'PRO' | 'ENTERPRISE';
    role: Role;
}

interface WorkspaceContextType {
    workspace: Workspace | null;
    setWorkspace: (workspace: Workspace | null) => void;
    hasPermission: (permission: Permission) => boolean;
    isLoading: boolean;
}

const WorkspaceContext = createContext<WorkspaceContextType | undefined>(undefined);

export function WorkspaceProvider({ children }: { children: ReactNode }) {
    const [workspace, setWorkspaceState] = useState<Workspace | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    // Load from localStorage or API on mount
    useEffect(() => {
        const initializeWorkspace = async () => {
            try {
                // 1. Try local storage
                const saved = localStorage.getItem('activeWorkspace');
                if (saved) {
                    const parsed = JSON.parse(saved);
                    setWorkspaceState(parsed);
                    return;
                }

                // 2. If not found, fetch from API
                const workspaces = await workspaceApi.list();

                if (workspaces && workspaces.length > 0) {
                    const defaultWorkspace = workspaces[0];
                    setWorkspaceState(defaultWorkspace);
                    localStorage.setItem('activeWorkspace', JSON.stringify(defaultWorkspace));
                } else {
                    // 3. If no workspaces exist, create a default one
                    console.log('[WorkspaceProvider] No workspaces found, creating default...');
                    try {
                        const newWorkspace = await workspaceApi.create({
                            name: 'My Workspace',
                            description: 'Default workspace'
                        });
                        setWorkspaceState(newWorkspace);
                        localStorage.setItem('activeWorkspace', JSON.stringify(newWorkspace));
                    } catch (createError) {
                        console.error('[WorkspaceProvider] Failed to create default workspace:', createError);
                    }
                }
            } catch (error) {
                console.error('[WorkspaceProvider] Failed to initialize workspace:', error);
            } finally {
                setIsLoading(false);
            }
        };

        initializeWorkspace();
    }, []);

    // Persist to localStorage when workspace changes
    const setWorkspace = (newWorkspace: Workspace | null) => {
        setWorkspaceState(newWorkspace);

        try {
            if (newWorkspace) {
                localStorage.setItem('activeWorkspace', JSON.stringify(newWorkspace));
            } else {
                localStorage.removeItem('activeWorkspace');
            }
        } catch (error) {
            // Silently handle localStorage errors
        }
    };

    // Check if current user has a specific permission
    const checkPermission = (permission: Permission): boolean => {
        if (!workspace) return false;
        return hasPermission(workspace.role, permission);
    };

    return (
        <WorkspaceContext.Provider
            value={{
                workspace,
                setWorkspace,
                hasPermission: checkPermission,
                isLoading
            }}
        >
            {children}
        </WorkspaceContext.Provider>
    );
}

export function useWorkspace() {
    const context = useContext(WorkspaceContext);
    if (!context) {
        throw new Error('useWorkspace must be used within WorkspaceProvider');
    }
    return context;
}
