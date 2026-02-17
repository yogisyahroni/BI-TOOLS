'use client';

import React from 'react';
import { EChartsWrapper } from './echarts-wrapper';
import type { EChartsOption } from 'echarts';

/**
 * TASK-CHART-017: Calendar Heatmap
 * GitHub-style contribution/activity calendar showing daily intensity over a year.
 */

interface CalendarDataPoint {
    date: string; // YYYY-MM-DD
    value: number;
}

interface CalendarHeatmapProps {
    data: CalendarDataPoint[];
    year: number;
    title?: string;
    colorRange?: string[];
    maxValue?: number;
    className?: string;
    onDayClick?: (date: string, value: number) => void;
}

export function CalendarHeatmap({
    data,
    year,
    title,
    colorRange = ['#ebedf0', '#9be9a8', '#40c463', '#30a14e', '#216e39'],
    maxValue: providedMax,
    className = 'h-full w-full min-h-[220px]',
    onDayClick,
}: CalendarHeatmapProps) {
    const maxVal = providedMax ?? Math.max(...data.map(d => d.value), 1);

    const option: EChartsOption = {
        title: title
            ? { text: title, left: 'center', top: 0, textStyle: { fontSize: 14, fontWeight: 600 } }
            : undefined,
        tooltip: {
            trigger: 'item',
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            formatter: (params: any) => {
                const d = params.data;
                if (!d || !Array.isArray(d)) return '';
                return `${d[0]}<br/>Value: <strong>${d[1].toLocaleString()}</strong>`;
            },
        },
        visualMap: {
            min: 0,
            max: maxVal,
            calculable: false,
            orient: 'horizontal',
            left: 'center',
            bottom: 0,
            inRange: { color: colorRange },
            textStyle: { fontSize: 10 },
            itemWidth: 12,
            itemHeight: 80,
            show: true,
        },
        calendar: {
            top: title ? 50 : 30,
            left: 40,
            right: 20,
            bottom: 40,
            range: String(year),
            cellSize: ['auto', 14],
            yearLabel: { show: false },
            monthLabel: {
                fontSize: 10,
                color: '#888',
            },
            dayLabel: {
                firstDay: 1,
                fontSize: 9,
                color: '#aaa',
                nameMap: ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'],
            },
            itemStyle: {
                borderWidth: 2,
                borderColor: 'transparent',
                color: colorRange[0],
            },
            splitLine: { show: false },
        },
        series: [
            {
                type: 'heatmap',
                coordinateSystem: 'calendar',
                data: data.map(d => [d.date, d.value]),
                emphasis: {
                    itemStyle: {
                        borderColor: '#333',
                        borderWidth: 1,
                    },
                },
                itemStyle: {
                    borderRadius: 2,
                },
            },
        ],
        animationDuration: 500,
    };

    const events = onDayClick
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ? {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
            click: (params: any) => {
                if (params.data && Array.isArray(params.data)) {
                    onDayClick(params.data[0], params.data[1]);
                }
            },
        }
        : undefined;

    return <EChartsWrapper options={option} className={className} onEvents={events} />;
}
