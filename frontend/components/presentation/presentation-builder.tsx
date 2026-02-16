"use client";

import React, { useState } from 'react';
import { type SlideDeck } from '@/types/presentation';
import { presentationApi } from '@/lib/api/presentation';
import { exportToPPT } from '@/lib/export-ppt';
import { SlideView } from './slide-view';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Carousel, CarouselContent, CarouselItem, CarouselNext, CarouselPrevious } from '@/components/ui/carousel';
import { Loader2, Download, Wand2 } from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';

interface PresentationBuilderProps {
    dashboardId: string;
    dashboardName: string;
}

export function PresentationBuilder({ dashboardId, dashboardName }: PresentationBuilderProps) {
    const [prompt, setPrompt] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [deck, setDeck] = useState<SlideDeck | null>(null);
    const { toast } = useToast();

    const handleGenerate = async () => {
        if (!prompt.trim()) return;

        setIsLoading(true);
        try {
            const result = await presentationApi.generate(dashboardId, prompt);
            setDeck(result);
            toast({
                title: "Presentation Generated",
                description: "Your slides are ready to review.",
            });
        } catch (error) {
            console.error(error);
            toast({
                title: "Generation Failed",
                description: "Could not generate presentation. Please try again.",
                variant: "destructive",
            });
        } finally {
            setIsLoading(false);
        }
    };

    const handleExport = async () => {
        if (!deck) return;
        try {
            await exportToPPT(deck);
            toast({
                title: "Export Successful",
                description: "PowerPoint file downloaded.",
            });
        } catch (error) {
            console.error(error);
            toast({
                title: "Export Failed",
                description: "Could not export to PowerPoint.",
                variant: "destructive",
            });
        }
    };

    return (
        <div className="flex flex-col space-y-6 h-full">
            <Card>
                <CardHeader>
                    <CardTitle>AI Presentation Builder</CardTitle>
                    <CardDescription>Generate a slide deck from {dashboardName} insights.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex flex-col gap-2">
                        <Textarea
                            placeholder="Describe the focus of your presentation (e.g., 'Analyze sales performance Q3 vs Q4')..."
                            value={prompt}
                            onChange={(e) => setPrompt(e.target.value)}
                            rows={3}
                        />
                        <div className="flex justify-end gap-2">
                            <Button onClick={handleGenerate} disabled={isLoading || !prompt.trim()}>
                                {isLoading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Wand2 className="mr-2 h-4 w-4" />}
                                Generate Slides
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {deck && (
                <div className="flex-grow flex flex-col space-y-4">
                    <div className="flex justify-between items-center bg-muted/40 p-4 rounded-lg">
                        <div>
                            <h3 className="text-xl font-bold">{deck.title}</h3>
                            <p className="text-sm text-muted-foreground">{deck.description}</p>
                        </div>
                        <Button variant="outline" onClick={handleExport}>
                            <Download className="mr-2 h-4 w-4" />
                            Export PPTX
                        </Button>
                    </div>

                    <div className="flex-grow flex justify-center items-center bg-muted/10 rounded-lg p-8">
                        <Carousel className="w-full max-w-4xl">
                            <CarouselContent>
                                {deck.slides.map((slide, index) => (
                                    <CarouselItem key={index}>
                                        <div className="p-1">
                                            <div className="aspect-video bg-white text-black rounded-lg shadow-xl overflow-hidden">
                                                {/* Enforce 16:9 Aspect Ratio Container for Slide View */}
                                                <div className="w-full h-full">
                                                    <SlideView slide={slide} />
                                                </div>
                                            </div>
                                            <div className="text-center mt-2 text-sm text-muted-foreground">
                                                Slide {index + 1} of {deck.slides.length}
                                            </div>
                                        </div>
                                    </CarouselItem>
                                ))}
                            </CarouselContent>
                            <CarouselPrevious />
                            <CarouselNext />
                        </Carousel>
                    </div>
                </div>
            )}
        </div>
    );
}
