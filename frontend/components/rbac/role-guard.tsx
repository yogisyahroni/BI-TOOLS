"use client";

import { useWorkspaceStore } from "@/stores/useWorkspaceStore";
import { type Permission } from "@/lib/rbac/permissions";
import { type ReactNode } from "react";

interface RoleGuardProps {
  permission: Permission;
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * Conditionally render children based on user's permission in active workspace
 *
 * @example
 * ```tsx
 * <RoleGuard permission="connection:create">
 *   <Button>Create Connection</Button>
 * </RoleGuard>
 * ```
 */
export function RoleGuard({ permission, children, fallback = null }: RoleGuardProps) {
  const hasPermission = useWorkspaceStore((state) => state.hasPermission);

  return hasPermission(permission) ? <>{children}</> : <>{fallback}</>;
}

/**
 * Show content only if user LACKS the permission (inverse of RoleGuard)
 */
export function RoleGuardInverse({ permission, children }: Omit<RoleGuardProps, "fallback">) {
  const hasPermission = useWorkspaceStore((state) => state.hasPermission);

  return !hasPermission(permission) ? <>{children}</> : null;
}
