"use client";

export const dynamic = "force-dynamic";

import { AnalyticsView } from "@/components/analytics/analytics-view";
import { PageLayout } from "@/components/page-layout";
import { PageHeader, PageContent } from "@/components/page-header";
import { BarChart3 } from "lucide-react";

import { motion } from "framer-motion";

export default function AnalyticsPage() {
  return (
    <PageLayout>
      <PageHeader
        title="Analytics"
        description="AI-powered insights and predictive analytics"
        icon={BarChart3}
      />
      <PageContent>
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4, ease: "easeOut" }}
        >
          <AnalyticsView />
        </motion.div>
      </PageContent>
    </PageLayout>
  );
}
