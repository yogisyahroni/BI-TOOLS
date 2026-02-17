"use client"

import React, { useCallback, useEffect, useMemo } from 'react';
import ReactFlow, {
    addEdge,
    ConnectionLineType,
    Panel,
    useNodesState,
    useEdgesState,
    ReactFlowProvider,
    Background,
    Controls,
    MiniMap,
    type Node,
    type Edge,
    type Connection,
    MarkerType
} from 'reactflow';
import dagre from 'dagre';
import 'reactflow/dist/style.css';

import { _Card, _CardContent, _CardHeader, _CardTitle } from "@/components/ui/card";
import { _Badge } from "@/components/ui/badge";

interface _LineageNodeData {
    label: string;
    type: string;
}

interface LineageGraphProps {
    data: {
        nodes: { id: string; type: string; label: string }[];
        edges: { id: string; source: string; target: string }[];
    };
}

const nodeWidth = 172;
const nodeHeight = 36;

const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));

    const isHorizontal = direction === 'LR';
    dagreGraph.setGraph({ rankdir: direction });

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
    });

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    nodes.forEach((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        node.targetPosition = isHorizontal ? 'left' : 'top' as any;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        node.sourcePosition = isHorizontal ? 'right' : 'bottom' as any;

        // We are shifting the dagre node position (anchor=center center) to the top left
        // so it matches the React Flow node anchor point (top left).
        node.position = {
            x: nodeWithPosition.x - nodeWidth / 2,
            y: nodeWithPosition.y - nodeHeight / 2,
        };

        return node;
    });

    return { nodes, edges };
};

const getNodeColor = (type: string) => {
    switch (type) {
        case 'source': return '#10b981'; // Green
        case 'table': return '#3b82f6'; // Blue
        case 'query': return '#8b5cf6'; // Purple
        case 'dashboard': return '#f59e0b'; // Orange
        default: return '#64748b';
    }
}

const LineageGraphRaw: React.FC<LineageGraphProps> = ({ data }) => {
    const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(() => {
        const initialNodes: Node[] = data.nodes.map(n => ({
            id: n.id,
            data: { label: n.label, type: n.type },
            position: { x: 0, y: 0 },
            style: {
                background: '#fff',
                border: `1px solid ${getNodeColor(n.type)}`,
                borderRadius: '5px',
                padding: '10px',
                fontSize: '12px',
                fontWeight: 'bold',
                width: nodeWidth,
                boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
            },
            type: 'default' // Using default node for now
        }));

        const initialEdges: Edge[] = data.edges.map(e => ({
            id: e.id,
            source: e.source,
            target: e.target,
            type: 'smoothstep',
            animated: true,
            markerEnd: { type: MarkerType.ArrowClosed, color: '#b1b1b7' },
            style: { stroke: '#b1b1b7' }
        }));

        return getLayoutedElements(initialNodes, initialEdges);
    }, [data]);

    const [nodes, setNodes, onNodesChange] = useNodesState(layoutedNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(layoutedEdges);

    // Update layout when data changes
    useEffect(() => {
        setNodes(layoutedNodes);
        setEdges(layoutedEdges);
    }, [layoutedNodes, layoutedEdges, setNodes, setEdges]);


    const onConnect = useCallback(
        (params: Connection) => setEdges((eds) => addEdge({ ...params, type: ConnectionLineType.SmoothStep, animated: true }, eds)),
        [setEdges]
    );

    return (
        <div className="h-[600px] w-full border rounded-lg bg-slate-50 relative">
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                onConnect={onConnect}
                fitView
                attributionPosition="bottom-right"
            >
                <MiniMap
                    nodeColor={(n) => getNodeColor(n.data.type)}
                    maskColor="rgba(240, 240, 240, 0.6)"
                />
                <Controls />
                <Background gap={12} size={1} />
                <Panel position="top-left" className="bg-white p-2 rounded shadow-md border flex gap-2">
                    <div className="flex items-center gap-1"><div className="w-3 h-3 rounded-full bg-emerald-500" /><span className="text-xs">Source</span></div>
                    <div className="flex items-center gap-1"><div className="w-3 h-3 rounded-full bg-blue-500" /><span className="text-xs">Table</span></div>
                    <div className="flex items-center gap-1"><div className="w-3 h-3 rounded-full bg-purple-500" /><span className="text-xs">Query</span></div>
                    <div className="flex items-center gap-1"><div className="w-3 h-3 rounded-full bg-amber-500" /><span className="text-xs">Dashboard</span></div>
                </Panel>
            </ReactFlow>
        </div>
    );
};

export const LineageGraph = (props: LineageGraphProps) => (
    <ReactFlowProvider>
        <LineageGraphRaw {...props} />
    </ReactFlowProvider>
);
