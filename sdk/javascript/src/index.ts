export class InsightEmbed {
    private options: {
        url: string;
        token: string;
        dashboardId: string;
        container: HTMLElement;
        width?: string;
        height?: string;
    };
    private iframe: HTMLIFrameElement | null = null;

    constructor(options: {
        url: string;
        token: string;
        dashboardId: string;
        container: HTMLElement;
        width?: string;
        height?: string;
    }) {
        this.options = options;
    }

    render() {
        if (!this.options.container) {
            throw new Error('Container element is required');
        }

        // Clear container
        this.options.container.innerHTML = '';

        // Create iframe
        this.iframe = document.createElement('iframe');
        const embedUrl = `${this.options.url}/embed/dashboard/${this.options.dashboardId}?token=${this.options.token}`;

        this.iframe.src = embedUrl;
        this.iframe.style.width = this.options.width || '100%';
        this.iframe.style.height = this.options.height || '100%';
        this.iframe.style.border = 'none';
        this.iframe.setAttribute('allow', 'clipboard-write');

        this.options.container.appendChild(this.iframe);

        // Listen for resize events from the embedded dashboard
        window.addEventListener('message', this.handleMessage.bind(this));
    }

    private handleMessage(event: MessageEvent) {
        // Verify origin if possible, but options.url might be relative or different
        // For now, accept all or check against options.url hostname

        const { type, payload } = event.data;
        if (type === 'RESIZE') {
            if (this.iframe && !this.options.height) {
                this.iframe.style.height = `${payload.height}px`;
            }
        }
    }

    setFilter(key: string, value: any) {
        if (this.iframe && this.iframe.contentWindow) {
            this.iframe.contentWindow.postMessage(
                { type: 'SET_FILTER', payload: { key, value } },
                '*'
            );
        }
    }

    destroy() {
        window.removeEventListener('message', this.handleMessage.bind(this));
        if (this.iframe) {
            this.iframe.remove();
            this.iframe = null;
        }
    }
}

// Export for UMD/CommonJS/ESM
if (typeof window !== 'undefined') {
    (window as any).InsightEmbed = InsightEmbed;
}
