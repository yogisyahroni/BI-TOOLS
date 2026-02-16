import type { EChartsOption } from 'echarts';
import { type VisualizationConfig } from '@/lib/types';
import { type FilterCriteria } from '@/lib/cross-filter-context';

// Default semantic colors (Insight Engine Palette)
const DEFAULT_COLORS = [
    '#3b82f6', // blue-500
    '#ef4444', // red-500
    '#10b981', // emerald-500
    '#f59e0b', // amber-500
    '#8b5cf6', // violet-500
    '#ec4899', // pink-500
];

export function buildEChartsOptions(
    data: Record<string, any>[],
    config: VisualizationConfig,
    theme: string = 'light',
    activeFilters?: FilterCriteria[],
    chartId?: string
): EChartsOption {
    const { type, xAxis, yAxis, title } = config;

    if (type === 'table') return {};

    // 0. Determine if this chart is the source of any active filter
    let highlightedValue: string | number | null = null;
    let highlightedField: string | null = null;

    if (activeFilters && chartId) {
        const myFilter = activeFilters.find(f => f.sourceChartId === chartId); // currently only support single selection per chart
        if (myFilter) {
            highlightedValue = myFilter.value;
            highlightedField = myFilter.fieldName;
        }
    }

    const getItemStyle = (dataItem: any) => {
        if (!highlightedValue || !highlightedField) return {};

        // Check if this item matches the filter
        const itemValue = dataItem[highlightedField];
        const isSelected = itemValue === highlightedValue; // loose equality for string/number mix

        return {
            opacity: isSelected ? 1 : 0.3,
            shadowBlur: isSelected ? 10 : 0,
            shadowColor: isSelected ? 'rgba(0,0,0,0.3)' : 'transparent'
        };
    };


    // Base configuration
    const options: EChartsOption = {
        title: {
            text: title,
            left: 'center',
            textStyle: {
                color: theme === 'dark' ? '#fff' : '#333'
            }
        },
        tooltip: {
            trigger: (type === 'pie' || type === 'funnel') ? 'item' : 'axis',
            axisPointer: {
                type: 'cross'
            },
            formatter: config.tooltipTemplate ? (params: any) => {
                const template = config.tooltipTemplate || '';

                const replaceVars = (str: string, item: any) => {
                    let res = str;
                    const name = item.name || item.axisValueLabel || '';
                    const value = item.value;
                    const seriesName = item.seriesName || '';
                    const color = item.color;

                    // Format value if numeric
                    let formattedValue = value;
                    if (!isNaN(Number(value))) {
                        if (config.yAxisFormat === 'currency') {
                            formattedValue = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(Number(value));
                        } else if (config.yAxisFormat === 'percent') {
                            formattedValue = `${Number(value)}%`;
                        } else {
                            formattedValue = new Intl.NumberFormat('en-US').format(Number(value));
                        }
                    }

                    res = res.replace(/\{\{name\}\}/g, name);
                    res = res.replace(/\{\{value\}\}/g, String(formattedValue));
                    res = res.replace(/\{\{series\}\}/g, seriesName);
                    // res = res.replace(/\{\{color\}\}/g, color); // TODO: Support color dot?
                    return res;
                };

                // Handle Array (Axis Trigger) vs Object (Item Trigger)
                if (Array.isArray(params)) {
                    // Axis Trigger: Header + List of items
                    // Default behavior for template: Repeat template for each item
                    // But usually you want a Header.
                    // For now, let's just join them with <br/>
                    return params.map(p => {
                        // Add a marker for visual consistence if not in template?
                        // Let's rely on user template entirely.
                        return replaceVars(template, p);
                    }).join('<br/>');
                } else {
                    return replaceVars(template, params);
                }
            } : undefined, // Fallback to default if no template

            valueFormatter: config.tooltipTemplate ? undefined : (value: any) => { // Disable if custom formatter used ? No, valueFormatter is specifically for axis pointer label usually or default content.
                const numVal = Number(value);
                if (isNaN(numVal)) return value as string;
                if (config.yAxisFormat === 'currency') {
                    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(numVal);
                }
                if (config.yAxisFormat === 'percent') {
                    return `${numVal}%`;
                }
                return new Intl.NumberFormat('en-US').format(numVal);
            }
        },
        toolbox: {
            feature: {
                dataZoom: {
                    yAxisIndex: 'none'
                },
                restore: {},
                saveAsImage: {}
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        color: config.colors || DEFAULT_COLORS,
    };

    // 1. Pie & Funnel Charts (Single Series, Non-Cartesian)
    if (type === 'pie' || type === 'funnel') {
        const seriesData = data.map(item => ({
            name: item[xAxis],
            value: item[yAxis[0]],
            itemStyle: getItemStyle({ [xAxis]: item[xAxis] }) // pass object to mimic data item structure if needed, or just pass item if we use item[field]
        }));

        // Correction: getItemStyle needs the whole item or we just pass the value logic here
        // Re-map with itemStyle:
        const processedSeriesData = data.map(item => {
            const isSelected = (!highlightedValue) || (item[xAxis] === highlightedValue);
            return {
                name: item[xAxis],
                value: item[yAxis[0]],
                itemStyle: {
                    opacity: isSelected ? 1 : 0.3
                }
            };
        });


        options.series = [{
            name: title || 'Data',
            type: type as 'pie' | 'funnel',
            radius: type === 'pie' ? '50%' : undefined,
            width: type === 'funnel' ? '60%' : undefined,
            left: type === 'funnel' ? 'center' : undefined,
            data: processedSeriesData,
            emphasis: {
                itemStyle: {
                    shadowBlur: 10,
                    shadowOffsetX: 0,
                    shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
            }
        }];

        // For funnel, we might want to sort
        if (type === 'funnel') {
            (options.series[0] as any).sort = 'descending';
        }

        return options;
    }

    // 2. Cartesian Charts (Bar, Line, Scatter, Area, Combo)

    let echartsType: 'bar' | 'line' | 'scatter' = 'bar';
    let areaStyle: any = undefined;

    if (type === 'line' || type === 'area') {
        echartsType = 'line';
        if (type === 'area') {
            areaStyle = { opacity: 0.5 };
        }
    } else if (type === 'scatter') {
        echartsType = 'scatter';
    } else if (type === 'bar') {
        echartsType = 'bar';
    }

    // X-Axis Config
    options.xAxis = {
        type: 'category',
        boundaryGap: type === 'bar' || type === 'combo',
        data: data.map(item => item[xAxis]),
        axisLabel: {
            color: theme === 'dark' ? '#ccc' : '#666',
            rotate: data.length > 20 ? 45 : 0
        }
    };

    // Y-Axis Config
    if (type === 'combo') {
        // Dual Axis for Combo
        options.yAxis = [
            {
                type: 'value',
                name: yAxis[0],
                position: 'left',
                axisLabel: { color: theme === 'dark' ? '#ccc' : '#666' }
            },
            {
                type: 'value',
                name: yAxis[1] || '',
                position: 'right',
                axisLabel: { color: theme === 'dark' ? '#ccc' : '#666' }
            }
        ];
    } else {
        // Single Y-Axis (can be multiple series on same axis)
        options.yAxis = yAxis.map(axisKey => ({
            type: 'value',
            name: config.yAxisLabel || axisKey,
            axisLabel: {
                color: theme === 'dark' ? '#ccc' : '#666',
                formatter: (value: number) => {
                    if (config.yAxisFormat === 'currency') {
                        return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', notation: 'compact' }).format(value);
                    }
                    if (config.yAxisFormat === 'percent') {
                        return `${value}%`;
                    }
                    return new Intl.NumberFormat('en-US', { notation: 'compact', compactDisplay: 'short' }).format(value);
                }
            }
        }));
    }

    // Prepare visualMap override or individual item styles?
    // For Bar charts, we can pass data as objects { value: X, itemStyle: ... }
    // For Line charts, dimming segments is hard. Usually we just dim the whole line if it doesn't match? 
    // BUT here we filter by X-axis (category). So distinct points.
    // Line chart points can be styled.

    // Check if we have active filter on X-axis
    const isXAxisFiltered = highlightedField === xAxis;

    const getDataWithStyle = (valueCol: string) => {
        return data.map(item => {
            const val = item[valueCol];
            if (!isXAxisFiltered || !highlightedValue) return val; // No filter or filter not on X-axis

            const isSelected = item[xAxis] === highlightedValue;
            return {
                value: val,
                itemStyle: {
                    opacity: isSelected ? 1 : 0.3
                }
            };
        });
    };


    // Series Config
    if (type === 'combo' && yAxis.length >= 2) {
        options.series = [
            {
                name: yAxis[0],
                type: 'bar',
                yAxisIndex: 0,
                data: getDataWithStyle(yAxis[0])
            },
            {
                name: yAxis[1],
                type: 'line',
                yAxisIndex: 1,
                data: getDataWithStyle(yAxis[1]),
                smooth: true
            }
        ];
    } else {
        // Split data into Historical and Forecast
        const historicalData = data.filter(d => !d._isForecast);
        const forecastData = data.filter(d => d._isForecast);

        options.series = yAxis.map((axisKey, index) => {
            const series: any[] = [];

            // 1. Historical Data Series
            series.push({
                name: axisKey,
                type: echartsType,
                // Use getDataWithStyle but restricted to historical? 
                // Easiest is to map main data and rely on nulls, but we need style objects.
                // data: data.map(item => !item._isForecast ? item[axisKey] : null),  <-- OLD

                data: data.map(item => {
                    if (item._isForecast) return null;
                    const val = item[axisKey];
                    // Apply style if filtered
                    if (isXAxisFiltered && highlightedValue) {
                        const isSelected = item[xAxis] === highlightedValue;
                        return {
                            value: val,
                            itemStyle: { opacity: isSelected ? 1 : 0.3 }
                        };
                    }
                    return val;
                }),

                smooth: true,
                symbolSize: echartsType === 'scatter' ? 10 : 4,
                areaStyle: areaStyle,
                // Ensure connections works if needed, usually null breaks line which is what we want if treating as separate
            });

            // 2. Forecast Data Series (Only for Line/Area/Bar, usually Line)
            if (forecastData.length > 0) {
                // To connect the lines, the forecast series should ideally start with the last historical point.
                // Or we rely on visual proximity. ECharts handles 'null' by breaking line.
                // We want the lines to connect.
                // Strategy: Forecast Series includes the LAST historical point.

                const lastHistorical = historicalData[historicalData.length - 1];

                const forecastSeriesData = data.map(item => {
                    // Logic: Must be forecast OR be the last historical point
                    const isTarget = item._isForecast || item === lastHistorical;
                    if (!isTarget) return null;

                    const val = item[axisKey];

                    // Apply style if filtered
                    if (isXAxisFiltered && highlightedValue) {
                        const isSelected = item[xAxis] === highlightedValue;
                        return {
                            value: val,
                            itemStyle: { opacity: isSelected ? 1 : 0.3 }
                        };
                    }
                    return val;
                });

                series.push({
                    name: `${axisKey} (Forecast)`,
                    type: 'line', // Forecast is always line usually
                    data: forecastSeriesData,
                    smooth: true,
                    lineStyle: { type: 'dashed' },
                    itemStyle: { opacity: 0.7 },
                    symbol: 'emptyCircle'
                });
            }

            return series;
        }).flat();
    }

    // DataZoom
    if (data.length > 50) {
        options.dataZoom = [
            {
                type: 'slider',
                show: true,
                xAxisIndex: [0],
                start: 0,
                end: Math.min(100, (50 / data.length) * 100)
            },
            {
                type: 'inside',
                xAxisIndex: [0],
                start: 0,
                end: 100
            }
        ];
    }

    return options;
}
