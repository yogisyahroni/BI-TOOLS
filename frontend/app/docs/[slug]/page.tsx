import { getDocBySlug } from '@/lib/docs';
import ReactMarkdown from 'react-markdown';
import { notFound } from 'next/navigation';
import Link from 'next/link';

interface DocPageProps {
    params: {
        slug: string;
    };
}

export default async function DocPage({ params }: DocPageProps) {
    const doc = getDocBySlug(params.slug);

    if (!doc) {
        notFound();
    }

    return (
        <article className="prose prose-slate max-w-none dark:prose-invert">
            <div className="mb-4">
                <Link href="/docs" className="text-sm text-muted-foreground hover:underline">
                    &larr; Back to Docs
                </Link>
            </div>
            <h1>{doc.frontmatter.title}</h1>
            <p className="lead text-xl text-muted-foreground">{doc.frontmatter.description}</p>
            <hr className="my-6" />
            <ReactMarkdown>{doc.content}</ReactMarkdown>
        </article>
    );
}

export async function generateStaticParams() {
    // In a real app, you'd list all files. 
    // For now, we rely on dynamic rendering or just let Next.js find them.
    // If output is 'export', we need this. For 'standalone', we might not strictly need it 
    // but it's good practice for SSG.
    return [
        { slug: 'quick-start' },
        { slug: 'dashboards' },
        { slug: 'data-sources' },
        { slug: 'alerts' },
    ];
}
