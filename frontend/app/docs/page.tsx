import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { ArrowRight, BookOpen, Code, _Layers } from 'lucide-react';

export default function DocsPage() {
    return (
        <div className="space-y-10">
            <div className="space-y-4">
                <h1 className="scroll-m-20 text-4xl font-extrabold tracking-tight lg:text-5xl">
                    Documentation
                </h1>
                <p className="text-xl text-muted-foreground">
                    Welcome to the InsightEngine documentation. Learn how to build dashboards, connect data sources, and manage your analytics platform.
                </p>
                <div className="flex gap-4">
                    <Button asChild>
                        <Link href="/docs/quick-start">
                            Get Started <ArrowRight className="ml-2 h-4 w-4" />
                        </Link>
                    </Button>
                    <Button variant="outline" asChild>
                        <Link href="/docs/api">
                            API Reference
                        </Link>
                    </Button>
                </div>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                <div className="rounded-lg border bg-card p-6 shadow-sm">
                    <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                        <BookOpen className="h-6 w-6 text-primary" />
                    </div>
                    <h3 className="mb-2 text-xl font-bold">Guides</h3>
                    <p className="mb-4 text-muted-foreground">
                        Detailed guides on how to use InsightEngine&apos;s features, including dashboards, reporting, and alerts.
                    </p>
                    <Link href="/docs/dashboards" className="text-primary hover:underline">
                        Explore Guides &rarr;
                    </Link>
                </div>

                <div className="rounded-lg border bg-card p-6 shadow-sm">
                    <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
                        <Code className="h-6 w-6 text-primary" />
                    </div>
                    <h3 className="mb-2 text-xl font-bold">API Reference</h3>
                    <p className="mb-4 text-muted-foreground">
                        Comprehensive API documentation for developers integrating with InsightEngine.
                    </p>
                    <Link href="/docs/api" className="text-primary hover:underline">
                        View API Docs &rarr;
                    </Link>
                </div>
            </div>
        </div>
    );
}
