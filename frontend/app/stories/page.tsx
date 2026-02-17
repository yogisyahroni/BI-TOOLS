"use client";

import { MainSidebar } from "@/components/main-sidebar";
import { StoryBuilder } from "@/components/story-builder/StoryBuilder";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/toaster";
import { useState } from "react";

export default function StoriesPage() {
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  return (
    <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
      <div className="flex h-screen bg-background">
        <MainSidebar
          isCollapsed={isSidebarCollapsed}
          onToggle={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
        />
        <main className="flex-1 overflow-y-auto">
          <StoryBuilder />
        </main>
        <Toaster />
      </div>
    </ThemeProvider>
  );
}
