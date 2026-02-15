import { Slide } from '@/types/presentation';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Plus, Trash2 } from 'lucide-react';

interface StorySidebarProps {
    slides: Slide[];
    currentSlideIndex: number;
    onChangeSlide: (index: number) => void;
    onAddSlide: () => void;
    onDeleteSlide?: (index: number) => void;
}

export function StorySidebar({ slides, currentSlideIndex, onChangeSlide, onAddSlide, onDeleteSlide }: StorySidebarProps) {
    return (
        <aside className="w-72 border-r flex flex-col bg-muted/30 h-full">
            <div className="p-4 border-b">
                <h2 className="font-semibold text-sm text-muted-foreground uppercase tracking-wider">Slides</h2>
            </div>
            <ScrollArea className="flex-1 p-4">
                <div className="space-y-3">
                    {slides.map((slide, index) => (
                        <div
                            key={index}
                            onClick={() => onChangeSlide(index)}
                            className={`
                                cursor-pointer rounded-lg border p-3 transition-all hover:shadow-md group relative
                                ${currentSlideIndex === index ? 'ring-2 ring-primary border-primary bg-background' : 'bg-card border-border hover:border-primary/50'}
                            `}
                        >
                            <div className="flex items-center justify-between mb-2">
                                <span className="text-xs font-medium text-muted-foreground">Slide {index + 1}</span>
                                {currentSlideIndex === index && <div className="h-2 w-2 rounded-full bg-primary" />}
                            </div>
                            <div className="h-16 bg-muted rounded mb-2 overflow-hidden relative flex items-center justify-center">
                                {/* Mini Preview Placeholder */}
                                <div className="absolute inset-0 flex items-center justify-center text-[8px] text-muted-foreground text-center p-1 leading-tight pointer-events-none">
                                    {slide.title}
                                </div>
                            </div>
                            <p className="text-xs font-medium truncate">{slide.title}</p>

                            {onDeleteSlide && (
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="absolute top-1 right-1 h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-destructive/10 hover:text-destructive"
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        onDeleteSlide(index);
                                    }}
                                >
                                    <Trash2 className="h-3 w-3" />
                                </Button>
                            )}
                        </div>
                    ))}
                </div>
                {slides.length === 0 && (
                    <div className="text-center py-10 text-muted-foreground text-sm">
                        <p>No slides yet.</p>
                        <Button variant="link" onClick={onAddSlide} className="mt-2 text-primary">Add your first slide</Button>
                    </div>
                )}
            </ScrollArea>
            <div className="p-4 border-t bg-background">
                <Button variant="outline" className="w-full justify-start gap-2" onClick={onAddSlide}>
                    <Plus className="h-4 w-4" />
                    Add Slide
                </Button>
            </div>
        </aside>
    );
}
