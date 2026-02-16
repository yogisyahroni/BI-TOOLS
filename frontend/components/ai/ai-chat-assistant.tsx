'use client';

import { useState, useRef, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import {
    Sparkles,
    Send,
    User,
    Bot,
    Loader2,
    StopCircle,
    ChevronDown,
    ChevronRight,
    BrainCircuit,
    MessageSquare,
    X,
    Minimize2,
    Maximize2,
} from 'lucide-react';
import { useAIStream } from '@/hooks/use-ai-stream';
import { useAIProviders } from '@/hooks/use-ai-providers';
import { aiApi } from '@/lib/api/ai';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';
import ReactMarkdown from 'react-markdown';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { vscDarkPlus } from 'react-syntax-highlighter/dist/esm/styles/prism';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select';
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet';

interface ChatMessage {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    reasoning?: {
        steps: { step: number; thought: string; durationMs: number }[];
        plan?: string;
        totalDurationMs: number;
    };
    isStreaming?: boolean;
    timestamp: Date;
}

interface AIChatAssistantProps {
    currentSQL?: string;
    currentPrompt?: string;
    onApplySQL?: (sql: string) => void;
    className?: string;
}

export function AIChatAssistant({ 
    currentSQL, 
    currentPrompt, 
    onApplySQL,
    className 
}: AIChatAssistantProps) {
    const [prompt, setPrompt] = useState('');
    const [selectedProviderId, setSelectedProviderId] = useState<string>('');
    const [enableReasoning, setEnableReasoning] = useState(false);
    const [chatHistory, setChatHistory] = useState<ChatMessage[]>([]);
    const [isOpen, setIsOpen] = useState(false);
    const [isExpanded, setIsExpanded] = useState(false);
    const scrollRef = useRef<HTMLDivElement>(null);

    const { providers } = useAIProviders();
    const activeProviders = providers.filter((p) => p.isActive);
    const defaultProvider = activeProviders.find((p) => p.isDefault);

    // Auto-scroll to bottom
    useEffect(() => {
        if (scrollRef.current) {
            scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
        }
    }, [chatHistory]);

    // Handle Streaming
    const { generateStream: startStream, isStreaming: streamActive, stopStream: stopGeneration } = useAIStream({
        onToken: (token) => {
            setChatHistory(prev => {
                const newHistory = [...prev];
                const lastMsg = newHistory[newHistory.length - 1];
                if (lastMsg.role === 'assistant') {
                    lastMsg.content += token;
                    lastMsg.isStreaming = true;
                }
                return newHistory;
            });
        },
        onComplete: () => {
            setChatHistory(prev => {
                const newHistory = [...prev];
                const lastMsg = newHistory[newHistory.length - 1];
                if (lastMsg.role === 'assistant') {
                    lastMsg.isStreaming = false;
                }
                return newHistory;
            });
        }
    });

    const handleSend = async () => {
        if (!prompt.trim()) return;

        const userMsg: ChatMessage = {
            id: Date.now().toString(),
            role: 'user',
            content: prompt,
            timestamp: new Date(),
        };

        setChatHistory(prev => [...prev, userMsg]);
        const currentPrompt = prompt;
        setPrompt('');

        let reasoningData = undefined;

        try {
            if (enableReasoning) {
                toast.info('Analyzing query steps...');
                const reasonRes = await aiApi.reason({
                    prompt: currentPrompt,
                    providerId: selectedProviderId || undefined,
                    maxSteps: 5
                });

                reasoningData = {
                    steps: reasonRes.steps,
                    plan: reasonRes.plan,
                    totalDurationMs: reasonRes.totalDurationMs
                };
            }

            const assistantMsg: ChatMessage = {
                id: (Date.now() + 2).toString(),
                role: 'assistant',
                content: '',
                reasoning: reasoningData,
                isStreaming: true,
                timestamp: new Date(),
            };
            setChatHistory(prev => [...prev, assistantMsg]);

            const context = reasoningData ? { reasoning_plan: reasoningData.plan } : {};
            await startStream(currentPrompt, context, selectedProviderId || undefined);

        } catch (error: any) {
            toast.error(error.message || 'Failed to generate response');
            setChatHistory(prev => {
                const last = prev[prev.length - 1];
                if (last.role === 'assistant' && last.content === '') {
                    return prev.slice(0, -1);
                }
                return prev;
            });
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    const extractSQL = (content: string): string | null => {
        const sqlMatch = content.match(/```sql\n([\s\S]*?)\n```/);
        return sqlMatch ? sqlMatch[1].trim() : null;
    };

    const handleApplySQL = (content: string) => {
        const sql = extractSQL(content);
        if (sql && onApplySQL) {
            onApplySQL(sql);
            toast.success('SQL applied to editor');
        }
    };

    // Quick actions
    const quickActions = [
        { label: 'Explain this SQL', action: () => setPrompt('Explain what this SQL query does:\n\n' + (currentSQL || '')) },
        { label: 'Optimize query', action: () => setPrompt('Optimize this SQL query for better performance:\n\n' + (currentSQL || '')) },
        { label: 'Find errors', action: () => setPrompt('Check this SQL for any errors or issues:\n\n' + (currentSQL || '')) },
        { label: 'Add filters', action: () => setPrompt('Add date range filters to this query:\n\n' + (currentSQL || '')) },
    ];

    return (
        <>
            {/* Floating Button */}
            {!isOpen && (
                <Button
                    onClick={() => setIsOpen(true)}
                    className="fixed bottom-6 right-6 h-14 w-14 rounded-full shadow-lg hover:shadow-xl transition-all z-50"
                    size="icon"
                >
                    <Sparkles className="h-6 w-6" />
                </Button>
            )}

            {/* Chat Panel */}
            <Sheet open={isOpen} onOpenChange={setIsOpen}>
                <SheetContent 
                    side="right" 
                    className={cn(
                        "w-full sm:max-w-xl p-0 flex flex-col",
                        isExpanded && "sm:max-w-2xl"
                    )}
                >
                    {/* Header */}
                    <SheetHeader className="px-4 py-3 border-b flex-shrink-0">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                                <Sparkles className="w-5 h-5 text-primary" />
                                <SheetTitle className="text-base">AI Assistant</SheetTitle>
                            </div>
                            <div className="flex items-center gap-1">
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8"
                                    onClick={() => setIsExpanded(!isExpanded)}
                                >
                                    {isExpanded ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
                                </Button>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8"
                                    onClick={() => setIsOpen(false)}
                                >
                                    <X className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>
                    </SheetHeader>

                    {/* Chat Area */}
                    <ScrollArea className="flex-1 p-4" ref={scrollRef}>
                        <div className="space-y-4">
                            {chatHistory.length === 0 && (
                                <div className="text-center text-muted-foreground py-10">
                                    <Bot className="w-12 h-12 mx-auto mb-4 opacity-50" />
                                    <p className="text-sm font-medium mb-2">How can I help you today?</p>
                                    <p className="text-xs mb-6">Ask me to write SQL, explain queries, or analyze your data</p>
                                    
                                    {/* Quick Actions */}
                                    {currentSQL && (
                                        <div className="flex flex-wrap gap-2 justify-center">
                                            {quickActions.map((action) => (
                                                <Button
                                                    key={action.label}
                                                    variant="outline"
                                                    size="sm"
                                                    className="text-xs"
                                                    onClick={action.action}
                                                >
                                                    {action.label}
                                                </Button>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            )}

                            {chatHistory.map((msg) => (
                                <div key={msg.id} className={cn(
                                    "flex gap-3",
                                    msg.role === 'user' ? "flex-row-reverse" : ""
                                )}>
                                    <div className={cn(
                                        "w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 text-white",
                                        msg.role === 'user' ? "bg-primary" : "bg-emerald-600"
                                    )}>
                                        {msg.role === 'user' ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />}
                                    </div>

                                    <div className="space-y-2 flex-1 min-w-0">
                                        {msg.role === 'assistant' && msg.reasoning && (
                                            <ReasoningDisplay steps={msg.reasoning.steps} plan={msg.reasoning.plan} />
                                        )}

                                        <Card className={cn(
                                            "border-0 shadow-sm",
                                            msg.role === 'user' ? "bg-primary text-primary-foreground ml-auto" : "bg-muted/30"
                                        )}>
                                            <CardContent className="p-3">
                                                {msg.role === 'user' ? (
                                                    <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                                                ) : (
                                                    <div className="prose prose-sm dark:prose-invert max-w-none">
                                                        <ReactMarkdown
                                                            components={{
                                                                code(props) {
                                                                    const { children, className, node, inline, ...rest } = props
                                                                    const match = /language-(\w+)/.exec(className || '')
                                                                    return match ? (
                                                                        <div className="relative">
                                                                            <SyntaxHighlighter
                                                                                style={vscDarkPlus}
                                                                                language={match[1]}
                                                                                PreTag="div"
                                                                                {...rest}
                                                                            >
                                                                                {String(children).replace(/\n$/, '')}
                                                                            </SyntaxHighlighter>
                                                                            {onApplySQL && match[1] === 'sql' && (
                                                                                <Button
                                                                                    size="sm"
                                                                                    variant="secondary"
                                                                                    className="absolute top-2 right-2 text-xs"
                                                                                    onClick={() => handleApplySQL(msg.content)}
                                                                                >
                                                                                    Apply SQL
                                                                                </Button>
                                                                            )}
                                                                        </div>
                                                                    ) : (
                                                                        <code className={className} {...rest}>
                                                                            {children}
                                                                        </code>
                                                                    )
                                                                }
                                                            }}
                                                        >
                                                            {msg.content}
                                                        </ReactMarkdown>
                                                        {msg.isStreaming && (
                                                            <span className="inline-block w-2 h-4 ml-1 align-middle bg-primary animate-pulse" />
                                                        )}
                                                    </div>
                                                )}
                                            </CardContent>
                                        </Card>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </ScrollArea>

                    {/* Input Area */}
                    <div className="p-4 border-t flex-shrink-0 space-y-3">
                        {/* Provider & Reasoning */}
                        <div className="flex items-center gap-2">
                            <Select
                                value={selectedProviderId}
                                onValueChange={setSelectedProviderId}
                                disabled={streamActive}
                            >
                                <SelectTrigger className="w-[140px] h-8 text-xs">
                                    <SelectValue placeholder={defaultProvider?.name || "Provider"} />
                                </SelectTrigger>
                                <SelectContent>
                                    {activeProviders.map((p) => (
                                        <SelectItem key={p.id} value={p.id}>{p.name}</SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>

                            <Button
                                variant={enableReasoning ? "secondary" : "outline"}
                                size="sm"
                                onClick={() => setEnableReasoning(!enableReasoning)}
                                className={cn("text-xs h-8", enableReasoning && "bg-primary/10 text-primary")}
                            >
                                <BrainCircuit className="w-3 h-3 mr-1" />
                                Reasoning
                            </Button>
                        </div>

                        {/* Input */}
                        <div className="flex gap-2 relative">
                            <Textarea
                                value={prompt}
                                onChange={(e) => setPrompt(e.target.value)}
                                onKeyDown={handleKeyDown}
                                placeholder="Ask about your data..."
                                className="min-h-[60px] resize-none pr-12"
                                disabled={streamActive}
                            />
                            <div className="absolute right-2 bottom-2">
                                {streamActive ? (
                                    <Button size="icon" variant="destructive" className="h-8 w-8" onClick={stopGeneration}>
                                        <StopCircle className="w-4 h-4" />
                                    </Button>
                                ) : (
                                    <Button size="icon" className="h-8 w-8" onClick={handleSend} disabled={!prompt.trim()}>
                                        <Send className="w-4 h-4" />
                                    </Button>
                                )}
                            </div>
                        </div>
                        
                        <div className="flex justify-between items-center text-xs text-muted-foreground">
                            <div className="flex gap-2">
                                <Badge variant="outline" className="text-[10px]">GPT-4</Badge>
                                <Badge variant="outline" className="text-[10px]">Streaming</Badge>
                            </div>
                            <span>Enter to send</span>
                        </div>
                    </div>
                </SheetContent>
            </Sheet>
        </>
    );
}

function ReasoningDisplay({ steps, plan }: { steps: { step: number; thought: string; durationMs: number }[], plan?: string }) {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <div className="mb-2">
            <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsOpen(!isOpen)}
                className="h-auto p-0 px-2 py-1 text-xs text-muted-foreground hover:bg-muted/50 w-full flex justify-between border border-dashed border-border rounded-md"
            >
                <div className="flex items-center gap-1.5">
                    <BrainCircuit className="w-3.5 h-3.5" />
                    <span>Thought Process ({steps.length} steps)</span>
                </div>
                {isOpen ? <ChevronDown className="w-3.5 h-3.5" /> : <ChevronRight className="w-3.5 h-3.5" />}
            </Button>

            {isOpen && (
                <div className="mt-2 pl-4 border-l-2 border-muted space-y-3 animate-in slide-in-from-top-2 duration-200">
                    {plan && (
                        <div className="text-xs text-muted-foreground italic mb-2">
                            Planning: {plan}
                        </div>
                    )}
                    {steps.map((step) => (
                        <div key={step.step} className="text-xs">
                            <div className="flex items-center gap-2 mb-1">
                                <span className="font-mono text-[10px] bg-muted px-1.5 py-0.5 rounded">Step {step.step}</span>
                                <span className="text-[10px] text-muted-foreground">{step.durationMs}ms</span>
                            </div>
                            <p className="text-muted-foreground pl-1">{step.thought}</p>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
