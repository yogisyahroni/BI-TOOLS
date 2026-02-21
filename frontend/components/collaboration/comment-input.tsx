"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import type { CreateCommentRequest, CommentEntityType } from "@/types/comments";

interface CommentInputProps {
  onSubmit: (data: CreateCommentRequest) => Promise<void>;
  entityType: CommentEntityType;
  entityId: string;
  parentId?: string | null;
  placeholder?: string;
  currentUserId: string;
  onCancel?: () => void;
  submitLabel?: string;
}

export function CommentInput({
  onSubmit,
  entityType,
  entityId,
  parentId = null,
  placeholder = "Add a comment...",
  currentUserId,
  onCancel,
  submitLabel = "Comment",
}: CommentInputProps) {
  const [content, setContent] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async () => {
    if (!content.trim()) return;

    setIsSubmitting(true);
    try {
      await onSubmit({
        entityType,
        entityId,
        content: content.trim(),
        parentId,
      });
      setContent("");
    } catch (error) {
      console.error("Failed to submit comment:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-2">
      <textarea
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder={placeholder}
        className="w-full min-h-[80px] p-2 text-sm border rounded-md resize-none focus:outline-none focus:ring-2 focus:ring-primary"
        disabled={isSubmitting}
      />
      <div className="flex gap-2">
        <Button size="sm" onClick={handleSubmit} disabled={!content.trim() || isSubmitting}>
          {isSubmitting ? "Posting..." : submitLabel}
        </Button>
        {onCancel && (
          <Button size="sm" variant="ghost" onClick={onCancel} disabled={isSubmitting}>
            Cancel
          </Button>
        )}
      </div>
    </div>
  );
}
