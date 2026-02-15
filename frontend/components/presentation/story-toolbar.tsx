import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Loader2, Play, Download, Sparkles, Share2, Save } from 'lucide-react';

interface StoryToolbarProps {
    storyTitle: string;
    onTitleChange: (title: string) => void;
    slideCount: number;
    currentSlideIndex: number;
    isGenerating: boolean;
    onGenerate: () => void;
    onExport: () => void;
    onSave?: () => void;
    onShare?: () => void;
    onPresent?: () => void;
}

export function StoryToolbar({
    storyTitle,
    onTitleChange,
    slideCount,
    currentSlideIndex,
    isGenerating,
    onGenerate,
    onExport,
    onSave,
    onShare,
    onPresent
}: StoryToolbarProps) {
    return (
        <header className="h-16 border-b flex items-center justify-between px-6 bg-background shrink-0">
            <div className="flex items-center gap-4 flex-1">
                <Input
                    value={storyTitle}
                    onChange={(e) => onTitleChange(e.target.value)}
                    className="font-semibold text-lg border-none hover:bg-muted/50 focus-visible:ring-0 w-auto min-w-[200px]"
                />
                <span className="text-muted-foreground text-sm">
                    {slideCount > 0 ? `Slide ${currentSlideIndex + 1} of ${slideCount}` : 'Empty'}
                </span>
            </div>

            <div className="flex items-center gap-2">
                <Button
                    variant="default"
                    onClick={onGenerate}
                    disabled={isGenerating}
                    className="gap-2 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white border-0"
                >
                    {isGenerating ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
                    AI Generate Story
                </Button>
                <Button variant="outline" size="icon" title="Present" onClick={onPresent}>
                    <Play className="h-4 w-4" />
                </Button>
                <Button variant="outline" onClick={onExport} disabled={slideCount === 0} className="gap-2">
                    <Download className="h-4 w-4" />
                    Export
                </Button>
                <Button variant="outline" size="icon" title="Share" onClick={onShare}>
                    <Share2 className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="icon" title="Save" onClick={onSave}>
                    <Save className="h-4 w-4" />
                </Button>
            </div>
        </header>
    );
}
