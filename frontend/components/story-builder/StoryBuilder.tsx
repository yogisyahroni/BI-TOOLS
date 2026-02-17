"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/components/ui/use-toast";
import { storyService } from "@/services/storyService";
import { type Story, type Slide } from "@/types/story";
import { Loader2, Plus, Save, Download, Trash2 } from "lucide-react";
import { useState, useEffect } from "react";

export function StoryBuilder() {
    const [stories, setStories] = useState<Story[]>([]);
    const [selectedStory, setSelectedStory] = useState<Story | null>(null);
    const [isLoading, setIsLoading] = useState(false);
    const [prompt, setPrompt] = useState("");
    const { toast } = useToast();

    useEffect(() => {
        loadStories();
        // eslint-disable-next-line react-hooks/exhaustive-deps
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    const loadStories = async () => {
        try {
            setIsLoading(true);
            const data = await storyService.getStories();
            setStories(data);
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to load stories",
                variant: "destructive",
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleCreateStory = async () => {
        if (!prompt) return;

        try {
            setIsLoading(true);
            // For now, passing a dummy dashboard ID. In a real app, user would select a dashboard.
            const newStory = await storyService.createStory({
                dashboard_id: "dashboard-123",
                prompt: prompt,
            });
            setStories([newStory, ...stories]);
            setSelectedStory(newStory);
            setPrompt("");
            toast({
                title: "Success",
                description: "Story created successfully",
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            });
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (error: any) {
            toast({
                title: "Error",
                description: error.message || "Failed to create story",
                variant: "destructive",
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleUpdateSlide = (index: number, field: keyof Slide, value: string) => {
        if (!selectedStory) return;

        const updatedSlides = [...selectedStory.content.slides];
        updatedSlides[index] = { ...updatedSlides[index], [field]: value };

        setSelectedStory({
            ...selectedStory,
            content: { ...selectedStory.content, slides: updatedSlides },
        });
    };

    const handleSaveStory = async () => {
        if (!selectedStory) return;

        try {
            setIsLoading(true);
            await storyService.updateStory(selectedStory.id, {
                title: selectedStory.title,
                description: selectedStory.description,
                content: selectedStory.content,
            });
            toast({
                title: "Success",
                description: "Story saved successfully",
            });
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to save story",
                variant: "destructive",
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleExportPPTX = async () => {
        if (!selectedStory) return;

        try {
            setIsLoading(true);
            await storyService.exportPPTX(selectedStory.id, selectedStory.title);
            toast({
                title: "Success",
                description: "Export started",
            });
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to export PPTX",
                variant: "destructive",
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleDeleteStory = async (id: string) => {
        try {
            setIsLoading(true);
            await storyService.deleteStory(id);
            setStories(stories.filter(s => s.id !== id));
            if (selectedStory?.id === id) {
                setSelectedStory(null);
            }
            toast({
                title: "Success",
                description: "Story deleted",
            })
        } catch (_error) {
            toast({
                title: "Error",
                description: "Failed to delete story",
                variant: "destructive"
            })
        } finally {
            setIsLoading(false);
        }
    }

    return (
        <div className="container mx-auto p-6 flex h-full gap-6">
            {/* Sidebar: List of Stories */}
            <div className="w-1/4 border-r pr-6 flex flex-col gap-4">
                <h2 className="text-2xl font-bold">My Stories</h2>
                <div className="flex gap-2">
                    <Input
                        placeholder="Enter prompt to generate..."
                        value={prompt}
                        onChange={(e) => setPrompt(e.target.value)}
                    />
                    <Button onClick={handleCreateStory} disabled={isLoading || !prompt} size="icon">
                        <Plus className="h-4 w-4" />
                    </Button>
                </div>

                <div className="flex-1 overflow-y-auto space-y-2">
                    {isLoading && stories.length === 0 && <div className="text-center p-4"><Loader2 className="animate-spin h-6 w-6 mx-auto" /></div>}

                    {stories.map(story => (
                        <div
                            key={story.id}
                            className={`p-3 border rounded-lg cursor-pointer flex justify-between items-center group ${selectedStory?.id === story.id ? 'bg-accent/50 border-primary' : 'hover:bg-accent/20'}`}
                            onClick={() => setSelectedStory(story)}
                        >
                            <div className="truncate font-medium">{story.title}</div>
                            <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity text-destructive"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    handleDeleteStory(story.id);
                                }}
                            >
                                <Trash2 className="h-3 w-3" />
                            </Button>
                        </div>
                    ))}
                </div>
            </div>

            {/* Main Content: Story Editor */}
            <div className="flex-1 flex flex-col gap-6 overflow-hidden">
                {selectedStory ? (
                    <>
                        <div className="flex justify-between items-center">
                            <div>
                                <Label htmlFor="story-title" className="text-sm text-muted-foreground">Title</Label>
                                <Input
                                    id="story-title"
                                    value={selectedStory.title}
                                    onChange={(e) => setSelectedStory({ ...selectedStory, title: e.target.value })}
                                    className="text-xl font-bold h-auto py-2 px-0 border-0 border-b rounded-none focus-visible:ring-0 px-1"
                                />
                            </div>
                            <div className="flex gap-2">
                                <Button onClick={handleSaveStory} disabled={isLoading}>
                                    <Save className="mr-2 h-4 w-4" /> Save
                                </Button>
                                <Button onClick={handleExportPPTX} disabled={isLoading} variant="outline">
                                    <Download className="mr-2 h-4 w-4" /> Export PPTX
                                </Button>
                            </div>
                        </div>

                        <div className="flex-1 overflow-y-auto space-y-6 pr-4">
                            {selectedStory.content.slides.map((slide, index) => (
                                <Card key={index} className="relative group">
                                    <CardContent className="p-6 space-y-4">
                                        <div className="flex justify-between items-center">
                                            <span className="text-sm font-mono text-muted-foreground">Slide {index + 1}</span>
                                            <span className="text-xs bg-secondary px-2 py-1 rounded capitalize">{slide.layout.replace('_', ' ')}</span>
                                        </div>

                                        <div className="space-y-2">
                                            <Label>Title</Label>
                                            <Input
                                                value={slide.title}
                                                onChange={(e) => handleUpdateSlide(index, 'title', e.target.value)}
                                            />
                                        </div>

                                        <div className="space-y-2">
                                            <Label>Content</Label>
                                            <Textarea
                                                value={slide.content}
                                                onChange={(e) => handleUpdateSlide(index, 'content', e.target.value)}
                                                rows={5}
                                            />
                                        </div>

                                        <div className="space-y-2">
                                            <Label>Speaker Notes</Label>
                                            <Textarea
                                                value={slide.notes || ''}
                                                onChange={(e) => handleUpdateSlide(index, 'notes', e.target.value)}
                                                rows={2}
                                                className="text-muted-foreground italic text-sm"
                                            />
                                        </div>
                                    </CardContent>
                                </Card>
                            ))}
                        </div>
                    </>
                ) : (
                    <div className="flex-1 flex justify-center items-center text-muted-foreground flex-col gap-4">
                        <Loader2 className="h-12 w-12 opacity-20" />
                        <p>Select a story or create a new one to get started</p>
                    </div>
                )}
            </div>
        </div>
    );
}
