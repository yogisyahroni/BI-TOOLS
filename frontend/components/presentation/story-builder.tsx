'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useStoryGeneration } from '@/hooks/use-story-generation';
import { useDashboard } from '@/hooks/use-dashboard';
import { DashboardCard } from '@/components/dashboard/dashboard-card';
import { DEFAULT_AI_MODEL, AIModel } from '@/lib/ai/registry';

// Atomic Components
import { StorySidebar } from '@/components/presentation/story-sidebar';
import { StoryToolbar } from '@/components/presentation/story-toolbar';
import { StoryCanvas } from '@/components/presentation/story-canvas';
import { StoryAIDialog } from '@/components/presentation/story-ai-dialog';
import { Slide } from '@/types/presentation';

interface StoryBuilderProps {
    dashboardId?: string;
    initialSlides?: Slide[];
}

export function StoryBuilder({ dashboardId, initialSlides }: StoryBuilderProps) {
    const router = useRouter();
    const searchParams = useSearchParams();
    const dbId = dashboardId || searchParams.get('dashboardId');

    // Custom Hook for State & Logic
    const {
        slides,
        currentSlideIndex,
        isGenerating,
        storyTitle,
        setStoryTitle,
        setCurrentSlideIndex,
        generateStory,
        exportStory,
        addSlide,
        deleteSlide
    } = useStoryGeneration({ dashboardId: dbId || undefined, initialSlides });

    // AI Generation State (Local UI State)
    const [showAiDialog, setShowAiDialog] = useState(false);
    const [aiPrompt, setAiPrompt] = useState('');
    const [selectedModel, setSelectedModel] = useState<AIModel>(DEFAULT_AI_MODEL);

    // Fetch Dashboard Data for Charts
    const { dashboard, isLoading: isDashboardLoading } = useDashboard(dbId || '');

    // Generate initial slides if none exist and we have a dashboard ID
    useEffect(() => {
        if (!slides.length && dbId && !isGenerating) {
            // Optional: Auto-generate on load logic here if needed
        }
    }, [dbId, slides.length, isGenerating]);

    const handleGenerate = async () => {
        const success = await generateStory(aiPrompt, selectedModel.providerId);
        if (success) {
            setShowAiDialog(false);
        }
    };

    const currentSlide = slides[currentSlideIndex];

    // Resolve Chart Component if slide has chart_id
    const renderChart = () => {
        if (!currentSlide?.chart_id || !dashboard) return null;

        const card = dashboard.cards?.find(c => c.id === currentSlide.chart_id);
        if (!card) return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground bg-muted/10 rounded-lg border border-dashed">
                <p>Chart not found ({currentSlide.chart_id})</p>
            </div>
        );

        return (
            <div className="h-full w-full pointer-events-none transform scale-95 origin-top-left">
                <DashboardCard
                    card={card}
                    isEditing={false}
                />
            </div>
        );
    };

    return (
        <div className="flex h-[calc(100vh-4rem)] overflow-hidden bg-background">
            {/* Left Sidebar */}
            <StorySidebar
                slides={slides}
                currentSlideIndex={currentSlideIndex}
                onChangeSlide={setCurrentSlideIndex}
                onAddSlide={addSlide}
                onDeleteSlide={deleteSlide}
            />

            {/* Main Content */}
            <div className="flex-1 flex flex-col min-w-0">
                <StoryToolbar
                    storyTitle={storyTitle}
                    onTitleChange={setStoryTitle}
                    slideCount={slides.length}
                    currentSlideIndex={currentSlideIndex}
                    isGenerating={isGenerating}
                    onGenerate={() => setShowAiDialog(true)}
                    onExport={exportStory}
                    onSave={() => { }} // Placeholder for Save
                    onShare={() => { }} // Placeholder for Share
                    onPresent={() => { }} // Placeholder for Present
                />

                <StoryCanvas
                    currentSlide={currentSlide}
                    chartComponent={renderChart()}
                    isGenerating={isGenerating}
                    onGenerateOpen={() => setShowAiDialog(true)}
                    dbId={dbId}
                />
            </div>

            {/* AI Dialog */}
            <StoryAIDialog
                open={showAiDialog}
                onOpenChange={setShowAiDialog}
                prompt={aiPrompt}
                onPromptChange={setAiPrompt}
                selectedModel={selectedModel}
                onModelSelect={setSelectedModel}
                isGenerating={isGenerating}
                onGenerate={handleGenerate}
            />
        </div>
    );
}
