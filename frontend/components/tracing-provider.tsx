"use client"

import { useEffect } from "react"
import { initializeTracing } from "@/lib/tracing"

export function TracingProvider({ children }: { children: React.ReactNode }) {
    useEffect(() => {
        // Initialize tracing only once when the app mounts
        initializeTracing()
    }, [])

    return <>{children}</>
}
