import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Loader2, Wand2 } from 'lucide-react';
import { AIModelSelector } from '@/components/ai-model-selector';
import { type AIModel } from '@/lib/ai/registry';

interface StoryAIDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    prompt: string;
    onPromptChange: (prompt: string) => void;
    selectedModel: AIModel;
    onModelSelect: (model: AIModel) => void;
    isGenerating: boolean;
    onGenerate: () => void;
}

export function StoryAIDialog({
    open,
    onOpenChange,
    prompt,
    onPromptChange,
    selectedModel,
    onModelSelect,
    isGenerating,
    onGenerate
}: StoryAIDialogProps) {
    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[700px]">
                <DialogHeader>
                    <DialogTitle>Generate Data Story</DialogTitle>
                    <DialogDescription>
                        Create a presentation based on your dashboard data.
                    </DialogDescription>
                </DialogHeader>

                <div className="py-6 space-y-4">
                    <div className="space-y-2">
                        <label className="text-sm font-medium text-foreground">
                            Ask in Natural Language
                        </label>
                        <div className="flex flex-col sm:flex-row gap-2">
                            <Input
                                placeholder="e.g., Show me top 5 customers by total sales last month"
                                value={prompt}
                                onChange={(e) => onPromptChange(e.target.value)}
                                className="flex-1"
                                autoFocus
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter') {
                                        onGenerate();
                                    }
                                }}
                            />
                            <div className="flex gap-2">
                                <AIModelSelector
                                    selectedModel={selectedModel}
                                    onSelect={onModelSelect}
                                />
                                <Button
                                    onClick={onGenerate}
                                    disabled={isGenerating || !prompt.trim()}
                                    className="gap-2"
                                >
                                    {isGenerating ? <Loader2 className="h-4 w-4 animate-spin" /> : <Wand2 className="h-4 w-4" />}
                                    Generate
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}
