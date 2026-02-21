"use client";

import { SidebarLayout } from "@/components/sidebar-layout";
import { StoryBuilder } from "@/components/story-builder/StoryBuilder";
import { motion } from "framer-motion";

export default function StoriesPage() {
  return (
    <SidebarLayout>
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: "easeOut" }}
        style={{ height: "100%", width: "100%" }}
      >
        <StoryBuilder />
      </motion.div>
    </SidebarLayout>
  );
}
