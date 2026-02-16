'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import {
  X,
  ChevronLeft,
  ChevronRight,
  Plus,
  Minus,
  Edit3,
  Layout,
  Filter,
  Type,
  Grid3X3,
  Code,
  Tag,
  GitCompare
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { Utils } from '@/lib/utils'
import { Alert, AlertDescription } from '@/components/ui/alert'
import type {
  DashboardVersionDiff,
  QueryVersionDiff,
  DashboardVersion,
  QueryVersion,
  VersionCard,
  DashboardCardChange
} from '@/types/versions'

interface VersionDiffProps {
  diff: DashboardVersionDiff | QueryVersionDiff
  version1: DashboardVersion | QueryVersion | undefined
  version2: DashboardVersion | QueryVersion | undefined
  onClose: () => void
}

export function VersionDiff({ diff, version1, version2, onClose }: VersionDiffProps) {
  const [activeTab, setActiveTab] = useState('overview')
  const isDashboardDiff = 'cardsDiff' in diff

  const renderChangeIndicator = (changed: boolean, from?: string, to?: string, label?: string) => {
    if (!changed) return null

    return (
      <div className="space-y-2 p-3 rounded-lg bg-muted/50">
        <div className="flex items-center gap-2 text-sm font-medium">
          {label && <span className="text-muted-foreground">{label}:</span>}
          <Badge variant="outline" className="text-yellow-600 bg-yellow-50">
            <Edit3 className="h-3 w-3 mr-1" />
            Modified
          </Badge>
        </div>
        {from && to && (
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div className="space-y-1">
              <span className="text-xs text-muted-foreground">Before</span>
              <div className="p-2 rounded bg-red-50 text-red-700 line-through">
                {from}
              </div>
            </div>
            <div className="space-y-1">
              <span className="text-xs text-muted-foreground">After</span>
              <div className="p-2 rounded bg-green-50 text-green-700">
                {to}
              </div>
            </div>
          </div>
        )}
      </div>
    )
  }

  const renderCardChanges = (cardChanges: DashboardCardChange[]) => {
    if (cardChanges.length === 0) return null

    return (
      <div className="space-y-3">
        {cardChanges.map((change, index) => (
          <div key={index} className="p-3 rounded-lg border border-yellow-200 bg-yellow-50/50">
            <div className="flex items-center gap-2 mb-2">
              <Edit3 className="h-4 w-4 text-yellow-600" />
              <span className="font-medium text-sm">Card Modified</span>
              <Badge variant="secondary" className="text-xs">
                {change.changes.join(', ')}
              </Badge>
            </div>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-xs text-muted-foreground block mb-1">Before</span>
                <div className="text-xs text-muted-foreground line-through">
                  {change.before.title || 'Untitled'}
                </div>
              </div>
              <div>
                <span className="text-xs text-muted-foreground block mb-1">After</span>
                <div className="text-xs text-foreground">
                  {change.after.title || 'Untitled'}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  const renderDashboardDiff = (diff: DashboardVersionDiff) => {
    const { cardsDiff } = diff

    return (
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="cards">
            Cards
            {cardsDiff.added.length + cardsDiff.removed.length + cardsDiff.modified.length > 0 && (
              <Badge variant="secondary" className="ml-2 text-xs">
                {cardsDiff.added.length + cardsDiff.removed.length + cardsDiff.modified.length}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value="layout">Layout</TabsTrigger>
          <TabsTrigger value="filters">Filters</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4 mt-4">
          {/* Summary stats */}
          <div className="grid grid-cols-4 gap-4">
            <div className="p-4 rounded-lg bg-muted text-center">
              <div className="text-2xl font-bold text-green-600">{cardsDiff.added.length}</div>
              <div className="text-xs text-muted-foreground">Added</div>
            </div>
            <div className="p-4 rounded-lg bg-muted text-center">
              <div className="text-2xl font-bold text-red-600">{cardsDiff.removed.length}</div>
              <div className="text-xs text-muted-foreground">Removed</div>
            </div>
            <div className="p-4 rounded-lg bg-muted text-center">
              <div className="text-2xl font-bold text-yellow-600">{cardsDiff.modified.length}</div>
              <div className="text-xs text-muted-foreground">Modified</div>
            </div>
            <div className="p-4 rounded-lg bg-muted text-center">
              <div className="text-2xl font-bold">{cardsDiff.unchanged.length}</div>
              <div className="text-xs text-muted-foreground">Unchanged</div>
            </div>
          </div>

          {/* Name change */}
          {diff.nameChanged && renderChangeIndicator(true, diff.nameFrom, diff.nameTo, 'Name')}

          {/* Description change */}
          {diff.descChanged && renderChangeIndicator(true, diff.descFrom, diff.descTo, 'Description')}
        </TabsContent>

        <TabsContent value="cards" className="space-y-4 mt-4">
          <ScrollArea className="h-[400px]">
            {/* Added cards */}
            {cardsDiff.added.length > 0 && (
              <div className="mb-6">
                <h4 className="text-sm font-medium flex items-center gap-2 mb-3">
                  <Plus className="h-4 w-4 text-green-600" />
                  Added ({cardsDiff.added.length})
                </h4>
                <div className="space-y-2">
                  {cardsDiff.added.map((card, index) => (
                    <div key={index} className="p-3 rounded-lg border border-green-200 bg-green-50">
                      <div className="text-sm font-medium">{card.title || 'Untitled Card'}</div>
                      {card.queryId && (
                        <div className="text-xs text-muted-foreground mt-1">
                          Query ID: {card.queryId}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Removed cards */}
            {cardsDiff.removed.length > 0 && (
              <div className="mb-6">
                <h4 className="text-sm font-medium flex items-center gap-2 mb-3">
                  <Minus className="h-4 w-4 text-red-600" />
                  Removed ({cardsDiff.removed.length})
                </h4>
                <div className="space-y-2">
                  {cardsDiff.removed.map((card, index) => (
                    <div key={index} className="p-3 rounded-lg border border-red-200 bg-red-50">
                      <div className="text-sm font-medium line-through">{card.title || 'Untitled Card'}</div>
                      {card.queryId && (
                        <div className="text-xs text-muted-foreground mt-1">
                          Query ID: {card.queryId}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Modified cards */}
            {cardsDiff.modified.length > 0 && (
              <div className="mb-6">
                <h4 className="text-sm font-medium flex items-center gap-2 mb-3">
                  <Edit3 className="h-4 w-4 text-yellow-600" />
                  Modified ({cardsDiff.modified.length})
                </h4>
                {renderCardChanges(cardsDiff.modified)}
              </div>
            )}

            {/* Unchanged cards */}
            {cardsDiff.unchanged.length > 0 && (
              <div>
                <h4 className="text-sm font-medium flex items-center gap-2 mb-3">
                  <Grid3X3 className="h-4 w-4 text-muted-foreground" />
                  Unchanged ({cardsDiff.unchanged.length})
                </h4>
                <div className="space-y-2">
                  {cardsDiff.unchanged.slice(0, 5).map((card, index) => (
                    <div key={index} className="p-3 rounded-lg border border-muted bg-muted/50">
                      <div className="text-sm text-muted-foreground">{card.title || 'Untitled Card'}</div>
                    </div>
                  ))}
                  {cardsDiff.unchanged.length > 5 && (
                    <div className="text-xs text-muted-foreground text-center py-2">
                      +{cardsDiff.unchanged.length - 5} more unchanged
                    </div>
                  )}
                </div>
              </div>
            )}
          </ScrollArea>
        </TabsContent>

        <TabsContent value="layout" className="space-y-4 mt-4">
          {diff.layoutChanged ? (
            <div className="space-y-4">
              <Alert className="bg-yellow-50 border-yellow-200">
                <Layout className="h-4 w-4 text-yellow-600" />
                <AlertDescription className="text-yellow-800">
                  Layout configuration has changed between versions
                </AlertDescription>
              </Alert>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 rounded-lg bg-muted">
                  <h5 className="text-sm font-medium mb-2">Previous Layout</h5>
                  <pre className="text-xs overflow-auto">{diff.layoutFrom || 'None'}</pre>
                </div>
                <div className="p-4 rounded-lg bg-muted">
                  <h5 className="text-sm font-medium mb-2">New Layout</h5>
                  <pre className="text-xs overflow-auto">{diff.layoutTo || 'None'}</pre>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Layout className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No layout changes detected</p>
            </div>
          )}
        </TabsContent>

        <TabsContent value="filters" className="space-y-4 mt-4">
          {diff.filtersChanged ? (
            <div className="space-y-4">
              <Alert className="bg-yellow-50 border-yellow-200">
                <Filter className="h-4 w-4 text-yellow-600" />
                <AlertDescription className="text-yellow-800">
                  Filters have changed between versions
                </AlertDescription>
              </Alert>
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 rounded-lg bg-muted">
                  <h5 className="text-sm font-medium mb-2">Previous Filters</h5>
                  <pre className="text-xs overflow-auto">{diff.filtersFrom || 'None'}</pre>
                </div>
                <div className="p-4 rounded-lg bg-muted">
                  <h5 className="text-sm font-medium mb-2">New Filters</h5>
                  <pre className="text-xs overflow-auto">{diff.filtersTo || 'None'}</pre>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Filter className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No filter changes detected</p>
            </div>
          )}
        </TabsContent>
      </Tabs>
    )
  }

  const renderQueryDiff = (diff: QueryVersionDiff) => {
    return (
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="sql">SQL</TabsTrigger>
          <TabsTrigger value="metadata">Metadata</TabsTrigger>
          <TabsTrigger value="tags">Tags</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4 mt-4">
          {/* Name change */}
          {diff.nameChanged && renderChangeIndicator(true, diff.nameFrom, diff.nameTo, 'Name')}

          {/* Description change */}
          {diff.descChanged && renderChangeIndicator(true, diff.descFrom, diff.descTo, 'Description')}

          {/* SQL change indicator */}
          {diff.sqlChanged && (
            <div className="p-3 rounded-lg bg-yellow-50 border border-yellow-200">
              <div className="flex items-center gap-2">
                <Code className="h-4 w-4 text-yellow-600" />
                <span className="font-medium text-sm">SQL Query Modified</span>
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                View the SQL tab to see the changes
              </p>
            </div>
          )}
        </TabsContent>

        <TabsContent value="sql" className="space-y-4 mt-4">
          {diff.sqlChanged ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="p-4 rounded-lg bg-muted">
                  <h5 className="text-sm font-medium mb-2 flex items-center gap-2">
                    <ChevronLeft className="h-4 w-4" />
                    Previous SQL
                  </h5>
                  <pre className="text-xs overflow-auto whitespace-pre-wrap font-mono bg-red-50 p-2 rounded">
                    {diff.sqlFrom}
                  </pre>
                </div>
                <div className="p-4 rounded-lg bg-muted">
                  <h5 className="text-sm font-medium mb-2 flex items-center gap-2">
                    <ChevronRight className="h-4 w-4" />
                    New SQL
                  </h5>
                  <pre className="text-xs overflow-auto whitespace-pre-wrap font-mono bg-green-50 p-2 rounded">
                    {diff.sqlTo}
                  </pre>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Code className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No SQL changes detected</p>
            </div>
          )}
        </TabsContent>

        <TabsContent value="metadata" className="space-y-4 mt-4">
          {diff.aiPromptChanged && renderChangeIndicator(true, diff.aiPromptFrom, diff.aiPromptTo, 'AI Prompt')}

          {diff.visualizationChanged && (
            <div className="p-3 rounded-lg bg-yellow-50 border border-yellow-200">
              <div className="flex items-center gap-2">
                <Type className="h-4 w-4 text-yellow-600" />
                <span className="font-medium text-sm">Visualization Config Modified</span>
              </div>
            </div>
          )}
        </TabsContent>

        <TabsContent value="tags" className="space-y-4 mt-4">
          {(diff.tagsAdded.length > 0 || diff.tagsRemoved.length > 0) ? (
            <div className="space-y-4">
              {diff.tagsAdded.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium flex items-center gap-2 mb-2">
                    <Plus className="h-4 w-4 text-green-600" />
                    Tags Added
                  </h4>
                  <div className="flex flex-wrap gap-2">
                    {diff.tagsAdded.map((tag, index) => (
                      <Badge key={index} variant="secondary" className="bg-green-100 text-green-800">
                        {tag}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}

              {diff.tagsRemoved.length > 0 && (
                <div>
                  <h4 className="text-sm font-medium flex items-center gap-2 mb-2">
                    <Minus className="h-4 w-4 text-red-600" />
                    Tags Removed
                  </h4>
                  <div className="flex flex-wrap gap-2">
                    {diff.tagsRemoved.map((tag, index) => (
                      <Badge key={index} variant="secondary" className="bg-red-100 text-red-800 line-through">
                        {tag}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="text-center py-8 text-muted-foreground">
              <Tag className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p>No tag changes detected</p>
            </div>
          )}
        </TabsContent>
      </Tabs>
    )
  }

  return (
    <Dialog open={true} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl h-[80vh] flex flex-col">
        <DialogHeader>
          <div className="flex items-center justify-between">
            <DialogTitle className="flex items-center gap-2">
              <GitCompare className="h-5 w-5" />
              Compare Versions
            </DialogTitle>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </DialogHeader>

        {/* Version headers */}
        <div className="grid grid-cols-2 gap-4 p-4 bg-muted/50 rounded-lg">
          <div className="text-center">
            <Badge variant="outline" className="mb-2">Version {version1?.version || '?'}</Badge>
            <p className="text-sm text-muted-foreground">
              {version1?.createdAt && format(new Date(version1.createdAt), 'PPp')}
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              by {version1?.createdByUser?.name || 'Unknown'}
            </p>
          </div>
          <div className="text-center">
            <Badge variant="outline" className="mb-2">Version {version2?.version || '?'}</Badge>
            <p className="text-sm text-muted-foreground">
              {version2?.createdAt && format(new Date(version2.createdAt), 'PPp')}
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              by {version2?.createdByUser?.name || 'Unknown'}
            </p>
          </div>
        </div>

        <Separator />

        {/* Diff content */}
        <div className="flex-1 overflow-hidden">
          {isDashboardDiff
            ? renderDashboardDiff(diff as DashboardVersionDiff)
            : renderQueryDiff(diff as QueryVersionDiff)
          }
        </div>
      </DialogContent>
    </Dialog>
  )
}
