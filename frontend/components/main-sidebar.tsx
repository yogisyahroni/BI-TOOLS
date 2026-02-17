"use client";

import { useState } from "react";
import { signOut } from "next-auth/react";
import Link from "next/link";
import { usePathname, useParams } from "next/navigation";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import {
  ChevronDown,
  Plus,
  BarChart3,
  FolderOpen,
  Settings,
  LogOut,
  Zap,
  Search,
  Home,
  BookOpen,
  Database,
  PieChart,
  Upload,
  Bell,
  Activity,
  Clock,
  Boxes,
  AlertTriangle,
  Sparkles,
  LayoutDashboard,
  LineChart,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { useDatabase } from "@/contexts/database-context";
import { useNotifications } from "@/hooks/use-notifications";
import { useWorkspace } from "@/contexts/workspace-context";
import { Workflow } from "lucide-react";

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
}

// Moved inside component to access workspace context

export function MainSidebar({ isOpen, onClose, isCollapsed, onToggleCollapse }: SidebarProps) {
  const pathname = usePathname();
  const { databases, selectedDatabase, setSelectedDatabase } = useDatabase();
  const { workspace } = useWorkspace();
  const { unreadCount } = useNotifications();
  const [expandedGroups, setExpandedGroups] = useState<string[]>([
    "Analytics",
    "Data",
    "Workspace",
  ]);
  const [showDatabases, setShowDatabases] = useState(true);

  const navGroups = [
    {
      label: "Analytics",
      items: [
        { icon: LayoutDashboard, label: "Dashboards", href: "/dashboards" },
        { icon: LineChart, label: "Query Editor", href: "/query-builder" },
        { icon: PieChart, label: "Analytics", href: "/analytics" },
        { icon: AlertTriangle, label: "Alerts", href: "/alerts" },
      ],
    },
    {
      label: "Data",
      items: [
        { icon: Database, label: "Connections", href: "/connections" },
        {
          icon: Workflow,
          label: "Pipelines",
          href: workspace ? `/workspace/${workspace.id}/pipelines` : "#",
        },
        { icon: Upload, label: "Upload Data", href: "/ingest" },
        { icon: Search, label: "Explorer", href: "/explorer" },
        { icon: Boxes, label: "Lineage", href: "/lineage" },
      ],
    },
    {
      label: "Workspace",
      items: [
        { icon: BookOpen, label: "Modeling", href: "/modeling" },
        { icon: FolderOpen, label: "Collections", href: "/saved-queries" },
        { icon: BarChart3, label: "Story Builder", href: "/stories" },
        { icon: Clock, label: "Scheduler", href: "/admin/scheduler" },
      ],
    },
  ];

  const isActive = (path: string) => {
    if (path === "/") return pathname === "/";
    return pathname.startsWith(path);
  };

  const toggleGroup = (label: string) => {
    setExpandedGroups((prev) =>
      prev.includes(label) ? prev.filter((l) => l !== label) : [...prev, label],
    );
  };

  return (
    <TooltipProvider delayDuration={0}>
      {/* Mobile Overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 lg:hidden"
          onClick={onClose}
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          "fixed lg:sticky top-0 left-0 z-50 h-screen bg-sidebar border-r border-sidebar-border flex flex-col transition-all duration-300 ease-in-out",
          isOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0",
          isCollapsed ? "w-16" : "w-64",
        )}
      >
        {/* Logo Section */}
        <div className="h-16 flex items-center justify-between px-4 border-b border-sidebar-border">
          {!isCollapsed && (
            <Link href="/" className="flex items-center gap-2.5">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-primary to-cyan-500 flex items-center justify-center shadow-lg shadow-primary/30">
                <Sparkles className="h-4 w-4 text-white" />
              </div>
              <span className="font-bold text-lg text-sidebar-foreground tracking-tight">
                InsightEngine
              </span>
            </Link>
          )}

          {isCollapsed && (
            <Link href="/" className="mx-auto">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-primary to-cyan-500 flex items-center justify-center shadow-lg shadow-primary/30">
                <Sparkles className="h-4 w-4 text-white" />
              </div>
            </Link>
          )}

          <button
            onClick={onClose}
            className="lg:hidden text-sidebar-foreground hover:bg-sidebar-accent p-1.5 rounded-md transition-colors"
          >
            <ChevronLeft className="h-5 w-5" />
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 overflow-y-auto py-4 scrollbar-thin scrollbar-thumb-sidebar-border">
          {navGroups.map((group) => (
            <div key={group.label} className="mb-2">
              {!isCollapsed && (
                <button
                  onClick={() => toggleGroup(group.label)}
                  className="w-full flex items-center justify-between px-4 py-2 text-xs font-semibold text-sidebar-foreground/60 uppercase tracking-wider hover:text-sidebar-foreground transition-colors"
                >
                  <span>{group.label}</span>
                  <ChevronDown
                    className={cn(
                      "h-3.5 w-3.5 transition-transform duration-200",
                      expandedGroups.includes(group.label) ? "" : "-rotate-90",
                    )}
                  />
                </button>
              )}

              {isCollapsed && (
                <div className="px-2 mb-2">
                  <div className="h-px bg-sidebar-border mx-2" />
                </div>
              )}

              <div
                className={cn(
                  "space-y-0.5",
                  !isCollapsed && !expandedGroups.includes(group.label) && "hidden",
                )}
              >
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const active = isActive(item.href);

                  if (isCollapsed) {
                    return (
                      <Tooltip key={item.label}>
                        <TooltipTrigger asChild>
                          <Link href={item.href} onClick={onClose}>
                            <Button
                              variant="ghost"
                              size="icon"
                              className={cn(
                                "w-full h-10 rounded-lg transition-all duration-200",
                                active
                                  ? "bg-primary text-primary-foreground shadow-md"
                                  : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground",
                              )}
                            >
                              <Icon className="h-5 w-5" />
                            </Button>
                          </Link>
                        </TooltipTrigger>
                        <TooltipContent side="right" className="border-sidebar-border">
                          {item.label}
                        </TooltipContent>
                      </Tooltip>
                    );
                  }

                  return (
                    <Link key={item.label} href={item.href} onClick={onClose}>
                      <Button
                        variant="ghost"
                        className={cn(
                          "w-full justify-start gap-3 h-10 px-4 rounded-lg transition-all duration-200",
                          active
                            ? "bg-primary text-primary-foreground shadow-md shadow-primary/20 font-medium"
                            : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground",
                        )}
                      >
                        <Icon className="h-4.5 w-4.5" />
                        <span className="flex-1 text-left">{item.label}</span>
                      </Button>
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}

          {/* Notifications */}
          <div className={cn("mt-2", isCollapsed && "px-2")}>
            {!isCollapsed && (
              <div className="px-4 py-2 text-xs font-semibold text-sidebar-foreground/60 uppercase tracking-wider">
                Activity
              </div>
            )}
            {isCollapsed && <div className="h-px bg-sidebar-border mx-2 mb-2" />}

            {isCollapsed ? (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Link href="/notifications" onClick={onClose}>
                    <Button
                      variant="ghost"
                      size="icon"
                      className={cn(
                        "w-full h-10 rounded-lg relative",
                        isActive("/notifications")
                          ? "bg-primary text-primary-foreground"
                          : "text-sidebar-foreground/70 hover:bg-sidebar-accent",
                      )}
                    >
                      <Bell className="h-5 w-5" />
                      {unreadCount > 0 && (
                        <Badge
                          variant="destructive"
                          className="absolute -top-1 -right-1 h-5 min-w-[18px] px-1 text-[10px]"
                        >
                          {unreadCount}
                        </Badge>
                      )}
                    </Button>
                  </Link>
                </TooltipTrigger>
                <TooltipContent side="right">Notifications</TooltipContent>
              </Tooltip>
            ) : (
              <Link href="/notifications" onClick={onClose}>
                <Button
                  variant="ghost"
                  className={cn(
                    "w-full justify-start gap-3 h-10 px-4 rounded-lg",
                    isActive("/notifications")
                      ? "bg-primary text-primary-foreground"
                      : "text-sidebar-foreground/70 hover:bg-sidebar-accent",
                  )}
                >
                  <Bell className="h-4.5 w-4.5" />
                  <span className="flex-1 text-left">Notifications</span>
                  {unreadCount > 0 && (
                    <Badge variant="destructive" className="h-5 min-w-[20px] px-1.5 text-xs">
                      {unreadCount > 99 ? "99+" : unreadCount}
                    </Badge>
                  )}
                </Button>
              </Link>
            )}
          </div>
        </nav>

        {/* Databases Section */}
        {!isCollapsed && (
          <div className="border-t border-sidebar-border px-3 py-3">
            <button
              onClick={() => setShowDatabases(!showDatabases)}
              className="w-full flex items-center justify-between px-3 py-2 text-xs font-semibold text-sidebar-foreground/60 uppercase tracking-wider hover:text-sidebar-foreground transition-colors"
            >
              <span>Databases</span>
              <ChevronDown
                className={cn(
                  "h-3.5 w-3.5 transition-transform duration-200",
                  showDatabases ? "" : "-rotate-90",
                )}
              />
            </button>

            {showDatabases && (
              <div className="mt-2 space-y-1">
                {databases.slice(0, 5).map((db) => {
                  const isSelected = selectedDatabase?.id === db.id;
                  return (
                    <button
                      key={db.id}
                      onClick={() => setSelectedDatabase(db)}
                      className={cn(
                        "w-full flex items-center gap-2 px-3 py-2 text-sm rounded-lg transition-all",
                        isSelected
                          ? "bg-primary text-primary-foreground font-medium"
                          : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground",
                      )}
                    >
                      <span
                        className={cn(
                          "w-2 h-2 rounded-full",
                          db.status === "connected" ? "bg-green-400" : "bg-red-400",
                        )}
                      />
                      <span className="truncate">{db.name}</span>
                    </button>
                  );
                })}
                {databases.length > 5 && (
                  <p className="px-3 text-xs text-sidebar-foreground/50">
                    +{databases.length - 5} more
                  </p>
                )}
                <Link href="/connections" onClick={onClose}>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start text-sidebar-foreground/50 hover:text-sidebar-foreground gap-2 text-xs mt-1"
                  >
                    <Plus className="h-3.5 w-3.5" />
                    Add Connection
                  </Button>
                </Link>
              </div>
            )}
          </div>
        )}

        {/* Footer */}
        <div className="border-t border-sidebar-border p-3 space-y-1">
          {/* Collapse Toggle */}
          <Button
            variant="ghost"
            size="sm"
            onClick={onToggleCollapse}
            className="w-full justify-center text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground"
          >
            {isCollapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4 mr-2" />
            )}
            {!isCollapsed && <span>Collapse</span>}
          </Button>

          {/* Settings & Logout */}
          {!isCollapsed ? (
            <>
              <Link href="/settings" onClick={onClose}>
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full justify-start gap-2 text-sidebar-foreground/70 hover:text-sidebar-foreground"
                >
                  <Settings className="h-4 w-4" />
                  Settings
                </Button>
              </Link>
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start gap-2 text-sidebar-foreground/70 hover:text-destructive transition-colors"
                onClick={() => signOut({ callbackUrl: "/auth/signin" })}
              >
                <LogOut className="h-4 w-4" />
                Logout
              </Button>
            </>
          ) : (
            <>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Link href="/settings" onClick={onClose}>
                    <Button variant="ghost" size="icon" className="w-full h-9">
                      <Settings className="h-4 w-4" />
                    </Button>
                  </Link>
                </TooltipTrigger>
                <TooltipContent side="right">Settings</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="w-full h-9 text-sidebar-foreground/70 hover:text-destructive"
                    onClick={() => signOut({ callbackUrl: "/auth/signin" })}
                  >
                    <LogOut className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent side="right">Logout</TooltipContent>
              </Tooltip>
            </>
          )}
        </div>
      </aside>
    </TooltipProvider>
  );
}
