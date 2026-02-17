'use client';

interface VideoPlayerProps {
    src: string;
    title: string;
    poster?: string;
}

export function VideoPlayer({ src, title, _poster }: VideoPlayerProps) {
    return (
        <div className="overflow-hidden rounded-lg border bg-black shadow-sm">
            <div className="relative aspect-video">
                <iframe
                    src={src}
                    title={title}
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                    allowFullScreen
                    className="absolute top-0 left-0 h-full w-full border-0"
                />
            </div>
            <div className="p-4">
                <h3 className="font-semibold text-white">{title}</h3>
            </div>
        </div>
    );
}
