'use client';

export const dynamic = 'force-dynamic';

import { redirect } from 'next/navigation';

export default function DashboardRootPage() {
  redirect('/dashboards');
}
