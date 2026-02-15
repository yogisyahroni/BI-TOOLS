'use client';

export const dynamic = 'force-dynamic';

import { useSession } from 'next-auth/react';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { Loader2 } from 'lucide-react';

export default function Home() {
  const { data: session, status } = useSession();
  const router = useRouter();

  useEffect(() => {
    if (status === 'loading') return;

    if (status === 'authenticated' && session) {
      // Redirect authenticated users to dashboards
      router.replace('/dashboards');
    } else {
      // Redirect unauthenticated users to login
      router.replace('/auth/signin');
    }
  }, [status, session, router]);

  // Show loading while redirecting
  return (
    <div className="h-screen w-full flex flex-col items-center justify-center bg-gradient-to-br from-muted/50 via-background to-muted/30">
      <Loader2 className="h-8 w-8 animate-spin text-primary mb-4" />
      <p className="text-sm text-muted-foreground">Redirecting...</p>
    </div>
  );
}
