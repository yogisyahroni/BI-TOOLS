import { VideoPlayer } from '@/components/ui/video-player';
import { Button } from '@/components/ui/button';
import Link from 'next/link';

const tutorials = [
    {
        title: 'Building Your First Dashboard',
        src: 'https://www.youtube.com/embed/dQw4w9WgXcQ', // Placeholder: Rick Roll (Standard Test Video)
        category: 'Getting Started',
    },
    {
        title: 'Connecting a PostgreSQL Database',
        src: 'https://www.youtube.com/embed/dQw4w9WgXcQ',
        category: 'Data Sources',
    },
    {
        title: 'Setting Up Alerts',
        src: 'https://www.youtube.com/embed/dQw4w9WgXcQ',
        category: 'Alerts',
    },
    {
        title: 'Advanced Filtering Techniques',
        src: 'https://www.youtube.com/embed/dQw4w9WgXcQ',
        category: 'Advanced',
    },
];

export default function TutorialsPage() {
    return (
        <div className="space-y-8">
            <div className="mb-4">
                <Link href="/docs" className="text-sm text-muted-foreground hover:underline">
                    &larr; Back to Docs
                </Link>
            </div>

            <div className="space-y-2">
                <h1 className="text-3xl font-bold">Video Tutorials</h1>
                <p className="text-muted-foreground">
                    Watch step-by-step guides to master InsightEngine.
                </p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {tutorials.map((tutorial, index) => (
                    <VideoPlayer
                        key={index}
                        title={tutorial.title}
                        src={tutorial.src}
                    />
                ))}
            </div>

            <div className="rounded-lg border bg-muted p-8 text-center">
                <h3 className="mb-2 text-lg font-semibold">Need more help?</h3>
                <p className="mb-4 text-sm text-muted-foreground">
                    Check out our written guides or contact support.
                </p>
                <Button variant="outline" asChild>
                    <Link href="/docs">Browse Guides</Link>
                </Button>
            </div>
        </div>
    );
}
