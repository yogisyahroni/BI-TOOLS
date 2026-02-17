'use client';

import dynamic from 'next/dynamic';
import { _Suspense, type ComponentType } from 'react';
import { cn } from '@/lib/utils';

// Skeleton loader for Resizable components (GEMINI.md 4.2.3)
function ResizableSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("animate-pulse bg-muted rounded-md h-full w-full", className)}>
      <div className="h-full w-full bg-gradient-to-r from-muted via-muted/50 to-muted" />
    </div>
  );
}

// Type definitions for resizable props
import type { 
  ComponentProps 
} from 'react';

// Dynamically import Resizable components to avoid SSR issues
const ResizablePanelGroup = dynamic(
  () => import('@/components/ui/resizable').then((mod) => mod.ResizablePanelGroup),
  { 
    ssr: false,
    loading: () => <ResizableSkeleton />
  }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
) as ComponentType<ComponentProps<any>>;

const ResizablePanel = dynamic(
  () => import('@/components/ui/resizable').then((mod) => mod.ResizablePanel),
  { 
    ssr: false,
    loading: () => <ResizableSkeleton />
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
  }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
) as ComponentType<ComponentProps<any>>;

const ResizableHandle = dynamic(
  () => import('@/components/ui/resizable').then((mod) => mod.ResizableHandle),
  { 
    ssr: false,
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    loading: () => <ResizableSkeleton className="w-px h-full" />
  }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
) as ComponentType<ComponentProps<any>>;

export { ResizablePanelGroup, ResizablePanel, ResizableHandle };
export { ResizableSkeleton };
