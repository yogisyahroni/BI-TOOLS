import { Slide } from '@/types/presentation';
import { SlideView } from '@/components/presentation/slide-view';
import { Button } from '@/components/ui/button';
import { Wand2 } from 'lucide-react';

interface StoryCanvasProps {
    currentSlide?: Slide;
    chartComponent?: React.ReactNode;
    isGenerating: boolean;
    onGenerateOpen: () => void;
    dbId?: string | null;
}

export function StoryCanvas({ currentSlide, chartComponent, isGenerating, onGenerateOpen, dbId }: StoryCanvasProps) {
    return (
        <main className="flex-1 flex flex-col min-w-0 overflow-hidden bg-background">
            <div className="flex-1 bg-muted/20 p-8 flex items-center justify-center overflow-auto relative h-full">
                {currentSlide ? (
                    <div className="w-full max-w-4xl aspect-[16/9] shadow-2xl rounded-xl overflow-hidden bg-white border border-border/50 transition-all duration-300 transform hover:scale-[1.01]">
                        <SlideView
                            slide={currentSlide}
                            chartComponent={chartComponent}
                        />
                    </div>
                ) : (
                    <div className="text-center text-muted-foreground max-w-md">
                        <div className="mb-4 flex justify-center">
                            <Wand2 className="h-12 w-12 text-primary/20" />
                        </div>
                        <h3 className="text-xl font-semibold mb-2">Ready to tell your story?</h3>
                        <p className="mb-6">
                            Generate a comprehensive presentation from your dashboard data using AI, or start from scratch.
                        </p>
                        <Button onClick={onGenerateOpen} disabled={isGenerating || !dbId}>
                            {isGenerating ? 'Generating...' : 'Generate with AI'}
                        </Button>
                    </div>
                )}
            </div>
        </main>
    );
}
