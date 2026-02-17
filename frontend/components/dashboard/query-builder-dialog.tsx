'use client';

import { useState } from 'react';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { DualEngineEditor } from '@/components/dual-engine-editor';
import { ResultsTable } from '@/components/query-results/results-table';
import { _Button } from '@/components/ui/button';
import { _X } from 'lucide-react';
import { type SavedQuery } from '@/lib/types';
import { toast } from 'sonner';

interface QueryBuilderDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSave: (query: SavedQuery) => void;
    connectionId?: string;
}

export function QueryBuilderDialog({
    open,
    onOpenChange,
    onSave,
    connectionId = 'db1', // Default for now, ideally passed from context
}: QueryBuilderDialogProps) {

    const [queryResults, setQueryResults] = useState<{
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        data: any[];
        columns: string[];
        rowCount: number;
        executionTime: number;
        isLoading: boolean;
    }>({
        data: [],
        columns: [],
        rowCount: 0,
        executionTime: 0,
        isLoading: false,
    });

    const handleSaveSuccess = (query: SavedQuery) => {
        toast.success('Query created and added to dashboard');
        onSave(query);
        onOpenChange(false);
    };

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const handleResultsUpdate = (results: {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        data: any[];
        columns: string[];
        rowCount: number;
        executionTime: number;
    }) => {
        setQueryResults({
            ...results,
            isLoading: false,
        });
    };

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-[90vw] h-[85vh] flex flex-col p-0 gap-0">
                <DialogHeader className="px-6 py-4 border-b border-border flex flex-row items-center justify-between">
                    <div className="space-y-1">
                        <DialogTitle>Create New Query</DialogTitle>
                        <DialogDescription>
                            Write SQL or ask AI to generate data for your dashboard
                        </DialogDescription>
                    </div>
                    {/* Close button is provided by DialogContent, but we can customize if needed */}
                </DialogHeader>

                <div className="flex-1 flex flex-col overflow-hidden bg-muted/10">
                    <div className="flex-shrink-0 border-b border-border">
                        <DualEngineEditor
                            mode="modal"
                            connectionId={connectionId}
                            onSchemaClick={() => toast.info('Schema browser is limited in quick mode')}
                            onSaveSuccess={handleSaveSuccess}
                            onResultsUpdate={handleResultsUpdate}
                        />
                    </div>

                    <div className="flex-1 overflow-auto bg-background">
                        {queryResults.data.length > 0 || queryResults.isLoading ? (
                            <div className="p-4 h-full">
                                <h4 className="text-xs font-semibold text-muted-foreground mb-2">Query Preview</h4>
                                <ResultsTable
                                    data={queryResults.data}
                                    columns={queryResults.columns}
                                    rowCount={queryResults.rowCount}
                                    executionTime={queryResults.executionTime}
                                    isLoading={queryResults.isLoading}
                                />
                            </div>
                        ) : (
                            <div className="h-full flex items-center justify-center text-muted-foreground text-sm">
                                Run a query to see results here
                            </div>
                        )}
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    );
}
