import type { Metadata } from 'next';

export const metadata: Metadata = {
    title: 'InsightEngine Embed',
    description: 'Embedded Analytics View',
};

export default function EmbedLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <div className="min-h-screen bg-transparent">
            {children}
        </div>
    );
}
