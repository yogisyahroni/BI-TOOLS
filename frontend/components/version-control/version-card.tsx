'use client'

import { useState } from 'react'
import { formatDistanceToNow } from 'date-fns'
import { 
  History, 
  RotateCcw, 
  Eye, 
  GitCompare, 
  Trash2,
  Clock,
  _User,
  ChevronRight
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import type { DashboardVersion, QueryVersion } from '@/types/versions'

interface VersionCardProps {
  version: DashboardVersion | QueryVersion
  isSelected: boolean
  onSelect: (id: string) => void
  onRestore: (version: DashboardVersion | QueryVersion) => void
  onPreview: (version: DashboardVersion | QueryVersion) => void
  onCompare: (version: DashboardVersion | QueryVersion) => void
  onDelete?: (version: DashboardVersion | QueryVersion) => void
  disabled?: boolean
}

export function VersionCard({
  version,
  isSelected,
  onSelect,
  onRestore,
  onPreview,
  onCompare,
  onDelete,
  disabled = false,
}: VersionCardProps) {
  const [isHovered, setIsHovered] = useState(false)
  
  const createdByUser = version.createdByUser
  const createdAt = new Date(version.createdAt)
  const timeAgo = formatDistanceToNow(createdAt, { addSuffix: true })
  
  // Get initials for avatar
  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }

  return (
    <div
      className={`
        relative group rounded-lg border p-4 transition-all duration-200
        ${isSelected ? 'border-primary bg-primary/5' : 'border-border bg-card hover:border-muted-foreground/20'}
        ${disabled ? 'opacity-50 pointer-events-none' : ''}
      `}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <div className="flex items-start gap-4">
        {/* Checkbox for selection */}
        <div className="pt-1">
          <Checkbox
            checked={isSelected}
            onCheckedChange={() => onSelect(version.id)}
            aria-label={`Select version ${version.version}`}
          />
        </div>

        {/* Version info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-2">
              <span className="text-lg font-semibold text-foreground">
                Version {version.version}
              </span>
              {version.isAutoSave && (
                <Badge variant="secondary" className="text-xs">
                  <Clock className="w-3 h-3 mr-1" />
                  Auto-save
                </Badge>
              )}
            </div>
            
            {/* Actions dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onClick={() => onPreview(version)}>
                  <Eye className="mr-2 h-4 w-4" />
                  Preview
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => onCompare(version)}>
                  <GitCompare className="mr-2 h-4 w-4" />
                  Compare
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={() => onRestore(version)}
                  className="text-destructive focus:text-destructive"
                >
                  <RotateCcw className="mr-2 h-4 w-4" />
                  Restore
                </DropdownMenuItem>
                {onDelete && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem 
                      onClick={() => onDelete(version)}
                      className="text-destructive focus:text-destructive"
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Change summary */}
          {version.changeSummary && (
            <p className="text-sm text-muted-foreground mt-1">
              {version.changeSummary}
            </p>
          )}

          {/* Metadata */}
          <div className="flex items-center gap-4 mt-3 text-sm text-muted-foreground">
            {/* User */}
            <div className="flex items-center gap-2">
              <Avatar className="h-6 w-6">
                <AvatarImage src={createdByUser?.avatar || createdByUser?.image} />
                <AvatarFallback className="text-xs">
                  {createdByUser?.name ? getInitials(createdByUser.name) : 'U'}
                </AvatarFallback>
              </Avatar>
              <span className="truncate max-w-[120px]">
                {createdByUser?.name || 'Unknown'}
              </span>
            </div>

            {/* Timestamp */}
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger className="flex items-center gap-1">
                  <History className="h-3.5 w-3.5" />
                  <span>{timeAgo}</span>
                </TooltipTrigger>
                <TooltipContent>
                  <p>{createdAt.toLocaleString()}</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>
      </div>

      {/* Quick actions (visible on hover) */}
      {isHovered && !disabled && (
        <div className="absolute right-4 bottom-4 flex items-center gap-2">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onPreview(version)}
                  className="h-8 w-8 p-0"
                >
                  <Eye className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Preview</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onCompare(version)}
                  className="h-8 w-8 p-0"
                >
                  <GitCompare className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Compare</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onRestore(version)}
                  className="h-8 w-8 p-0 text-destructive hover:text-destructive"
                >
                  <RotateCcw className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>Restore</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      )}
    </div>
  )
}
