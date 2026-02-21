"use client";

import React from "react";
import { ErrorBoundary as ReactErrorBoundary, FallbackProps } from "react-error-boundary";
import { AlertTriangle, RefreshCcw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export function ErrorFallback({ error, resetErrorBoundary }: FallbackProps) {
  const errorMessage = error instanceof Error ? error.message : String(error);
  return (
    <Card className="w-full h-full flex flex-col justify-center items-center p-4 border-destructive/50 bg-destructive/5 shadow-none overflow-hidden">
      <CardHeader className="p-0 mb-4 text-center">
        <CardTitle className="flex justify-center items-center text-destructive text-sm font-semibold">
          <AlertTriangle className="mr-2 h-4 w-4" />
          Widget Error
        </CardTitle>
        <CardDescription className="text-xs max-w-[250px] truncate" title={errorMessage}>
          {errorMessage}
        </CardDescription>
      </CardHeader>
      <CardContent className="p-0">
        <Button onClick={resetErrorBoundary} variant="outline" size="sm" className="h-8">
          <RefreshCcw className="mr-2 h-3 w-3" />
          Try Again
        </Button>
      </CardContent>
    </Card>
  );
}

export function ErrorBoundary({
  children,
  fallback,
}: {
  children: React.ReactNode;
  fallback?: React.ComponentType<FallbackProps>;
}) {
  return (
    <ReactErrorBoundary FallbackComponent={fallback || ErrorFallback}>
      {children}
    </ReactErrorBoundary>
  );
}
