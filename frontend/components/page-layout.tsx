"use client";

import React from "react";
import { MainSidebar } from "./main-sidebar";
import { TopBar } from "./top-bar";
import { useSidebarStore } from "@/stores/useSidebarStore";
import { cn } from "@/lib/utils";

interface PageLayoutProps {
  children: React.ReactNode;
  className?: string;
}

export function PageLayout({ children, className }: PageLayoutProps) {
  const { isOpen, close, isCollapsed, toggleCollapse } = useSidebarStore((state) => state);

  return (
    <div className="flex min-h-screen bg-background">
      {/* Sidebar */}
      <MainSidebar
        isOpen={isOpen}
        onClose={close}
        isCollapsed={isCollapsed}
        onToggleCollapse={toggleCollapse}
      />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 transition-all duration-300">
        <TopBar />

        <main className={cn("flex-1 overflow-auto p-4 lg:p-8", className)}>
          <div className="max-w-7xl mx-auto">{children}</div>
        </main>
      </div>
    </div>
  );
}
