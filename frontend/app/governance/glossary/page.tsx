'use client';

export const dynamic = 'force-dynamic';

import { useState, useEffect } from 'react';
import { glossaryApi, type BusinessTerm } from '@/services/glossary';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Plus, Search, Edit, Trash, Book } from 'lucide-react';
import { toast } from 'sonner';

export default function GlossaryPage() {
    const [terms, setTerms] = useState<BusinessTerm[]>([]);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState('');
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [currentTerm, setCurrentTerm] = useState<Partial<BusinessTerm>>({});
    const [isEditing, setIsEditing] = useState(false);

    useEffect(() => {
        fetchTerms();
    }, []);

    const fetchTerms = async () => {
        try {
            setLoading(true);
            const data = await glossaryApi.listTerms();
            setTerms(data);
        } catch (error) {
            toast.error('Failed to load glossary terms');
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        try {
            if (isEditing && currentTerm.id) {
                await glossaryApi.updateTerm(currentTerm.id, currentTerm);
                toast.success('Term updated');
            } else {
                await glossaryApi.createTerm(currentTerm);
                toast.success('Term created');
            }
            setIsDialogOpen(false);
            fetchTerms();
        } catch (error) {
            toast.error('Failed to save term');
        }
    };

    const handleDelete = async (id: string) => {
        // eslint-disable-next-line no-alert
        if (!confirm('Are you sure you want to delete this term?')) return;
        try {
            await glossaryApi.deleteTerm(id);
            toast.success('Term deleted');
            fetchTerms();
        } catch (error) {
            toast.error('Failed to delete term');
        }
    };

    const openCreate = () => {
        setCurrentTerm({ status: 'draft', synonyms: [], tags: [] });
        setIsEditing(false);
        setIsDialogOpen(true);
    };

    const openEdit = (term: BusinessTerm) => {
        setCurrentTerm(term);
        setIsEditing(true);
        setIsDialogOpen(true);
    };

    const filteredTerms = terms.filter(t =>
        t.name.toLowerCase().includes(search.toLowerCase()) ||
        t.definition.toLowerCase().includes(search.toLowerCase())
    );

    return (
        <div className="p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Business Glossary</h1>
                    <p className="text-muted-foreground">Define and manage standardized business terminology.</p>
                </div>
                <Button onClick={openCreate}><Plus className="mr-2 h-4 w-4" /> New Term</Button>
            </div>

            <div className="flex items-center space-x-2">
                <Search className="h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search terms..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="max-w-sm"
                />
            </div>

            {loading ? (
                <div className="text-center py-10">Loading glossary...</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {filteredTerms.map(term => (
                        <Card key={term.id} className="hover:shadow-md transition-shadow">
                            <CardHeader className="pb-2">
                                <div className="flex justify-between items-start">
                                    <CardTitle className="text-xl flex items-center gap-2">
                                        <Book className="h-4 w-4 text-primary" />
                                        {term.name}
                                    </CardTitle>
                                    <Badge variant={term.status === 'approved' ? 'default' : 'secondary'}>
                                        {term.status}
                                    </Badge>
                                </div>
                            </CardHeader>
                            <CardContent>
                                <p className="text-sm text-muted-foreground mb-4 line-clamp-3">
                                    {term.definition}
                                </p>

                                {term.tags && term.tags.length > 0 && (
                                    <div className="flex flex-wrap gap-1 mb-4">
                                        {term.tags.map(tag => (
                                            <Badge key={tag} variant="outline" className="text-xs">{tag}</Badge>
                                        ))}
                                    </div>
                                )}

                                <div className="flex justify-end gap-2 mt-auto">
                                    <Button variant="ghost" size="sm" onClick={() => openEdit(term)}>
                                        <Edit className="h-4 w-4" />
                                    </Button>
                                    <Button variant="ghost" size="sm" onClick={() => handleDelete(term.id)} className="text-destructive hover:text-destructive">
                                        <Trash className="h-4 w-4" />
                                    </Button>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}

            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>{isEditing ? 'Edit Term' : 'Create Business Term'}</DialogTitle>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid gap-2">
                            <label htmlFor="name">Term Name</label>
                            <Input
                                id="name"
                                value={currentTerm.name || ''}
                                onChange={(e) => setCurrentTerm({ ...currentTerm, name: e.target.value })}
                            />
                        </div>
                        <div className="grid gap-2">
                            <label htmlFor="definition">Definition</label>
                            <Textarea
                                id="definition"
                                value={currentTerm.definition || ''}
                                onChange={(e) => setCurrentTerm({ ...currentTerm, definition: e.target.value })}
                            />
                        </div>
                        <div className="grid gap-2">
                            <label htmlFor="status">Status</label>
                            <select
                                id="status"
                                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                                value={currentTerm.status || 'draft'}
                                onChange={(e) => setCurrentTerm({ ...currentTerm, status: e.target.value as any })}
                            >
                                <option value="draft">Draft</option>
                                <option value="approved">Approved</option>
                                <option value="deprecated">Deprecated</option>
                            </select>
                        </div>
                    </div>
                    <div className="flex justify-end gap-2">
                        <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                        <Button onClick={handleSave}>Save</Button>
                    </div>
                </DialogContent>
            </Dialog>
        </div>
    );
}
