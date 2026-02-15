export type HelpTopic = {
    id: string;
    title: string;
    content: string; // Markdown content
    links?: { title: string; href: string }[];
};

export const helpTopics: Record<string, HelpTopic> = {
    default: {
        id: 'default',
        title: 'Help & Support',
        content: `
# Welcome to InsightEngine Help

Use the navigation to find specific guides or browse the topics below.

- **Dashboards**: Visualize your data.
- **Data Sources**: Connect databases and APIs.
- **Alerts**: Set up notifications.

Need more? [Visit the Documentation](/docs).
        `,
    },
    '/dashboard': {
        id: 'dashboard',
        title: 'Dashboards Help',
        content: `
# Working with Dashboards

Dashboards allow you to arrange multiple visualizations in a single view.

## Actions
- **Add Card**: Click the "+" button to add a new chart.
- **Edit Layout**: Drag and drop cards to rearrange.
- **Filters**: Use the filter bar to slice data across all charts.
        `,
        links: [
            { title: 'Full Dashboard Guide', href: '/docs/dashboards' },
            { title: 'Video Tutorial', href: '/docs/tutorials' },
        ],
    },
    '/settings/data-sources': {
        id: 'data-sources',
        title: 'Data Sources Help',
        content: `
# Managing Data Sources

Connect your databases to start analyzing data.

## Supported Types
- PostgreSQL
- MySQL
- DuckDB
- SQLite

## Troubleshooting
If a connection fails, check:
1. Hostname and Port reachability.
2. Username and Password correctness.
3. Firewall settings.
        `,
        links: [
            { title: 'Connection Guide', href: '/docs/data-sources' },
        ],
    },
};

export function getHelpTopic(pathname: string): HelpTopic {
    // Exact match
    if (helpTopics[pathname]) {
        return helpTopics[pathname];
    }

    // Prefix match (e.g., /dashboard/123 -> /dashboard)
    const key = Object.keys(helpTopics).find((k) => k !== 'default' && pathname.startsWith(k));
    if (key) {
        return helpTopics[key];
    }

    return helpTopics.default;
}
