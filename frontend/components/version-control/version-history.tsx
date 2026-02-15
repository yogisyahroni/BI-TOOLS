'use client'

import { useState, useEffect, useCallback } from 'react'
import { format } from 'date-fns'
import { groupBy } from 'lodash'
import { 
  History, 
  X, 
  GitCompare,
  Loader2,
  AlertCircle,
  ChevronDown,
  ChevronUp
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { VersionCard } from './version-card'
import { VersionDiff } from './version-diff'
import { VersionRestoreDialog } from './version-restore-dialog'
import type { 
  DashboardVersion, 
  QueryVersion, 
  VersionResourceType,
  VersionTimelineGroup,
  DashboardVersionDiff,
  QueryVersionDiff
} from '@/types/versions'
import { 
  getDashboardVersions, 
  getQueryVersions,
  compareDashboardVersions,
  compareQueryVersions,
  MAX_COMPARE_VERSIONS 
} from '@/lib/api/versions'

interface VersionHistoryProps {
  isOpen: boolean
  onClose: () => void
  resourceType: VersionResourceType
  resourceId: string
  resourceName: string
}

export function VersionHistory({
  isOpen,
  onClose,
  resourceType,
  resourceId,
  resourceName,
}: VersionHistoryProps) {
  const [versions, setVersions] = useState<(DashboardVersion | QueryVersion)[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [selectedVersions, setSelectedVersions] = useState<string[]>([])
  const [expandedGroups, setExpandedGroups] = useState<string[]>(['Today'])
  const [restoreVersion, setRestoreVersion] = useState<DashboardVersion | QueryVersion | null>(null)
  const [previewVersion, setPreviewVersion] = useState<DashboardVersion | QueryVersion | null>(null)
  const [diffData, setDiffData] = useState<DashboardVersionDiff | QueryVersionDiff | null>(null)
  const [isComparing, setIsComparing] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const [offset, setOffset] = useState(0)

  const LIMIT = 20

  // Fetch versions
  const fetchVersions = useCallback(async (reset = false) => {
    if (!resourceId) return
    
    setIsLoading(true)
    setError(null)

    try {
      const newOffset = reset ? 0 : offset
      const filter = { limit: LIMIT, offset: newOffset, orderBy: 'date_desc' as const }

      let response
      if (resourceType === 'dashboard') {
        response = await getDashboardVersions(resourceId, filter)
      } else {
        response = await getQueryVersions(resourceId, filter)
      }

      if (reset) {
        setVersions(response.versions)
      } else {
        setVersions(prev => [...prev, ...response.versions])
      }

      setHasMore(response.total > newOffset + response.versions.length)
      setOffset(newOffset + response.versions.length)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load versions')
    } finally {
      setIsLoading(false)
    }
  }, [resourceId, resourceType, offset])

  // Initial load
  useEffect(() => {
    if (isOpen) {
      fetchVersions(true)
    }
  }, [isOpen, resourceId, resourceType])

  // Group versions by date
  const groupedVersions = groupBy(versions, (version) => {
    const date = new Date(version.createdAt)
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)

    if (date.toDateString() === today.toDateString()) {
      return 'Today'
    } else if (date.toDateString() === yesterday.toDateString()) {
      return 'Yesterday'
    } else if (date > new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000)) {
      return 'Last Week'
    } else if (date > new Date(today.getTime() - 30 * 24 * 60 * 60 * 1000)) {
      return 'Last Month'
    } else {
      return 'Older'
    }
  })

  const timelineGroups: VersionTimelineGroup[] = [
    { label: 'Today', versions: groupedVersions['Today'] || [] },
    { label: 'Yesterday', versions: groupedVersions['Yesterday'] || [] },
    { label: 'Last Week', versions: groupedVersions['Last Week'] || [] },
    { label: 'Last Month', versions: groupedVersions['Last Month'] || [] },
    { label: 'Older', versions: groupedVersions['Older'] || [] },
  ].filter(group => group.versions.length > 0)

  // Handle version selection
  const handleSelectVersion = (id: string) => {
    setSelectedVersions(prev => {
      if (prev.includes(id)) {
        return prev.filter(v => v !== id)
      }
      if (prev.length >= MAX_COMPARE_VERSIONS) {
        return [...prev.slice(1), id]
      }
      return [...prev, id]
    })
  }

  // Handle compare
  const handleCompare = async () => {
    if (selectedVersions.length !== MAX_COMPARE_VERSIONS) {
      return
    }

    setIsComparing(true)
    setError(null)

    try {
      let diff
      if (resourceType === 'dashboard') {
        diff = await compareDashboardVersions(selectedVersions[0], selectedVersions[1])
      } else {
        diff = await compareQueryVersions(selectedVersions[0], selectedVersions[1])
      }
      setDiffData(diff)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to compare versions')
    } finally {
      setIsComparing(false)
    }
  }

  // Toggle group expansion
  const toggleGroup = (label: string) => {
    setExpandedGroups(prev => 
      prev.includes(label) 
        ? prev.filter(g => g !== label)
        : [...prev, label]
    )
  }

  // Handle restore
  const handleRestore = (version: DashboardVersion | QueryVersion) => {
    setRestoreVersion(version)
  }

  // Handle preview
  const handlePreview = (version: DashboardVersion | QueryVersion) => {
    setPreviewVersion(version)
  }

  // Handle compare from card
  const handleCompareFromCard = (version: DashboardVersion | QueryVersion) => {
    if (selectedVersions.length === 0) {
      setSelectedVersions([version.id])
    } else if (selectedVersions.length === 1 && selectedVersions[0] !== version.id) {
      setSelectedVersions([selectedVersions[0], version.id])
      handleCompare()
    } else {
      setSelectedVersions([version.id])
    }
  }

  // Load more versions
  const loadMore = () => {
    fetchVersions(false)
  }

  // Find version by ID
  const findVersion = (id: string) => versions.find(v => v.id === id)

  return (
    <>
      <Dialog open={isOpen} onOpenChange={onClose}>
        <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
          <DialogHeader>
            <div className="flex items-center justify-between">
              <div>
                <DialogTitle className="flex items-center gap-2">
                  <History className="h-5 w-5" />
                  Version History
                </DialogTitle>
                <DialogDescription>
                  {resourceName}
                  <Badge variant="outline" className="ml-2 capitalize">
                    {resourceType}
                  </Badge>
                </DialogDescription>
              </div>
              
              {/* Compare button */}
              {selectedVersions.length === MAX_COMPARE_VERSIONS && (
                <Button 
                  onClick={handleCompare}
                  disabled={isComparing}
                  size="sm"
                >
                  {isComparing ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : (
                    <GitCompare className="h-4 w-4 mr-2" />
                  )}
                  Compare
                </Button>
              )}
            </div>
          </DialogHeader>

          {/* Selection indicator */}
          {selectedVersions.length > 0 && selectedVersions.length < MAX_COMPARE_VERSIONS && (
            <Alert className="mt-2">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Select {MAX_COMPARE_VERSIONS - selectedVersions.length} more version{MAX_COMPARE_VERSIONS - selectedVersions.length > 1 ? 's' : ''} to compare
              </AlertDescription>
            </Alert>
          )}

          {/* Error */}
          {error && (
            <Alert variant="destructive" className="mt-2">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Timeline */}
          <ScrollArea className="flex-1 mt-4">
            {timelineGroups.map((group) => (
              <div key={group.label} className="mb-6">
                <button
                  onClick={() => toggleGroup(group.label)}
                  className="flex items-center gap-2 w-full text-left mb-3 hover:bg-muted/50 p-2 rounded-lg transition-colors"
                >
                  {expandedGroups.includes(group.label) ? (
                    <ChevronUp className="h-4 w-4 text-muted-foreground" />
                  ) : (
                    <ChevronDown className="h-4 w-4 text-muted-foreground" />
                  )}
                  <h3 className="font-semibold text-sm text-muted-foreground uppercase tracking-wide">
                    {group.label}
                  </h3>
                  <Badge variant="secondary" className="text-xs">
                    {group.versions.length}
                  </Badge>
                </button>

                {expandedGroups.includes(group.label) && (
                  <div className="space-y-3 pl-6">
                    {group.versions.map((version) => (
                      <VersionCard
                        key={version.id}
                        version={version}
                        isSelected={selectedVersions.includes(version.id)}
                        onSelect={handleSelectVersion}
                        onRestore={handleRestore}
                        onPreview={handlePreview}
                        onCompare={handleCompareFromCard}
                        disabled={isLoading}
                      />
                    ))}
                  </div>
                )}

                <Separator className="mt-4" />
              </div>
            ))}

            {/* Load more */}
            {hasMore && (
              <div className="flex justify-center py-4">
                <Button 
                  variant="outline" 
                  onClick={loadMore}
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  ) : null}
                  Load More
                </Button>
              </div>
            )}

            {/* Empty state */}
            {!isLoading && versions.length === 0 && (
              <div className="text-center py-12">
                <History className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-medium text-foreground">No versions yet</h3>
                <p className="text-muted-foreground mt-1">
                  Versions are created when you save changes or auto-save is triggered.
                </p>
              </div>
            )}
          </ScrollArea>
        </DialogContent>
      </Dialog>

      {/* Restore Dialog */}
      <VersionRestoreDialog
        isOpen={!!restoreVersion}
        onClose={() => setRestoreVersion(null)}
        version={restoreVersion}
        resourceType={resourceType}
        resourceName={resourceName}
        onRestored={() => {
          setRestoreVersion(null)
          fetchVersions(true)
        }}
      />

      {/* Diff View */}
      {diffData && (
        <VersionDiff
          diff={diffData}
          version1={findVersion(selectedVersions[0])}
          version2={findVersion(selectedVersions[1])}
          onClose={() => {
            setDiffData(null)
            setSelectedVersions([])
          }}
        />
      )}
    </>
  )
}
