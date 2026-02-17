"use client";

import * as React from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { dataGovernanceApi } from "@/lib/api/data-governance";
import { Badge } from "@/components/ui/badge";

import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Check, Shield, Tag } from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

interface ClassificationTagProps {
  datasourceId: string;
  tableName: string;
  columnName: string;
  currentClassificationId?: number;
  isReadOnly?: boolean;
}

export function ClassificationTag({
  datasourceId,
  tableName,
  columnName,
  currentClassificationId,
  isReadOnly = false,
}: ClassificationTagProps) {
  const [open, setOpen] = React.useState(false);
  const queryClient = useQueryClient();

  // 1. Fetch Classifications
  const { data: classifications = [], isLoading } = useQuery({
    queryKey: ["classifications"],
    queryFn: dataGovernanceApi.getClassifications,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  // 2. Mutation to update classification
  const mutation = useMutation({
    mutationFn: (classificationId: number) => {
      // Find the classification object to optionally optimistically update (not strictly needed here)
      return dataGovernanceApi.updateColumnMetadata({
        datasource_id: datasourceId,
        table_name: tableName,
        column_name: columnName,
        data_classification_id: classificationId,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["column-metadata", datasourceId, tableName] });
      toast.success("Classification updated");
      setOpen(false);
    },
    onError: (error) => {
      toast.error("Failed to update classification: " + error.message);
    },
  });

  const currentClassification = classifications.find((c) => c.id === currentClassificationId);

  const handleSelect = (classificationId: number) => {
    mutation.mutate(classificationId);
  };

  if (isLoading) {
    return <div className="h-5 w-16 animate-pulse rounded bg-muted" />;
  }

  const badgeContent = (
    <Badge
      variant="outline"
      className={cn(
        "cursor-pointer gap-1 transition-colors hover:bg-muted/50",
        !currentClassification && "text-muted-foreground border-dashed",
      )}
      style={{
        borderColor: currentClassification?.color || undefined,
        backgroundColor: (() => {
          if (currentClassification?.color) {
            return `${currentClassification.color}10`; // 10% opacity
          }
          return undefined;
        })(),
        color: currentClassification?.color || undefined,
      }}
    >
      {currentClassification ? (
        <>
          <Shield className="h-3 w-3" />
          {currentClassification.name}
        </>
      ) : (
        <>
          <Tag className="h-3 w-3" />
          Add Tag
        </>
      )}
    </Badge>
  );

  if (isReadOnly) {
    return badgeContent;
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>{badgeContent}</PopoverTrigger>
      <PopoverContent className="w-[200px] p-0" align="start">
        <Command>
          <CommandInput placeholder="Change classification..." />
          <CommandList>
            <CommandEmpty>No classification found.</CommandEmpty>
            <CommandGroup>
              {classifications.map((classification) => (
                <CommandItem
                  key={classification.id}
                  value={classification.name}
                  onSelect={() => handleSelect(classification.id)}
                  className="cursor-pointer"
                >
                  <div
                    className="mr-2 h-2 w-2 rounded-full"
                    style={{ backgroundColor: classification.color }}
                  />
                  {classification.name}
                  {currentClassificationId === classification.id && (
                    <Check className="ml-auto h-4 w-4 opacity-50" />
                  )}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
