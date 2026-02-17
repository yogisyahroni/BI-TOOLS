"use client";

export const dynamic = 'force-dynamic';

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Plus, Search, Book, Database, Tag } from "lucide-react";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";

interface GlossaryTerm {
    id: string;
    term: string;
    definition: string;
    owner: string;
    status: "Draft" | "Approved" | "Deprecated";
    updatedAt: string;
}

const MOCK_TERMS: GlossaryTerm[] = [
    {
        id: "1",
        term: "Active User",
        definition: "A user who has logged in at least once in the last 30 days.",
        owner: "Product Team",
        status: "Approved",
        updatedAt: "2026-02-12",
    },
    {
        id: "2",
        term: "Churn Rate",
        definition: "Percentage of subscribers who discontinue their subscriptions within a given time period.",
        owner: "Data Team",
        status: "Approved",
        updatedAt: "2026-02-10",
    },
    {
        id: "3",
        term: "MRR",
        definition: "Monthly Recurring Revenue.",
        owner: "Finance",
        status: "Draft",
        updatedAt: "2026-02-11",
    },
];

export default function GlossaryPage() {
    const [searchTerm, setSearchTerm] = useState("");
    const [terms, _setTerms] = useState<GlossaryTerm[]>(MOCK_TERMS);

    const filteredTerms = terms.filter((t) =>
        t.term.toLowerCase().includes(searchTerm.toLowerCase()) ||
        t.definition.toLowerCase().includes(searchTerm.toLowerCase())
    );

    return (
        <div className="p-6 space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Business Glossary</h1>
                    <p className="text-muted-foreground mt-2">
                        Define and manage standardized business terms and metrics.
                    </p>
                </div>
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Add Term
                </Button>
            </div>

            <div className="flex gap-4 items-center">
                <div className="relative flex-1">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search terms..."
                        className="pl-8"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
                <div className="flex gap-2">
                    <Badge variant="outline" className="h-9 px-3 cursor-pointer hover:bg-muted">All Status</Badge>
                    <Badge variant="outline" className="h-9 px-3 cursor-pointer hover:bg-muted">My Terms</Badge>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Total Terms</CardTitle>
                        <Book className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{terms.length}</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Mapped Columns</CardTitle>
                        <Database className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">142</div>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Pending Approval</CardTitle>
                        <Tag className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">5</div>
                    </CardContent>
                </Card>
            </div>

            <Card>
                <CardContent className="p-0">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Term</TableHead>
                                <TableHead>Definition</TableHead>
                                <TableHead>Owner</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead>Last Updated</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {filteredTerms.map((term) => (
                                <TableRow key={term.id} className="cursor-pointer hover:bg-muted/50">
                                    <TableCell className="font-medium">{term.term}</TableCell>
                                    <TableCell className="max-w-md truncate" title={term.definition}>{term.definition}</TableCell>
                                    <TableCell>{term.owner}</TableCell>
                                    <TableCell>
                                        <Badge variant={(() => {
                                            if (term.status === "Approved") return "default";
                                            if (term.status === "Draft") return "secondary";
                                            return "destructive";
                                        })()}>
                                            {term.status}
                                        </Badge>
                                    </TableCell>
                                    <TableCell>{term.updatedAt}</TableCell>
                                </TableRow>
                            ))}
                            {filteredTerms.length === 0 && (
                                <TableRow>
                                    <TableCell colSpan={5} className="text-center h-24 text-muted-foreground">
                                        No terms found.
                                    </TableCell>
                                </TableRow>
                            )}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
