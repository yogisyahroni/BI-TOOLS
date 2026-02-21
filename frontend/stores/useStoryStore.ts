import { create } from "zustand";
import { Story, StoryPage, StoryCard } from "@/lib/types";

interface StoryState {
  stories: Story[];
  currentStory: Story | null;
  currentPageIndex: number;

  addStory: (story: Story) => void;
  updateStory: (id: string, updates: Partial<Story>) => void;
  deleteStory: (id: string) => void;
  loadStory: (id: string) => void;

  addPage: (page: StoryPage) => void;
  updatePage: (pageId: string, updates: Partial<StoryPage>) => void;
  removePage: (pageId: string) => void;

  addCardToPage: (pageId: string, card: StoryCard) => void;
  updateCard: (pageId: string, cardId: string, updates: Partial<StoryCard>) => void;
  removeCard: (pageId: string, cardId: string) => void;

  setCurrentPage: (index: number) => void;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  applyFilter: (filterId: string, value: any) => void;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  getFilteredResults: (queryResults: any[]) => any[];
}

export const useStoryStore = create<StoryState>((set, get) => ({
  stories: [],
  currentStory: null,
  currentPageIndex: 0,

  addStory: (story) => set((state) => ({ stories: [...state.stories, story] })),

  updateStory: (id, updates) =>
    set((state) => {
      const updatedStories = state.stories.map((s) => (s.id === id ? { ...s, ...updates } : s));
      const isCurrent = state.currentStory?.id === id;
      return {
        stories: updatedStories,
        currentStory:
          isCurrent && state.currentStory
            ? { ...state.currentStory, ...updates }
            : state.currentStory,
      };
    }),

  deleteStory: (id) =>
    set((state) => ({
      stories: state.stories.filter((s) => s.id !== id),
      currentStory: state.currentStory?.id === id ? null : state.currentStory,
    })),

  loadStory: (id) =>
    set((state) => {
      const story = state.stories.find((s) => s.id === id);
      if (story) {
        return { currentStory: story, currentPageIndex: 0 };
      }
      return state;
    }),

  addPage: (page) =>
    set((state) => {
      if (!state.currentStory) return state;
      const updatedPages = [...state.currentStory.pages, page];
      return {
        currentStory: { ...state.currentStory, pages: updatedPages },
      };
    }),

  updatePage: (pageId, updates) =>
    set((state) => {
      if (!state.currentStory) return state;
      const updatedPages = state.currentStory.pages.map((p) =>
        p.id === pageId ? { ...p, ...updates } : p,
      );
      return {
        currentStory: { ...state.currentStory, pages: updatedPages },
      };
    }),

  removePage: (pageId) =>
    set((state) => {
      if (!state.currentStory) return state;
      const updatedPages = state.currentStory.pages.filter((p) => p.id !== pageId);
      return {
        currentStory: { ...state.currentStory, pages: updatedPages },
      };
    }),

  addCardToPage: (pageId, card) =>
    set((state) => {
      if (!state.currentStory) return state;
      const updatedPages = state.currentStory.pages.map((p) =>
        p.id === pageId ? { ...p, cards: [...p.cards, card] } : p,
      );
      return {
        currentStory: { ...state.currentStory, pages: updatedPages },
      };
    }),

  updateCard: (pageId, cardId, updates) =>
    set((state) => {
      if (!state.currentStory) return state;
      const updatedPages = state.currentStory.pages.map((p) =>
        p.id === pageId
          ? {
              ...p,
              cards: p.cards.map((c) => (c.id === cardId ? { ...c, ...updates } : c)),
            }
          : p,
      );
      return {
        currentStory: { ...state.currentStory, pages: updatedPages },
      };
    }),

  removeCard: (pageId, cardId) =>
    set((state) => {
      if (!state.currentStory) return state;
      const updatedPages = state.currentStory.pages.map((p) =>
        p.id === pageId ? { ...p, cards: p.cards.filter((c) => c.id !== cardId) } : p,
      );
      return {
        currentStory: { ...state.currentStory, pages: updatedPages },
      };
    }),

  setCurrentPage: (index) =>
    set((state) => {
      if (state.currentStory && index >= 0 && index < state.currentStory.pages.length) {
        return { currentPageIndex: index };
      }
      return state;
    }),

  applyFilter: (filterId, value) => {
    set((state) => {
      if (!state.currentStory) return state;
      const updatedFilters = state.currentStory.filters.map((f) =>
        f.id === filterId ? { ...f, selectedValue: value } : f,
      );
      return {
        currentStory: { ...state.currentStory, filters: updatedFilters },
      };
    });
  },

  getFilteredResults: (queryResults) => {
    const { currentStory } = get();
    if (!currentStory || !currentStory.filters || currentStory.filters.length === 0) {
      return queryResults;
    }

    return queryResults.filter((row) => {
      return currentStory.filters.every((filter) => {
        if (!filter.selectedValue) return true;

        const rowValue = row[filter.column];
        if (filter.type === "select") {
          return rowValue === filter.selectedValue;
        } else if (filter.type === "multi-select") {
          return Array.isArray(filter.selectedValue)
            ? filter.selectedValue.includes(rowValue)
            : rowValue === filter.selectedValue;
        } else if (filter.type === "range") {
          const [min, max] = filter.selectedValue;
          return rowValue >= min && rowValue <= max;
        } else if (filter.type === "date-range") {
          const [startDate, endDate] = filter.selectedValue;
          return rowValue >= startDate && rowValue <= endDate;
        }
        return true;
      });
    });
  },
}));
