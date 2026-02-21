"use client";

import { useState, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import {
  Sparkles,
  Send,
  User,
  Bot,
  Loader2,
  RefreshCw,
  StopCircle,
  ChevronDown,
  ChevronRight,
  BrainCircuit,
  CheckCircle2,
} from "lucide-react";
import { useAIStream } from "@/hooks/use-ai-stream";
import { useAIProviders } from "@/hooks/use-ai-providers";
import { aiApi } from "@/lib/api/ai";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import ReactMarkdown from "react-markdown";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { vscDarkPlus } from "react-syntax-highlighter/dist/esm/styles/prism";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

// Type for a message in the chat UI
interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  reasoning?: {
    steps: { step: number; thought: string; durationMs: number }[];
    plan?: string;
    totalDurationMs: number;
  };
  isStreaming?: boolean;
  timestamp: Date;
}

export function AIChat() {
  const [prompt, setPrompt] = useState("");
  const [selectedProviderId, setSelectedProviderId] = useState<string>("");
  const [enableReasoning, setEnableReasoning] = useState(false);
  const [chatHistory, setChatHistory] = useState<ChatMessage[]>([]);
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
  const {
    generateStream,
    messages: _streamMessages,
    isStreaming,
    stopStream,
  } = useAIStream({
    onComplete: (fullResponse) => {
      // Update the last assistant message with full response and mark streaming as done
      setChatHistory((prev) => {
        const lastMsg = prev[prev.length - 1];
        if (lastMsg && lastMsg.role === "assistant") {
          return [...prev.slice(0, -1), { ...lastMsg, content: fullResponse, isStreaming: false }];
        }
        return prev;
      });
    },
    onError: (err) => {
      setChatHistory((prev) => {
        const lastMsg = prev[prev.length - 1];
        if (lastMsg && lastMsg.role === "assistant") {
          return [
            ...prev.slice(0, -1),
            { ...lastMsg, content: lastMsg.content + `\n\n**Error:** ${err}`, isStreaming: false },
          ];
        }
        return prev;
      });
    },
  });

  // Update chat history from stream messages logic
  // Actually use-ai-stream keeps its own 'messages' state which is transient.
  // We want to persist specific chat history.
  // Let's sync them or just manage history here and feed prompt to hook?
  // The hook allows simple generation.

  // Better approach:
  // 1. User clicks send.
  // 2. Add user message to history.
  // 3. Add temporary assistant message to history (loading/streaming).
  // 4. Call `generateStream`.
  // 5. Hook calls `onToken`. We update the last message in history.

  // Let's re-instantiate hook with onToken to update local state efficiently.
  const {
    generateStream: startStream,
    isStreaming: streamActive,
    stopStream: stopGeneration,
  } = useAIStream({
    onToken: (token) => {
      setChatHistory((prev) => {
        const newHistory = [...prev];
        const lastMsg = newHistory[newHistory.length - 1];
        if (lastMsg.role === "assistant") {
          lastMsg.content += token;
          lastMsg.isStreaming = true; // Still streaming
        }
        return newHistory;
      });
    },
    onComplete: () => {
      setChatHistory((prev) => {
        const newHistory = [...prev];
        const lastMsg = newHistory[newHistory.length - 1];
        if (lastMsg.role === "assistant") {
          lastMsg.isStreaming = false;
        }
        return newHistory;
      });
    },
  });

  const handleSend = async () => {
    if (!prompt.trim()) return;

    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      role: "user",
      content: prompt,
      timestamp: new Date(),
    };

    setChatHistory((prev) => [...prev, userMsg]);
    const currentPrompt = prompt;
    setPrompt("");

    let reasoningData = undefined;

    try {
      // 1. If Reasoning Enabled: Call Reason Endpoint First
      if (enableReasoning) {
        const _loadingId = Date.now() + 1;
        // Add placeholder for reasoning
        // Actually better to have the assistant message contain the reasoning data
        // So we start the assistant message now with empty content but potentially reasoning data loading...

        toast.info("Analyzing query steps...");

        const reasonRes = await aiApi.reason({
          prompt: currentPrompt,
          providerId: selectedProviderId || undefined,
          maxSteps: 5,
        });

        reasoningData = {
          steps: reasonRes.steps,
          plan: reasonRes.plan,
          totalDurationMs: reasonRes.totalDurationMs,
        };
      }

      // 2. Add Assistant Message Placeholder
      const assistantMsg: ChatMessage = {
        id: (Date.now() + 2).toString(),
        role: "assistant",
        content: "",
        reasoning: reasoningData,
        isStreaming: true,
        timestamp: new Date(),
      };
      setChatHistory((prev) => [...prev, assistantMsg]);

      // 3. Start Stream (with reasoning plan as context if available)
      const context = reasoningData ? { reasoning_plan: reasoningData.plan } : {};

      await startStream(currentPrompt, context, selectedProviderId || undefined);

      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (error: any) {
      toast.error(error.message || "Failed to generate response");
      setChatHistory((prev) => {
        // Remove the placeholder if it exists and is empty/streaming
        const last = prev[prev.length - 1];
        if (last.role === "assistant" && last.content === "") {
          return prev.slice(0, -1);
        }
        return prev;
      });
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const toggleReasoning = () => setEnableReasoning(!enableReasoning);

  return (
    <div className="flex flex-col h-[calc(100vh-12rem)] min-h-[500px] border rounded-lg bg-background overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b bg-muted/20">
        <div className="flex items-center gap-2">
          <Sparkles className="w-5 h-5 text-primary" />
          <h3 className="font-semibold">AI Assistant</h3>
        </div>
        <div className="flex items-center gap-2">
          {/* Provider Selector */}
          <Select
            value={selectedProviderId}
            onValueChange={setSelectedProviderId}
            disabled={streamActive}
          >
            <SelectTrigger className="w-[180px] h-8 text-xs">
              <SelectValue placeholder={defaultProvider?.name || "Select Provider"} />
            </SelectTrigger>
            <SelectContent>
              {activeProviders.map((p) => (
                <SelectItem key={p.id} value={p.id}>
                  {p.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Button
            variant={enableReasoning ? "secondary" : "ghost"}
            size="sm"
            onClick={toggleReasoning}
            className={cn(
              "gap-1 text-xs",
              enableReasoning && "bg-primary/10 text-primary hover:bg-primary/20",
            )}
            title="Enable Multi-Step Reasoning"
          >
            <BrainCircuit className="w-3 h-3" />
            Reasoning
          </Button>
        </div>
      </div>

      {/* Chat Area */}
      <ScrollArea className="flex-1 p-4" ref={scrollRef}>
        <div className="space-y-6">
          {chatHistory.length === 0 && (
            <div className="text-center text-muted-foreground mt-20">
              <Bot className="w-12 h-12 mx-auto mb-4 opacity-50" />
              <p className="text-sm">How can I help you regarding your data today?</p>
              <p className="text-xs mt-1">
                Try asking for SQL queries, explanations, or data insights.
              </p>
            </div>
          )}

          {chatHistory.map((msg) => (
            <div
              key={msg.id}
              className={cn(
                "flex gap-3 max-w-[85%]",
                msg.role === "user" ? "ml-auto flex-row-reverse" : "",
              )}
            >
              <div
                className={cn(
                  "w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0 text-white",
                  msg.role === "user" ? "bg-primary" : "bg-emerald-600",
                )}
              >
                {msg.role === "user" ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />}
              </div>

              <div className="space-y-2 w-full">
                {/* Reasoning Accordion (Only for Assistant) */}
                {msg.role === "assistant" && msg.reasoning && (
                  <ReasoningDisplay steps={msg.reasoning.steps} plan={msg.reasoning.plan} />
                )}

                <Card
                  className={cn(
                    "p-4 border-0 shadow-sm",
                    msg.role === "user" ? "bg-primary text-primary-foreground" : "bg-muted/30",
                  )}
                >
                  {msg.role === "user" ? (
                    <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                  ) : (
                    <div className="prose prose-sm dark:prose-invert max-w-none">
                      <ReactMarkdown
                        components={{
                          code(props) {
                            const { children, className, node, inline, ...rest } = props;
                            const match = /language-(\w+)/.exec(className || "");
                            return match ? (
                              <SyntaxHighlighter
                                // @ts-expect-error
                                style={vscDarkPlus}
                                language={match[1]}
                                PreTag="div"
                                {...rest}
                              >
                                {String(children).replace(/\n$/, "")}
                              </SyntaxHighlighter>
                            ) : (
                              <code className={className} {...rest}>
                                {children}
                              </code>
                            );
                          },
                        }}
                      >
                        {msg.content}
                      </ReactMarkdown>
                      {msg.isStreaming && (
                        <span className="inline-block w-2 h-4 ml-1 align-middle bg-primary animate-pulse" />
                      )}
                    </div>
                  )}
                </Card>
              </div>
            </div>
          ))}
        </div>
      </ScrollArea>

      {/* Input Area */}
      <div className="p-4 border-t bg-background">
        <div className="flex gap-2 relative">
          <Textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask a question about your data..."
            className="min-h-[60px] resize-none pr-12"
            disabled={streamActive}
          />
          <div className="absolute right-3 bottom-3 flex gap-2">
            {streamActive ? (
              <Button
                size="icon"
                variant="destructive"
                className="h-8 w-8"
                onClick={stopGeneration}
              >
                <StopCircle className="w-4 h-4" />
              </Button>
            ) : (
              <Button
                size="icon"
                className="h-8 w-8"
                onClick={handleSend}
                disabled={!prompt.trim()}
              >
                <Send className="w-4 h-4" />
              </Button>
            )}
          </div>
        </div>
        <div className="flex justify-between items-center mt-2 text-xs text-muted-foreground">
          <div className="flex gap-2">
            <Badge variant="outline" className="text-[10px] font-normal">
              GPT-4 Supported
            </Badge>
            <Badge variant="outline" className="text-[10px] font-normal">
              Streaming Enabled
            </Badge>
          </div>
          <span>Press Enter to send, Shift+Enter for new line</span>
        </div>
      </div>
    </div>
  );
}

// Sub-component for Reasoning Display
function ReasoningDisplay({
  steps,
  plan,
}: {
  steps: { step: number; thought: string; durationMs: number }[];
  plan?: string;
}) {
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
        {isOpen ? (
          <ChevronDown className="w-3.5 h-3.5" />
        ) : (
          <ChevronRight className="w-3.5 h-3.5" />
        )}
      </Button>

      {isOpen && (
        <div className="mt-2 pl-4 border-l-2 border-muted space-y-3 animate-in slide-in-from-top-2 duration-200">
          {plan && (
            <div className="text-xs text-muted-foreground italic mb-2">Planning: {plan}</div>
          )}
          {steps.map((step) => (
            <div key={step.step} className="text-xs">
              <div className="flex items-center gap-2 mb-1">
                <span className="font-mono text-[10px] bg-muted px-1.5 py-0.5 rounded text-foreground">
                  Step {step.step}
                </span>
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
