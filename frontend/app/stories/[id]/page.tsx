import { SlideDeck } from "@/types/presentation";
import { StoryBuilder } from "@/components/presentation/story-builder";

interface StoryPageProps {
  params: {
    id: string;
  };
  searchParams: { [key: string]: string | string[] | undefined };
}

export default function StoryPage({ params, searchParams }: StoryPageProps) {
  // In a real app, we'd fetch the story by ID here.
  // For now, we allow creating a "draft" or loading by ID (mocked/client-side).
  // If ID is truly persistent, we'd fetch it.

  // We pass dashboardId if present to create context.
  const dashboardId =
    typeof searchParams?.dashboardId === "string" ? searchParams.dashboardId : undefined;

  return (
    <StoryBuilder
      dashboardId={dashboardId}
      // initialSlides={fetchedSlides} // If we loaded from DB
    />
  );
}
