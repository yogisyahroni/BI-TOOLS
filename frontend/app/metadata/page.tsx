'use client';

import { useState, useMemo } from 'react';
import { ArrowLeft, Edit2, Database } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { ClassificationTag } from './components/classification-tags';
import Link from 'next/link';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { connectionsApi } from '@/lib/api/connections';
import { dataGovernanceApi } from '@/lib/api/data-governance';
import { toast } from 'sonner';
import { Skeleton } from '@/components/ui/skeleton';
import { type Connection } from '@/types/index';
// ConnectionSchema is SchemaTable[]
import { type ConnectionSchema } from '@/lib/api/connections';
import { type ColumnMetadata } from '@/types/data-governance';

// UI-specific type combining schema and metadata
interface TableEntry {
  id: string;
  tableName: string;
  alias: string;
  description: string;
  columns: {
    name: string;
    type: string;
    tags: string[];
    classificationId?: string | number;
    description: string;
    alias: string;
  }[];
}

export default function MetadataPage() {
  const [selectedConnectionId, setSelectedConnectionId] = useState<string>('');
  const [editingTable, setEditingTable] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<TableEntry | null>(null);
  const queryClient = useQueryClient();

  // 1. Fetch Connections
  const { data: connections = [], isLoading: isLoadingConnections } = useQuery<Connection[]>({
    queryKey: ['connections'],
    queryFn: connectionsApi.list,
  });

  // 2. Fetch Schema (Base Tables)
  const { data: schema = [], isLoading: isLoadingSchema } = useQuery<ConnectionSchema>({
    queryKey: ['schema', selectedConnectionId],
    queryFn: () => connectionsApi.getSchema(selectedConnectionId),
    enabled: !!selectedConnectionId,
  });

  // 3. Fetch Metadata (Tags & Descriptions)
  const { data: metadataList = [], isLoading: isLoadingMetadata } = useQuery<ColumnMetadata[]>({
    queryKey: ['column-metadata', selectedConnectionId],
    queryFn: () => dataGovernanceApi.getColumnMetadata(selectedConnectionId),
    enabled: !!selectedConnectionId,
  });

  // 4. Update Metadata Mutation
  const updateMetadataMutation = useMutation({
    mutationFn: async (data: TableEntry) => {
      // We update column by column
      const promises = data.columns.map((col) => {
        return dataGovernanceApi.updateColumnMetadata({
          datasource_id: selectedConnectionId,
          table_name: data.tableName,
          column_name: col.name,
          description: col.description,
          alias: col.alias, // Now supported
        });
      });
      await Promise.all(promises);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['column-metadata', selectedConnectionId] });
      setEditingTable(null);
      setEditForm(null);
      toast.success('Metadata updated successfully');
    },
    onError: () => {
      toast.error('Failed to update metadata');
    }
  });

  // Merge Schema & Metadata
  const tableEntries: TableEntry[] = useMemo(() => {
    if (!schema || schema.length === 0) return [];

    return schema.map((table) => { // TableInfo
      const tableMetadata = metadataList.filter((m) => m.table_name === table.name);

      // Map columns
      const columns = table.columns.map((col) => {
        const meta = tableMetadata.find((m) => m.column_name === col.name);
        return {
          name: col.name,
          type: col.type,
          tags: meta?.data_classification?.name ? [meta.data_classification.name] : [],
          classificationId: meta?.data_classification_id,
          description: meta?.description || '',
          alias: meta?.alias || '',
        };
      });

      return {
        id: table.name,
        tableName: table.name,
        alias: table.name, // Fallback alias for the table itself (not persisted yet)
        description: '', // Table description not in simple column_metadata yet
        columns: columns,
      };
    });
  }, [schema, metadataList]);

  const handleEdit = (entry: TableEntry) => {
    setEditingTable(entry.tableName);
    // Deep copy for form
    setEditForm(JSON.parse(JSON.stringify(entry)));
  };

  const handleSave = () => {
    if (editForm) {
      updateMetadataMutation.mutate(editForm);
    }
  };

  return (
    <main className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border bg-card sticky top-0 z-10 w-full">
        <div className="max-w-7xl mx-auto px-6 py-6">
          <div className="flex items-center gap-4 mb-4">
            <Link href="/">
              <Button variant="ghost" size="icon">
                <ArrowLeft className="w-4 h-4" />
              </Button>
            </Link>
            <div>
              <h1 className="text-3xl font-bold text-foreground">Metadata Editor</h1>
              <p className="text-muted-foreground mt-1">Configure business meanings and classifications</p>
            </div>
          </div>

          <div className="flex items-center gap-4 mt-6">
            <div className="w-[300px]">
              <Select value={selectedConnectionId} onValueChange={setSelectedConnectionId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select Data Source" />
                </SelectTrigger>
                <SelectContent>
                  {connections.map((conn) => (
                    <SelectItem key={conn.id} value={conn.id}>
                      <div className="flex items-center gap-2">
                        <Database className="w-4 h-4 text-muted-foreground" />
                        {conn.name}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {isLoadingConnections && <Skeleton className="h-10 w-24" />}
          </div>
        </div>
      </header>

      {/* Content */}
      <div className="max-w-7xl mx-auto px-6 py-8">

        {!selectedConnectionId ? (
          <div className="text-center py-20 text-muted-foreground">
            <Database className="w-12 h-12 mx-auto mb-4 opacity-20" />
            <p>Select a data source to begin editing metadata</p>
          </div>
        ) : (
          <>
            {/* Action Bar */}
            <div className="flex items-center justify-between mb-8">
              <div>
                <p className="text-sm text-muted-foreground">
                  {isLoadingSchema ? 'Loading schema...' : `${tableEntries.length} tables found`}
                </p>
              </div>
            </div>

            {/* Loading State */}
            {(isLoadingSchema || isLoadingMetadata) && (
              <div className="grid gap-6">
                {[1, 2, 3].map(i => <Skeleton key={i} className="h-40 w-full" />)}
              </div>
            )}

            {/* Tables Grid */}
            {!isLoadingSchema && !isLoadingMetadata && (
              <div className="grid gap-6">
                {tableEntries.map((entry) => (
                  <Card key={entry.id} className="p-6 border border-border hover:border-primary/50 transition-colors">
                    <div className="flex items-start justify-between mb-4">
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-2">
                          <h3 className="text-lg font-bold text-foreground">{entry.alias}</h3>
                          <code className="text-xs font-mono bg-muted px-2 py-1 rounded text-muted-foreground">
                            {entry.tableName}
                          </code>
                        </div>
                        <p className="text-sm text-muted-foreground">{entry.description || 'No description'}</p>
                      </div>
                      <div className="flex gap-2 ml-4">
                        <Dialog open={editingTable === entry.tableName} onOpenChange={(open) => {
                          if (!open) setEditingTable(null);
                        }}>
                          <DialogTrigger asChild>
                            <Button
                              variant="outline"
                              size="sm"
                              className="gap-2 bg-transparent"
                              onClick={() => handleEdit(entry)}
                            >
                              <Edit2 className="w-4 h-4" />
                              Edit
                            </Button>
                          </DialogTrigger>
                          <DialogContent className="max-w-2xl bg-card">
                            <DialogHeader>
                              <DialogTitle>Edit Metadata - {entry.alias}</DialogTitle>
                              <DialogDescription>
                                Update table description and column descriptions
                              </DialogDescription>
                            </DialogHeader>
                            {editForm && (
                              <div className="space-y-6">
                                <div className="space-y-3">
                                  <label className="text-sm font-medium">Columns</label>
                                  <div className="space-y-2 max-h-64 overflow-y-auto pr-2">
                                    {editForm.columns.map((col, idx) => (
                                      <div key={idx} className="flex gap-2 items-start p-3 bg-muted/50 rounded border border-border">
                                        <div className="flex-1">
                                          <div className="flex justify-between">
                                            <div className="text-xs font-mono font-bold text-foreground">{col.name}</div>
                                            <div className="text-[10px] text-muted-foreground uppercase">{col.type}</div>
                                          </div>
                                          <div className="mt-2 space-y-2">
                                            <Input
                                              value={col.alias || ''}
                                              onChange={(e) => {
                                                const newCols = [...editForm.columns];
                                                newCols[idx].alias = e.target.value;
                                                setEditForm({ ...editForm, columns: newCols });
                                              }}
                                              placeholder="Alias (Friendly Name)..."
                                              className="text-xs h-8 bg-background"
                                            />
                                            <Input
                                              value={col.description || ''}
                                              onChange={(e) => {
                                                const newCols = [...editForm.columns];
                                                newCols[idx].description = e.target.value;
                                                setEditForm({ ...editForm, columns: newCols });
                                              }}
                                              placeholder="Description..."
                                              className="text-xs h-8 bg-background"
                                            />
                                          </div>
                                        </div>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                                <div className="flex gap-2 justify-end pt-4">
                                  <Button variant="outline" onClick={() => setEditingTable(null)}>
                                    Cancel
                                  </Button>
                                  <Button onClick={handleSave} disabled={updateMetadataMutation.isPending}>
                                    {updateMetadataMutation.isPending ? 'Saving...' : 'Save Changes'}
                                  </Button>
                                </div>
                              </div>
                            )}
                          </DialogContent>
                        </Dialog>
                      </div>
                    </div>

                    {/* Columns Preview */}
                    <div className="mt-4">
                      <p className="text-xs font-semibold text-muted-foreground mb-2">Columns ({entry.columns.length})</p>
                      <div className="flex flex-wrap gap-2">
                        {entry.columns.map((col) => (
                          <div key={col.name} className="flex items-center gap-1 px-2 py-1 bg-muted/30 border border-border rounded">
                            <code className="text-xs font-mono text-muted-foreground mr-2">{col.name}</code>
                            {col.alias && <span className="text-xs text-foreground font-medium mr-2">({col.alias})</span>}
                            <ClassificationTag
                              datasourceId={selectedConnectionId}
                              tableName={entry.tableName}
                              columnName={col.name}
                              currentClassificationId={col.classificationId} // Pass the ID!
                            />
                          </div>
                        ))}
                      </div>
                    </div>
                  </Card>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </main>
  );
}
