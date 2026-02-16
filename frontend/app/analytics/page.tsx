'use client';

export const dynamic = 'force-dynamic';

import { AnalyticsView } from "@/components/analytics/analytics-view"
import { PageLayout } from "@/components/page-layout"
import { PageHeader, PageContent } from "@/components/page-header"
import { BarChart3 } from "lucide-react"

export default function AnalyticsPage() {
    return (
        <PageLayout>
            <PageHeader
                title="Analytics"
                description="AI-powered insights and predictive analytics"
                icon={BarChart3}
            />
            <PageContent>
                <AnalyticsView />
            </PageContent>
        </PageLayout>
    )
}
