"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useToast } from "@/hooks/use-toast";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { type Story, type Slide } from "@/types/story";
import { useStories } from "@/hooks/use-stories";
import {
  Loader2,
  Plus,
  Save,
  Download,
  Trash2,
  Sparkles,
  PencilLine,
  Film,
  Share2,
  Play,
  Copy,
  Globe,
  Lock,
} from "lucide-react";
import { useState, useEffect } from "react";

// Import new 3-column components
import { SlideNavigator } from "./SlideNavigator";
import { SlideCanvas } from "./SlideCanvas";
import { SlideProperties } from "./SlideProperties";
import { ErrorBoundary } from "@/components/ui/error-boundary";

// --- Constants --------------------------------------------------------------

const DEFAULT_SLIDE: Slide = {
  title: "New Slide",
  content: "",
  layout: "bullet_points",
  notes: "",
};

type CreationMode = "ai" | "manual";

// --- Empty State -------------------------------------------------------------

function EmptyState({ onSelectMode }: { onSelectMode: (mode: CreationMode) => void }) {
  return (
    <div className="flex-1 flex flex-col items-center justify-center gap-8 p-12">
      <div className="text-center space-y-2">
        <Film className="h-12 w-12 mx-auto text-muted-foreground/40" />
        <h3 className="text-lg font-semibold">Create a Story</h3>
        <p className="text-sm text-muted-foreground max-w-xs">
          Build a presentation from your data. Choose how you&apos;d like to get started.
        </p>
      </div>
      <div className="grid grid-cols-2 gap-4 w-full max-w-lg">
        <Card
          className="cursor-pointer border-2 hover:border-primary hover:bg-primary/5 transition-all duration-200 group"
          onClick={() => onSelectMode("ai")}
        >
          <CardContent className="p-6 flex flex-col items-center gap-3 text-center">
            <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shadow-md group-hover:scale-110 transition-transform duration-200">
              <Sparkles className="h-6 w-6 text-white" />
            </div>
            <div>
              <p className="font-semibold text-sm">AI Generate</p>
              <p className="text-xs text-muted-foreground mt-1">
                Describe what you need; AI writes the slides
              </p>
            </div>
          </CardContent>
        </Card>

        <Card
          className="cursor-pointer border-2 hover:border-primary hover:bg-primary/5 transition-all duration-200 group"
          onClick={() => onSelectMode("manual")}
        >
          <CardContent className="p-6 flex flex-col items-center gap-3 text-center">
            <div className="h-12 w-12 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-md group-hover:scale-110 transition-transform duration-200">
              <PencilLine className="h-6 w-6 text-white" />
            </div>
            <div>
              <p className="font-semibold text-sm">Start Manually</p>
              <p className="text-xs text-muted-foreground mt-1">
                Build slide-by-slide with full control
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

// --- Creation Dialog ----------------------------------------------------------

function CreateStoryDialog({
  mode,
  open,
  isLoading,
  onClose,
  onCreateAI,
  onCreateManual,
}: {
  mode: CreationMode;
  open: boolean;
  isLoading: boolean;
  onClose: () => void;
  onCreateAI: (prompt: string) => void;
  onCreateManual: (title: string, description: string) => void;
}) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [prompt, setPrompt] = useState("");

  const handleSubmit = () => {
    if (mode === "ai") {
      if (!prompt.trim()) return;
      onCreateAI(prompt.trim());
    } else {
      if (!title.trim()) return;
      onCreateManual(title.trim(), description.trim());
    }
  };

  const canSubmit = mode === "ai" ? !!prompt.trim() : !!title.trim();

  return (
    <Dialog open={open} onOpenChange={() => onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {mode === "ai" ? (
              <>
                <Sparkles className="h-5 w-5 text-violet-500" />
                Generate with AI
              </>
            ) : (
              <>
                <PencilLine className="h-5 w-5 text-emerald-500" />
                New Manual Story
              </>
            )}
          </DialogTitle>
          <DialogDescription>
            {mode === "ai"
              ? "Describe the story you want. AI will generate slides for you."
              : "Give your story a title. You can add and edit slides manually."}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 pt-2">
          {mode === "ai" ? (
            <div className="space-y-2">
              <Label htmlFor="prompt">Prompt</Label>
              <Textarea
                id="prompt"
                placeholder="E.g. Quarterly sales report highlighting top performers and revenue trends..."
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                rows={4}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) handleSubmit();
                }}
              />
            </div>
          ) : (
            <>
              <div className="space-y-2">
                <Label htmlFor="story-title">Title *</Label>
                <Input
                  id="story-title"
                  placeholder="My Presentation"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  autoFocus
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleSubmit();
                  }}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="story-desc">Description (optional)</Label>
                <Input
                  id="story-desc"
                  placeholder="Brief description of this story..."
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>
            </>
          )}

          <div className="flex justify-end gap-2 pt-2">
            <Button variant="outline" onClick={onClose} disabled={isLoading}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={!canSubmit || isLoading}>
              {isLoading ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : mode === "ai" ? (
                <Sparkles className="mr-2 h-4 w-4" />
              ) : (
                <Plus className="mr-2 h-4 w-4" />
              )}
              {mode === "ai" ? "Generate" : "Create Story"}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// --- Share Dialog -------------------------------------------------------------

function ShareStoryDialog({
  story,
  open,
  onClose,
  onTogglePublic,
}: {
  story: Story | null;
  open: boolean;
  onClose: () => void;
  onTogglePublic: (isPublic: boolean) => Promise<void>;
}) {
  const { toast } = useToast();
  const [isLoading, setIsLoading] = useState(false);

  if (!story) return null;

  const publicUrl = story.share_token
    ? `${typeof window !== "undefined" ? window.location.origin : ""}/public/present/${story.share_token}`
    : "";

  const handleToggle = async () => {
    setIsLoading(true);
    try {
      await onTogglePublic(!story.is_public);
      toast({
        title: story.is_public ? "Made Private" : "Made Public",
        description: story.is_public
          ? "This story is no longer accessible via link."
          : "Anyone with the link can now view this story.",
      });
    } catch (_error) {
      toast({
        title: "Error",
        description: "Failed to update sharing settings.",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopyLink = () => {
    if (!publicUrl) return;
    navigator.clipboard
      .writeText(publicUrl)
      .then(() => {
        toast({ title: "Link copied!", description: "Public link copied to clipboard." });
      })
      .catch(() => {
        toast({
          title: "Failed to copy",
          description: "Could not copy link.",
          variant: "destructive",
        });
      });
  };

  return (
    <Dialog open={open} onOpenChange={() => onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Share2 className="h-5 w-5 text-blue-500" />
            Share Story
          </DialogTitle>
          <DialogDescription>
            Manage who can view this story. Public stories can be viewed by anyone with the link.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-6 py-4">
          <div className="flex items-center justify-between p-4 border rounded-xl bg-muted/20">
            <div className="space-y-1 pr-4">
              <h4 className="font-semibold text-sm flex items-center gap-2">
                {story.is_public ? (
                  <>
                    <Globe className="h-4 w-4 text-emerald-500" /> Public Access
                  </>
                ) : (
                  <>
                    <Lock className="h-4 w-4 text-muted-foreground" /> Private Access
                  </>
                )}
              </h4>
              <p className="text-xs text-muted-foreground">
                {story.is_public
                  ? "Anyone with the link can view this story."
                  : "Only you can view this story."}
              </p>
            </div>
            <Button
              variant={story.is_public ? "destructive" : "default"}
              size="sm"
              onClick={handleToggle}
              disabled={isLoading}
            >
              {isLoading && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              {story.is_public ? "Make Private" : "Make Public"}
            </Button>
          </div>

          {story.is_public && publicUrl && (
            <div className="space-y-2">
              <Label className="text-xs font-semibold uppercase text-muted-foreground">
                Public Link
              </Label>
              <div className="flex items-center gap-2">
                <Input
                  value={publicUrl}
                  readOnly
                  className="font-mono text-xs bg-muted/50"
                  onClick={(e) => e.currentTarget.select()}
                />
                <Button size="icon" variant="outline" onClick={handleCopyLink} title="Copy Link">
                  <Copy className="h-4 w-4" />
                </Button>
                <Button
                  size="icon"
                  variant="outline"
                  onClick={() => window.open(publicUrl, "_blank")}
                  title="Open in new tab"
                >
                  <Globe className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </div>
        <div className="flex justify-end gap-3 pt-4 border-t">
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// --- Main Component -----------------------------------------------------------

export function StoryBuilder() {
  const {
    stories,
    isLoading: isStoriesLoading,
    createAIStory,
    createManualStory,
    updateStory,
    deleteStory,
  } = useStories();

  // The selected story must be managed locally, but we should sync it with the latest
  // remote data if it updates.
  const [selectedStoryId, setSelectedStoryId] = useState<string | null>(null);
  const selectedStory = stories.find((s) => s.id === selectedStoryId) || null;

  const [activeSlideIndex, setActiveSlideIndex] = useState<number>(0);
  const [isSaving, setIsSaving] = useState(false);
  const [dialogMode, setDialogMode] = useState<CreationMode | null>(null);
  const [isShareDialogOpen, setIsShareDialogOpen] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const { toast } = useToast();

  // Sync selected story locally if it was deleted
  useEffect(() => {
    if (selectedStoryId && !stories.some((s) => s.id === selectedStoryId) && !isStoriesLoading) {
      setSelectedStoryId(null);
      setActiveSlideIndex(0);
    }
  }, [stories, selectedStoryId, isStoriesLoading]);

  // -- Create via AI -------------------------------------------------------

  const handleCreateAI = async (prompt: string) => {
    setIsProcessing(true);
    const { success, data, error } = await createAIStory(prompt);
    setIsProcessing(false);

    if (success && data) {
      setSelectedStoryId(data.id);
      setActiveSlideIndex(0);
      setDialogMode(null);
    }
  };

  // -- Create manually -----------------------------------------------------

  const handleCreateManual = async (title: string, description: string) => {
    setIsProcessing(true);
    const { success, data, error } = await createManualStory(title, description);
    setIsProcessing(false);

    if (success && data) {
      setSelectedStoryId(data.id);
      setActiveSlideIndex(0);
      setDialogMode(null);
    }
  };

  // -- Slide operations ----------------------------------------------------

  const handleAddSlide = async () => {
    if (!selectedStory) return;
    const slides = [...(selectedStory.content?.slides ?? []), { ...DEFAULT_SLIDE }];
    await updateStory(selectedStory.id, { content: { slides } });
    setActiveSlideIndex(slides.length - 1);
  };

  const handleUpdateSlide = async (field: keyof Slide, value: string) => {
    if (!selectedStory) return;
    const slides = [...selectedStory.content.slides];
    slides[activeSlideIndex] = { ...slides[activeSlideIndex], [field]: value as any };
    await updateStory(selectedStory.id, { content: { slides } });
  };

  const handleDeleteSlide = async (index: number) => {
    if (!selectedStory) return;
    const slides = selectedStory.content.slides.filter((_, i) => i !== index);
    await updateStory(selectedStory.id, { content: { slides } });
    if (activeSlideIndex >= slides.length) {
      setActiveSlideIndex(Math.max(0, slides.length - 1));
    }
  };

  const handleMoveSlide = async (index: number, direction: "up" | "down") => {
    if (!selectedStory) return;
    const slides = [...selectedStory.content.slides];
    const target = direction === "up" ? index - 1 : index + 1;
    if (target < 0 || target >= slides.length) return;
    [slides[index], slides[target]] = [slides[target], slides[index]];
    await updateStory(selectedStory.id, { content: { slides } });

    // Follow the active slide
    if (activeSlideIndex === index) {
      setActiveSlideIndex(target);
    } else if (activeSlideIndex === target) {
      setActiveSlideIndex(index);
    }
  };

  // -- Save / export / delete / present -----------------------------------

  const handleSaveStory = async () => {
    if (!selectedStory) return;
    setIsSaving(true);
    // Note: The visual UI has auto-save via handleUpdateSlide now!
    // To keep the explicit "Save" button functional, we just force an update
    await updateStory(selectedStory.id, {
      title: selectedStory.title,
      description: selectedStory.description,
      content: selectedStory.content,
    });
    setIsSaving(false);
  };

  const handleExportPPTX = async () => {
    if (!selectedStory) return;
    try {
      setIsProcessing(true);
      const { storyService } = await import("@/services/storyService");
      await storyService.exportPPTX(selectedStory.id, selectedStory.title);
      toast({ title: "Exported!", description: "PPTX download started." });
    } catch (_error) {
      toast({ title: "Error", description: "Failed to export PPTX.", variant: "destructive" });
    } finally {
      setIsProcessing(false);
    }
  };

  const handlePresent = () => {
    if (!selectedStory) return;
    window.open(`/stories/${selectedStory.id}/present`, "_blank");
  };

  const handleTogglePublic = async (isPublic: boolean) => {
    if (!selectedStory) return;
    // Since togglePublicShare has specific logic returning share_token, we use the service
    // and force update the local query manually for now.
    const { storyService } = await import("@/services/storyService");
    const result = await storyService.togglePublicShare(selectedStory.id, isPublic);
    await updateStory(selectedStory.id, {
      is_public: result.is_public,
      share_token: result.share_token,
    });
  };

  const handleDeleteStory = async (id: string, e: React.MouseEvent) => {
    e.stopPropagation();
    await deleteStory(id);
  };

  // -- Render --------------------------------------------------------------

  const slides = selectedStory?.content?.slides ?? [];
  const activeSlide = slides[activeSlideIndex];

  return (
    <div className="flex h-full overflow-hidden bg-background">
      {/* -- 1. Story Sidebar ---------------------------------------- */}
      <aside className="w-[300px] border-r flex flex-col bg-background shrink-0 z-10 shadow-[2px_0_8px_-4px_rgba(0,0,0,0.1)]">
        <div className="p-5 border-b space-y-4">
          <h2 className="text-xl font-bold tracking-tight">My Stories</h2>

          <Tabs defaultValue="manual" className="w-full">
            <TabsList className="w-full grid grid-cols-2 h-9 p-1">
              <TabsTrigger
                value="ai"
                className="text-xs font-semibold"
                onClick={() => setDialogMode("ai")}
              >
                <Sparkles className="h-3.5 w-3.5 mr-1.5 text-violet-500" />
                AI Content
              </TabsTrigger>
              <TabsTrigger
                value="manual"
                className="text-xs font-semibold"
                onClick={() => setDialogMode("manual")}
              >
                <PencilLine className="h-3.5 w-3.5 mr-1.5 text-emerald-500" />
                Build Manual
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <div className="flex-1 overflow-y-auto p-4 space-y-2">
          {isStoriesLoading && stories.length === 0 && (
            <div className="flex justify-center items-center h-24">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
            </div>
          )}
          {!isStoriesLoading && stories.length === 0 && (
            <p className="text-sm text-muted-foreground text-center py-8">No stories yet.</p>
          )}
          {stories.map((story) => (
            <div
              key={story.id}
              className={`group flex items-center justify-between gap-2 p-3 rounded-lg cursor-pointer transition-all duration-200 border-2 ${
                selectedStory?.id === story.id
                  ? "bg-primary/5 border-primary shadow-sm"
                  : "hover:bg-accent hover:border-border border-transparent"
              }`}
              onClick={() => {
                setSelectedStoryId(story.id);
                setActiveSlideIndex(0);
              }}
            >
              <div className="min-w-0">
                <p className="text-sm font-semibold truncate leading-tight mb-1 group-hover:text-primary transition-colors">
                  {story.title}
                </p>
                <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">
                  {story.content?.slides?.length ?? 0} slide
                  {(story.content?.slides?.length ?? 0) !== 1 ? "s" : ""}
                </p>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity text-destructive hover:bg-destructive/10 shrink-0"
                onClick={(e) => handleDeleteStory(story.id, e)}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      </aside>

      {/* -- 2. Main Studio Area --------------------------------------- */}
      <main className="flex-1 flex flex-col min-w-0 relative bg-muted/20">
        {selectedStory ? (
          <>
            {/* Top toolbar */}
            <div className="h-16 px-6 border-b flex items-center gap-4 bg-background shrink-0 z-10">
              <div className="flex-1 min-w-0">
                <Input
                  value={selectedStory.title}
                  onChange={(e) => updateStory(selectedStory.id, { title: e.target.value })}
                  className="text-lg font-bold h-10 py-1 px-3 border-transparent hover:border-border focus-visible:ring-1 bg-transparent transition-all"
                  placeholder="Story title..."
                />
              </div>
              <div className="flex items-center gap-3 shrink-0">
                <Badge variant="secondary" className="px-3 h-7 font-semibold">
                  {slides.length} slide{slides.length !== 1 ? "s" : ""}
                </Badge>
                <div className="h-6 w-px bg-border mx-1"></div>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => setIsShareDialogOpen(true)}
                  disabled={slides.length === 0}
                  className="h-9 px-3 font-semibold transition-all hover:bg-accent"
                  title="Share Story"
                >
                  <Share2 className="h-4 w-4 mr-2 text-blue-500" />
                  Share
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handlePresent}
                  disabled={slides.length === 0}
                  className="h-9 px-3 font-semibold shadow-sm hover:bg-accent transition-all"
                >
                  <Play className="h-4 w-4 mr-2 text-emerald-500" />
                  Present
                </Button>
                <div className="h-6 w-px bg-border mx-1"></div>
                <Button
                  size="sm"
                  onClick={handleSaveStory}
                  disabled={isSaving}
                  className="h-9 px-4 font-semibold shadow-sm transition-all hover:scale-[1.02] active:scale-[0.98]"
                >
                  {isSaving ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <Save className="h-4 w-4 mr-2" />
                  )}
                  Save
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={handleExportPPTX}
                  disabled={isProcessing || slides.length === 0}
                  className="h-9 px-4 font-semibold shadow-sm hover:bg-accent transition-all"
                >
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
              </div>
            </div>

            {/* Studio Workspace */}
            <div className="flex-1 flex overflow-hidden">
              {/* Slide List Sidebar */}
              <SlideNavigator
                slides={slides}
                activeIndex={activeSlideIndex}
                onSelect={setActiveSlideIndex}
                onAdd={handleAddSlide}
                onDelete={handleDeleteSlide}
                onMoveUp={(i) => handleMoveSlide(i, "up")}
                onMoveDown={(i) => handleMoveSlide(i, "down")}
              />

              {/* Canvas Stage */}
              <ErrorBoundary>
                <SlideCanvas slide={activeSlide} />
              </ErrorBoundary>

              {/* Properties Panel */}
              <SlideProperties slide={activeSlide} onChange={handleUpdateSlide} />
            </div>
          </>
        ) : (
          <EmptyState onSelectMode={(mode) => setDialogMode(mode)} />
        )}
      </main>

      {/* -- Creation Dialog -------------------------------------------- */}
      <CreateStoryDialog
        mode={dialogMode ?? "manual"}
        open={dialogMode !== null}
        isLoading={isProcessing}
        onClose={() => setDialogMode(null)}
        onCreateAI={handleCreateAI}
        onCreateManual={handleCreateManual}
      />

      {/* -- Share Dialog ----------------------------------------------- */}
      <ShareStoryDialog
        story={selectedStory}
        open={isShareDialogOpen}
        onClose={() => setIsShareDialogOpen(false)}
        onTogglePublic={handleTogglePublic}
      />
    </div>
  );
}
