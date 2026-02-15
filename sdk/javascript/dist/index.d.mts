declare class InsightEmbed {
    private options;
    private iframe;
    constructor(options: {
        url: string;
        token: string;
        dashboardId: string;
        container: HTMLElement;
        width?: string;
        height?: string;
    });
    render(): void;
    private handleMessage;
    setFilter(key: string, value: any): void;
    destroy(): void;
}

export { InsightEmbed };
