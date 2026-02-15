"use client";

import React from 'react';
import { PresentationBuilder } from '@/components/presentation/presentation-builder';
import { useParams } from 'next/navigation';

export default function PresentationPage() {
    const params = useParams();
    const dashboardId = typeof params.id === 'string' ? params.id : Array.isArray(params.id) ? params.id[0] : '';

    // In a real app we'd fetch dashboard name here, for now using ID or generic
    const dashboardName = "Dashboard Analysis";

    return (
        <div className="container mx-auto py-8 h-[calc(100vh-4rem)]">
            <PresentationBuilder dashboardId={dashboardId} dashboardName={dashboardName} />
        </div>
    );
}
