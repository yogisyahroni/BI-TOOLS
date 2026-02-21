"use client";

import { useState } from "react";
import { signOut } from "next-auth/react";
import Link from "next/link";
import { usePathname } from "next/navigation";
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
  Search,
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
  Workflow,
} from "lucide-react";
import { useDatabaseStore, type Database as DatabaseType } from "@/stores/useDatabaseStore";
import { useNotifications } from "@/hooks/use-notifications";
import { useWorkspaceStore } from "@/stores/useWorkspaceStore";

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
  isCollapsed: boolean;
  onToggleCollapse: () => void;
}

export function MainSidebar({ isOpen, onClose, isCollapsed, onToggleCollapse }: SidebarProps) {
  const pathname = usePathname();
  const { databases, selectedDatabase, setSelectedDatabase } = useDatabaseStore((state) => state);
  const workspace = useWorkspaceStore((state) => state.workspace);
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
        { icon: Activity, label: "Pulses", href: "/admin/pulses" },
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
          "fixed lg:sticky top-0 left-0 z-50 h-screen bg-sidebar text-sidebar-foreground border-r border-sidebar-border flex flex-col transition-all duration-300 ease-in-out",
          isOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0",
          isCollapsed ? "w-16" : "w-64",
        )}
      >
        {/* Logo Section */}
        <div
          className={cn(
            "h-14 flex items-center justify-between px-4 border-b border-sidebar-border/50 bg-sidebar-accent/10",
            isCollapsed ? "justify-center" : "",
          )}
        >
          {!isCollapsed && (
            <Link href="/" className="flex items-center gap-3 hover:opacity-90 transition-opacity">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-primary to-cyan-500 flex items-center justify-center shadow-lg shadow-primary/20 ring-1 ring-white/10">
                <Sparkles className="h-4 w-4 text-white" />
              </div>
              <span className="font-bold text-lg tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-sidebar-foreground to-sidebar-foreground/70 bg-clip-text">
                InsightEngine
              </span>
            </Link>
          )}

          {isCollapsed && (
            <Link href="/" className="mx-auto">
              <div className="h-8 w-8 rounded-lg bg-gradient-to-br from-primary to-cyan-500 flex items-center justify-center shadow-lg shadow-primary/20">
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
        <nav className="flex-1 overflow-y-auto py-6 scrollbar-thin scrollbar-thumb-sidebar-border scrollbar-track-transparent">
          {navGroups.map((group) => (
            <div key={group.label} className="mb-6 px-3">
              {!isCollapsed && (
                <button
                  onClick={() => toggleGroup(group.label)}
                  className="w-full flex items-center justify-between px-2 py-1.5 mb-2 text-[10px] font-bold text-sidebar-foreground/50 uppercase tracking-widest hover:text-sidebar-foreground transition-colors group"
                >
                  <span>{group.label}</span>
                  <ChevronDown
                    className={cn(
                      "h-3 w-3 transition-transform duration-200 opacity-0 group-hover:opacity-100",
                      expandedGroups.includes(group.label) ? "" : "-rotate-90",
                    )}
                  />
                </button>
              )}

              {isCollapsed && (
                <div className="px-1 mb-2">
                  <div className="h-px bg-sidebar-border/50 mx-2" />
                </div>
              )}

              <div
                className={cn(
                  "space-y-1",
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
                                  ? "bg-sidebar-primary text-sidebar-primary-foreground shadow-sm"
                                  : "text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground",
                              )}
                            >
                              <Icon className="h-5 w-5" />
                            </Button>
                          </Link>
                        </TooltipTrigger>
                        <TooltipContent
                          side="right"
                          className="bg-sidebar-accent text-sidebar-foreground border-sidebar-border font-medium"
                        >
                          {item.label}
                        </TooltipContent>
                      </Tooltip>
                    );
                  }

                  return (
                    <Link key={item.label} href={item.href} onClick={onClose}>
                      <div
                        className={cn(
                          "w-full flex items-center gap-3 h-9 px-3 rounded-md transition-all duration-200 group relative",
                          active
                            ? "bg-sidebar-accent text-sidebar-foreground font-medium"
                            : "text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground",
                        )}
                      >
                        {active && (
                          <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-5 bg-primary rounded-r-full" />
                        )}
                        <Icon
                          className={cn(
                            "h-4 w-4",
                            active
                              ? "text-primary"
                              : "text-sidebar-foreground/50 group-hover:text-sidebar-foreground",
                          )}
                        />
                        <span className="flex-1 text-sm">{item.label}</span>
                      </div>
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}

          {/* Notifications in Nav */}
          <div className={cn("px-3 mt-4", isCollapsed && "px-2")}>
            {!isCollapsed && (
              <div className="px-2 py-1.5 mb-2 text-[10px] font-bold text-sidebar-foreground/50 uppercase tracking-widest">
                Communication
              </div>
            )}
            {isCollapsed ? (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Link href="/notifications">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="w-full h-10 rounded-lg text-sidebar-foreground/70 hover:bg-sidebar-accent relative"
                    >
                      <Bell className="h-5 w-5" />
                      {unreadCount > 0 && (
                        <span className="absolute top-2 right-2 h-2 w-2 rounded-full bg-destructive ring-2 ring-sidebar" />
                      )}
                    </Button>
                  </Link>
                </TooltipTrigger>
                <TooltipContent side="right">Notifications</TooltipContent>
              </Tooltip>
            ) : (
              <Link href="/notifications">
                <div
                  className={cn(
                    "w-full flex items-center gap-3 h-9 px-3 rounded-md transition-all duration-200 group relative",
                    isActive("/notifications")
                      ? "bg-sidebar-accent text-sidebar-foreground font-medium"
                      : "text-sidebar-foreground/70 hover:bg-sidebar-accent/50",
                  )}
                >
                  {isActive("/notifications") && (
                    <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-5 bg-primary rounded-r-full" />
                  )}
                  <Bell
                    className={cn(
                      "h-4 w-4",
                      isActive("/notifications")
                        ? "text-primary"
                        : "text-sidebar-foreground/50 group-hover:text-sidebar-foreground",
                    )}
                  />
                  <span className="flex-1 text-sm">Notifications</span>
                  {unreadCount > 0 && (
                    <Badge variant="destructive" className="h-5 px-1.5 text-[10px] min-w-[18px]">
                      {unreadCount}
                    </Badge>
                  )}
                </div>
              </Link>
            )}
          </div>
        </nav>

        {/* Databases Section - Collapsible */}
        {!isCollapsed && (
          <div className="border-t border-sidebar-border/50 px-4 py-4 bg-sidebar-accent/5">
            <button
              onClick={() => setShowDatabases(!showDatabases)}
              className="w-full flex items-center justify-between text-xs font-bold text-sidebar-foreground/60 uppercase tracking-widest hover:text-sidebar-foreground transition-colors mb-2"
            >
              <span>Databases</span>
              <ChevronDown
                className={cn(
                  "h-3 w-3 transition-transform duration-200",
                  showDatabases ? "" : "-rotate-90",
                )}
              />
            </button>

            {showDatabases && (
              <div className="space-y-1">
                {databases.slice(0, 5).map((db: DatabaseType) => {
                  const isSelected = selectedDatabase?.id === db.id;
                  return (
                    <button
                      key={db.id}
                      onClick={() => setSelectedDatabase(db)}
                      className={cn(
                        "w-full flex items-center gap-2 px-3 py-2 text-sm rounded-md transition-all group",
                        isSelected
                          ? "bg-sidebar-accent text-sidebar-foreground border-l-2 border-primary"
                          : "text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground hover:translate-x-1",
                      )}
                    >
                      <span
                        className={cn(
                          "w-1.5 h-1.5 rounded-full ring-2 ring-sidebar",
                          db.status === "connected" ? "bg-emerald-500" : "bg-rose-500",
                        )}
                      />
                      <span className="truncate flex-1 text-left">{db.name}</span>
                    </button>
                  );
                })}
                <Link href="/connections" onClick={onClose}>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start text-sidebar-foreground/50 hover:text-sidebar-foreground gap-2 text-xs h-8 pl-3 mt-2 border border-dashed border-sidebar-border/50"
                  >
                    <Plus className="h-3 w-3" />
                    Connect Database
                  </Button>
                </Link>
              </div>
            )}
          </div>
        )}

        {/* User Profile & Footer */}
        <div className="border-t border-sidebar-border p-3 space-y-1 bg-sidebar-accent/5">
          <Button
            variant="ghost"
            className={cn(
              "w-full justify-start gap-3 h-10 px-3 hover:bg-sidebar-accent group",
              isCollapsed ? "justify-center px-0" : "",
            )}
            onClick={() => signOut({ callbackUrl: "/auth/signin" })}
          >
            <LogOut className="h-4 w-4 text-sidebar-foreground/50 group-hover:text-destructive transition-colors" />
            {!isCollapsed && (
              <span className="text-sm text-sidebar-foreground/80 group-hover:text-destructive">
                Sign Out
              </span>
            )}
          </Button>

          <Button
            variant="ghost"
            className={cn(
              "w-full justify-center h-8 text-sidebar-foreground/40 hover:text-sidebar-foreground hover:bg-sidebar-accent/50 mt-1",
              isCollapsed ? "px-0" : "",
            )}
            onClick={onToggleCollapse}
          >
            {isCollapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>
      </aside>
    </TooltipProvider>
  );
}
