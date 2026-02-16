import { useState, useRef, useCallback } from 'react';
import { fetchWithAuth } from '@/lib/utils';
import { toast } from 'sonner';

interface UseAIStreamOptions {
    onToken?: (token: string) => void;
    onComplete?: (fullResponse: string) => void;
    onError?: (error: string) => void;
}

interface StreamMessage {
    role: 'user' | 'assistant' | 'system';
    content: string;
    id?: string;
}

export function useAIStream(options: UseAIStreamOptions = {}) {
    const [messages, setMessages] = useState<StreamMessage[]>([]);
    const [isStreaming, setIsStreaming] = useState(false);
    const [currentResponse, setCurrentResponse] = useState('');
    const abortControllerRef = useRef<AbortController | null>(null);

    const generateStream = useCallback(async (
        prompt: string,
        context: Record<string, any> = {},
        providerId?: string
    ) => {
        // Reset state
        setIsStreaming(true);
        setCurrentResponse('');
        const userMessage: StreamMessage = { role: 'user', content: prompt, id: Date.now().toString() };

        setMessages(prev => [...prev, userMessage]);

        // Create placeholder for assistant message
        const assistantMessageId = (Date.now() + 1).toString();
        setMessages(prev => [...prev, { role: 'assistant', content: '', id: assistantMessageId }]);

        abortControllerRef.current = new AbortController();

        try {
            const response = await fetchWithAuth('/api/go/ai/stream', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ prompt, context, providerId }),
                signal: abortControllerRef.current.signal,
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to start stream');
            }

            if (!response.body) {
                throw new Error('ReadableStream not supported');
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let accumulatedResponse = '';

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value);
                const lines = chunk.split('\n\n');

                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6);

                        if (data === '[DONE]') {
                            break;
                        }

                        try {
                            // Try to parse as JSON first (for error/status messages if any)
                            // But usually it's just raw text or simple tokens
                            // Our backend sends "data: <content>"
                            // Standard SSE sends "data: " prefix.
                            // If the content itself contains newlines, they might be escaped or sent as separate data lines.
                            // Let's assume the backend sends raw text content in the data field.

                            // Check if it's a JSON object with error
                            if (data.startsWith('{') && data.includes('"error"')) {
                                const jsonData = JSON.parse(data);
                                if (jsonData.error) throw new Error(jsonData.error);
                            }

                            const token = data;
                            accumulatedResponse += token;
                            setCurrentResponse(accumulatedResponse);

                            // Update the assistant message in the list
                            setMessages(prev => prev.map(msg =>
                                msg.id === assistantMessageId
                                    ? { ...msg, content: accumulatedResponse }
                                    : msg
                            ));

                            options.onToken?.(token);

                        } catch (e) {
                            // If parsing fails, treat as raw text token
                            const token = data;
                            accumulatedResponse += token;
                            setCurrentResponse(accumulatedResponse);
                            setMessages(prev => prev.map(msg =>
                                msg.id === assistantMessageId
                                    ? { ...msg, content: accumulatedResponse }
                                    : msg
                            ));
                        }
                    }
                }
            }

            options.onComplete?.(accumulatedResponse);

        } catch (error: any) {
            if (error.name === 'AbortError') {
                toast.info('Generation stopped');
            } else {
                const errorMessage = error.message || 'Unknown error';
                toast.error(errorMessage);
                options.onError?.(errorMessage);

                // Update message to show error? Or just leave partial?
                // For now leave partial.
            }
        } finally {
            setIsStreaming(false);
            abortControllerRef.current = null;
        }
    }, [options]);

    const stopStream = useCallback(() => {
        if (abortControllerRef.current) {
            abortControllerRef.current.abort();
            abortControllerRef.current = null;
            setIsStreaming(false);
        }
    }, []);

    const clearMessages = useCallback(() => {
        setMessages([]);
        setCurrentResponse('');
    }, []);

    return {
        messages,
        currentResponse,
        isStreaming,
        generateStream,
        stopStream,
        clearMessages
    };
}
