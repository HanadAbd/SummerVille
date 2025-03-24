type Pos = { 
    x: number; 
    y: number; 
    z: number;
};

type Node = {
    id: string;
    name: string;
    event: string;
    nextNodes: string[];
    nodesWithin: string[];
    processingTime: number;
    position?: {
        x: number;
        y: number;
    };
};

type Edge = {
    source: string;
    target: string;
    color: string;
    width: number;
};

type GraphData = {
    nodes: {[key: string]: Node};
    layout?: {
        width: number;
        height: number;
    };
};

export class Graph {
    canvas: HTMLCanvasElement;
    ctx: CanvasRenderingContext2D;
    homeButton: HTMLButtonElement;
    tooltip: HTMLElement;
    parentContainer: HTMLElement;
    graphData: GraphData | null;
    nodePositions: Map<string, {x: number, y: number}>;
    nodeElements: Map<string, {width: number, height: number}>;
    eventColors: {[key: string]: string};
    hoverNode: Node | null;
    legendVisible: boolean = true;

    partsInTransit: Map<string, {from: string, to: string, progress: number}> = new Map();
    partStates: Map<string, {nodeId: string, state: string}> = new Map();
    nodeQueues: Map<string, string[]> = new Map();

    private offsetX: number;
    private offsetY: number;
    private scale: number;
    private isDragging: boolean;
    private lastX: number;
    private lastY: number;
    

    constructor(canvas: HTMLCanvasElement, homeButton: HTMLButtonElement, tooltip: HTMLElement) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d') as CanvasRenderingContext2D;
        this.homeButton = homeButton;
        this.tooltip = tooltip;
        this.parentContainer = canvas.parentElement as HTMLElement;
        
        this.offsetX = 0;
        this.offsetY = 0;
        this.scale = 1;
        this.isDragging = false;
        this.lastX = 0;
        this.lastY = 0;
        this.graphData = null;
        this.nodePositions = new Map();
        this.nodeElements = new Map();
        this.eventColors = {
            'Idle': '#45B7D1',
            'Processing': '#34A853',
            'Processed': '#FBBC05',
            'Faulty': '#EA4335'
        };
        this.hoverNode = null;
        
        this.partsInTransit = new Map();
        this.partStates = new Map();
        this.nodeQueues = new Map();

        this.addEventListeners();
        this.setCanvasSize(this.parentContainer);
        this.startAnimationLoop();
    }

    setData(data: GraphData): void {
        this.graphData = data;
        this.calculateNodePositions();
        this.renderCanvas();
    }

    initializeElements(): void {
        this.offsetX = 0;
        this.offsetY = 0;
        this.scale = 1;
        this.isDragging = false;
        this.lastX = 0;
        this.lastY = 0;
        this.hoverNode = null;
    }

    setCanvasSize(parentContainer: HTMLElement): void {
        const width = parentContainer.clientWidth;
        const height = parentContainer.clientHeight;
        this.canvas.width = width;
        this.canvas.height = height;
        
        if (this.graphData) {
            this.calculateNodePositions();
        }
        
        this.renderCanvas();
    }

    calculateNodePositions(): void {
        if (!this.graphData) return;

        const nodes = this.graphData.nodes;
        const nodeIds = Object.keys(nodes);
        
        // Filter top-level nodes (those that aren't contained within other nodes)
        const topLevelNodeIds = nodeIds.filter(id => {
            return !nodeIds.some(otherId => 
                nodes[otherId].nodesWithin && 
                nodes[otherId].nodesWithin.includes(id)
            );
        });
        
        const width = this.canvas.width;
        const height = this.canvas.height;
        const padding = 50;
        const availableWidth = width - padding * 2;
        const availableHeight = height - padding * 2;
        
        // Position top-level nodes in a grid layout
        const cols = Math.ceil(Math.sqrt(topLevelNodeIds.length));
        const rows = Math.ceil(topLevelNodeIds.length / cols);
        const cellWidth = availableWidth / cols;
        const cellHeight = availableHeight / rows;
        
        this.nodePositions.clear();
        this.nodeElements.clear();

        topLevelNodeIds.forEach((id, index) => {
            const row = Math.floor(index / cols);
            const col = index % cols;
            
            const x = padding + col * cellWidth + cellWidth / 2;
            const y = padding + row * cellHeight + cellHeight / 2;
            
            // Store the position
            this.nodePositions.set(id, { x, y });
            
            // Set size based on whether the node contains other nodes
            const hasContainedNodes = nodes[id].nodesWithin && nodes[id].nodesWithin.length > 0;
            const nodeWidth = hasContainedNodes ? cellWidth * 0.8 : cellWidth * 0.4;
            const nodeHeight = hasContainedNodes ? cellHeight * 0.8 : cellHeight * 0.4;
            
            this.nodeElements.set(id, { width: nodeWidth, height: nodeHeight });
            
            // Position contained nodes within their parent
            if (hasContainedNodes) {
                const containedNodes = nodes[id].nodesWithin;
                const containerPadding = 20;
                const innerWidth = nodeWidth - containerPadding * 2;
                const innerHeight = nodeHeight - containerPadding * 2;
                
                // Position in a grid inside the parent
                const innerCols = Math.ceil(Math.sqrt(containedNodes.length));
                const innerRows = Math.ceil(containedNodes.length / innerCols);
                const innerCellWidth = innerWidth / innerCols;
                const innerCellHeight = innerHeight / innerRows;
                
                containedNodes.forEach((childId, childIndex) => {
                    const childRow = Math.floor(childIndex / innerCols);
                    const childCol = childIndex % innerCols;
                    
                    const childX = x - nodeWidth/2 + containerPadding + childCol * innerCellWidth + innerCellWidth / 2;
                    const childY = y - nodeHeight/2 + containerPadding + childRow * innerCellHeight + innerCellHeight / 2;
                    
                    this.nodePositions.set(childId, { x: childX, y: childY });
                    this.nodeElements.set(childId, { 
                        width: innerCellWidth * 0.8, 
                        height: innerCellHeight * 0.8 
                    });
                });
            }
        });
    }

    renderCanvas(): void {
        if (!this.ctx) return;

        // Clear canvas
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        this.ctx.fillStyle = '#DFE0EF';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

        if (!this.graphData) {
            // Draw placeholder node if no data
            const node = { x: 200, y: 300, radius: 22, color: '#45B7D1' };
            this.ctx.beginPath();
            this.ctx.arc(node.x + this.offsetX, node.y + this.offsetY, node.radius * this.scale, 0, Math.PI * 2);
            this.ctx.fillStyle = node.color;
            this.ctx.fill();
            this.ctx.strokeStyle = '#35354F';
            this.ctx.lineWidth = 3;
            this.ctx.stroke();
            return;
        }

        // Apply transformations for nodes and edges
        this.ctx.save();
        this.ctx.translate(this.offsetX, this.offsetY);
        this.ctx.scale(this.scale, this.scale);
        
        // Draw edges first (so they appear under nodes)
        this.drawEdges();
        
        // Draw nodes
        this.drawNodes();

        this.ctx.restore();

        // Draw legend with fixed position (no transform)
        if (this.legendVisible) {
            this.drawLegend();
        }
    }

    drawNodes(): void {
        if (!this.graphData) return;
        
        const nodes = this.graphData.nodes;
        
        // First, draw container nodes (those that contain other nodes)
        for (const [id, node] of Object.entries(nodes)) {
            if (!node.nodesWithin || node.nodesWithin.length === 0) continue;
            
            const pos = this.nodePositions.get(id);
            const size = this.nodeElements.get(id);
            
            if (!pos || !size) continue;
            
            // Draw container as a rectangle
            this.ctx.fillStyle = this.eventColors[node.event] || '#999999';
            this.ctx.strokeStyle = '#35354F';
            this.ctx.lineWidth = 2;
            
            // Draw fill first
            this.roundRect(
                pos.x - size.width / 2,
                pos.y - size.height / 2,
                size.width,
                size.height,
                10,
                true,
                false
            );
            
            // Then draw stroke (border) separately
            this.roundRect(
                pos.x - size.width / 2,
                pos.y - size.height / 2,
                size.width,
                size.height,
                10,
                false,
                true
            );
            
            // Draw container label
            this.ctx.fillStyle = '#333333';
            this.ctx.font = '14px Arial';
            this.ctx.textAlign = 'center';
            this.ctx.fillText(id, pos.x, pos.y - size.height / 2 + 20);
        }
        
        // Then, draw regular nodes
        for (const [id, node] of Object.entries(nodes)) {
            if (node.nodesWithin && node.nodesWithin.length > 0) continue; // Skip containers, already drawn
            
            const pos = this.nodePositions.get(id);
            const size = this.nodeElements.get(id);
            
            if (!pos || !size) continue;
            
            this.ctx.fillStyle = this.eventColors[node.event] || '#999999';
            this.ctx.strokeStyle = '#35354F';
            this.ctx.lineWidth = node.id === this.hoverNode?.id ? 4 : 2;
            
            // Get queue information
            const queueContents = this.nodeQueues.get(id) || [];
            const hasItems = queueContents.length > 0;
            
            // Check if node has parts currently being processed
            const partsAtNode = Array.from(this.partStates.entries())
                .filter(([_, state]) => state.nodeId === id)
                .map(([partId, _]) => partId);
            
            // Adjust visual based on queue and processing status
            if (hasItems || partsAtNode.length > 0) {
                // Use a glow effect for nodes with items in queue or processing
                this.ctx.shadowColor = node.event === 'Processing' ? 'rgba(52, 168, 83, 0.6)' : 'rgba(251, 188, 5, 0.6)';
                this.ctx.shadowBlur = 10;
                this.ctx.shadowOffsetX = 0;
                this.ctx.shadowOffsetY = 0;
                
                // Increase line width for nodes with items
                this.ctx.lineWidth += 1;
            }
            
            // Draw based on processing time (circles for 0, rectangles for others)
            if (node.processingTime === 0) {
                // Draw circle fill
                this.ctx.beginPath();
                const radius = Math.min(size.width, size.height) / 2;
                this.ctx.arc(pos.x, pos.y, radius, 0, Math.PI * 2);
                this.ctx.fill();
                
                // Draw circle stroke separately
                this.ctx.beginPath();
                this.ctx.arc(pos.x, pos.y, radius, 0, Math.PI * 2);
                this.ctx.stroke();
            } else {
                // Draw rectangular fill
                this.roundRect(
                    pos.x - size.width / 2,
                    pos.y - size.height / 2,
                    size.width,
                    size.height,
                    8,
                    true,
                    false
                );
                
                // Draw rectangular stroke separately
                this.roundRect(
                    pos.x - size.width / 2,
                    pos.y - size.height / 2,
                    size.width,
                    size.height,
                    8,
                    false,
                    true
                );
            }
            
            // Reset shadow
            this.ctx.shadowBlur = 0;
            this.ctx.shadowOffsetX = 0;
            this.ctx.shadowOffsetY = 0;
            
            // Draw node label
            this.ctx.fillStyle = '#333333';
            this.ctx.font = '12px Arial';
            this.ctx.textAlign = 'center';
            this.ctx.fillText(id, pos.x, pos.y - 5);
            
            // Show queue size if any
            if (hasItems) {
                this.ctx.fillStyle = '#FF5722';
                this.ctx.font = 'bold 10px Arial';
                this.ctx.fillText(`Queue: ${queueContents.length}`, pos.x, pos.y + 15);
            }
            
            // Show processing status if any parts are here
            if (partsAtNode.length > 0) {
                this.ctx.fillStyle = '#4285F4';
                this.ctx.font = 'bold 10px Arial';
                
                if (partsAtNode.length === 1) {
                    // Show the part ID if only one
                    this.ctx.fillText(`Part: ${partsAtNode[0]}`, pos.x, pos.y + (hasItems ? 30 : 15));
                } else {
                    // Just show count if multiple
                    this.ctx.fillText(`Parts: ${partsAtNode.length}`, pos.x, pos.y + (hasItems ? 30 : 15));
                }
            }
        }
    }

    drawEdges(): void {
        if (!this.graphData) return;
        
        const nodes = this.graphData.nodes;
        
        // Draw connections between nodes with reduced opacity
        this.ctx.globalAlpha = 0.5;
        
        for (const [sourceId, sourceNode] of Object.entries(nodes)) {
            if (!sourceNode.nextNodes || sourceNode.nextNodes.length === 0) continue;
            
            const sourcePos = this.nodePositions.get(sourceId);
            if (!sourcePos) continue;
            
            for (const targetId of sourceNode.nextNodes) {
                const targetPos = this.nodePositions.get(targetId);
                if (!targetPos) continue;
                
                // Check if there's any part transitioning on this edge
                const transitingParts = Array.from(this.partsInTransit.entries())
                    .filter(([_, transit]) => transit.from === sourceId && transit.to === targetId);
                
                const isActiveEdge = transitingParts.length > 0;
                
                // Calculate direction vector
                const dx = targetPos.x - sourcePos.x;
                const dy = targetPos.y - sourcePos.y;
                const length = Math.sqrt(dx * dx + dy * dy);
                
                if (length === 0) continue;
                
                // Normalize
                const nx = dx / length;
                const ny = dy / length;
                
                // Get node sizes
                const sourceSize = this.nodeElements.get(sourceId);
                const targetSize = this.nodeElements.get(targetId);
                
                if (!sourceSize || !targetSize) continue;
                
                // Determine start and end points (on the edge of nodes)
                let startX, startY, endX, endY;
                
                // Check if source is a circle (0 processing time)
                if (nodes[sourceId].processingTime === 0) {
                    const radius = Math.min(sourceSize.width, sourceSize.height) / 2;
                    startX = sourcePos.x + nx * radius;
                    startY = sourcePos.y + ny * radius;
                } else {
                    // For rectangles, find intersection with rectangle border
                    const halfWidth = sourceSize.width / 2;
                    const halfHeight = sourceSize.height / 2;
                    
                    // Simple approximation - not perfect but works for most cases
                    if (Math.abs(nx) * halfWidth > Math.abs(ny) * halfHeight) {
                        // Horizontal intersection
                        startX = sourcePos.x + (nx > 0 ? halfWidth : -halfWidth);
                        startY = sourcePos.y + ny * (halfWidth / Math.abs(nx));
                    } else {
                        // Vertical intersection
                        startX = sourcePos.x + nx * (halfHeight / Math.abs(ny));
                        startY = sourcePos.y + (ny > 0 ? halfHeight : -halfHeight);
                    }
                }
                
                // Check if target is a circle
                if (nodes[targetId].processingTime === 0) {
                    const radius = Math.min(targetSize.width, targetSize.height) / 2;
                    endX = targetPos.x - nx * radius;
                    endY = targetPos.y - ny * radius;
                } else {
                    // For rectangles, find intersection with rectangle border
                    const halfWidth = targetSize.width / 2;
                    const halfHeight = targetSize.height / 2;
                    
                    if (Math.abs(nx) * halfWidth > Math.abs(ny) * halfHeight) {
                        // Horizontal intersection
                        endX = targetPos.x - (nx > 0 ? halfWidth : -halfWidth);
                        endY = targetPos.y - ny * (halfWidth / Math.abs(nx));
                    } else {
                        // Vertical intersection
                        endX = targetPos.x - nx * (halfHeight / Math.abs(ny));
                        endY = targetPos.y - (ny > 0 ? halfHeight : -halfHeight);
                    }
                }
                
                // Draw edge with pulse effect if active
                this.ctx.beginPath();
                this.ctx.moveTo(startX, startY);
                this.ctx.lineTo(endX, endY);
                
                if (isActiveEdge) {
                    // Active edge with pulse effect
                    this.ctx.strokeStyle = '#FF5722';
                    this.ctx.lineWidth = 3;
                    
                    // Create gradient for pulse effect
                    const gradient = this.ctx.createLinearGradient(startX, startY, endX, endY);
                    
                    // Pulse effect based on most progressed part
                    const mostProgressed = transitingParts.reduce((max, current) => 
                        current[1].progress > max[1].progress ? current : max, transitingParts[0]);
                    
                    const progress = mostProgressed[1].progress;
                    
                    // Create pulsing gradient
                    gradient.addColorStop(Math.max(0, progress - 0.2), 'rgba(255, 87, 34, 0.1)');
                    gradient.addColorStop(progress, 'rgba(255, 87, 34, 0.8)');
                    gradient.addColorStop(Math.min(1, progress + 0.2), 'rgba(255, 87, 34, 0.1)');
                    
                    this.ctx.strokeStyle = gradient;
                } else {
                    // Normal edge
                    this.ctx.strokeStyle = '#666666';
                    this.ctx.lineWidth = 1.5;
                }
                
                this.ctx.stroke();
                
                // Draw arrow head at end point
                const arrowSize = 10;
                const angle = Math.atan2(targetPos.y - sourcePos.y, targetPos.x - sourcePos.x);
                
                this.ctx.beginPath();
                this.ctx.moveTo(endX, endY);
                this.ctx.lineTo(
                    endX - arrowSize * Math.cos(angle - Math.PI/6),
                    endY - arrowSize * Math.sin(angle - Math.PI/6)
                );
                this.ctx.lineTo(
                    endX - arrowSize * Math.cos(angle + Math.PI/6),
                    endY - arrowSize * Math.sin(angle + Math.PI/6)
                );
                this.ctx.closePath();
                this.ctx.fillStyle = isActiveEdge ? '#FF5722' : '#666666';
                this.ctx.fill();
            }
        }
        
        // Draw parts in transit
        for (const [partId, transit] of this.partsInTransit.entries()) {
            const sourcePos = this.nodePositions.get(transit.from);
            const targetPos = this.nodePositions.get(transit.to);
            
            if (sourcePos && targetPos) {
                // Calculate position along the edge based on progress
                const x = sourcePos.x + (targetPos.x - sourcePos.x) * transit.progress;
                const y = sourcePos.y + (targetPos.y - sourcePos.y) * transit.progress;
                
                // Draw part indicator with glow effect
                this.ctx.shadowColor = 'rgba(255, 87, 34, 0.6)';
                this.ctx.shadowBlur = 10;
                this.ctx.shadowOffsetX = 0;
                this.ctx.shadowOffsetY = 0;
                
                this.ctx.beginPath();
                this.ctx.arc(x, y, 8, 0, Math.PI * 2);
                this.ctx.fillStyle = '#FF5722';
                this.ctx.fill();
                this.ctx.strokeStyle = '#FFF';
                this.ctx.lineWidth = 2;
                this.ctx.stroke();
                
                // Reset shadow
                this.ctx.shadowBlur = 0;
                
                // Draw part ID
                this.ctx.fillStyle = '#000';
                this.ctx.font = 'bold 10px Arial';
                this.ctx.textAlign = 'center';
                this.ctx.fillText(partId, x, y - 12);
            }
        }

        // Reset opacity
        this.ctx.globalAlpha = 1.0;
    }

    drawArrow(from: {x: number, y: number}, to: {x: number, y: number}): void {
        const headSize = 10;
        const angle = Math.atan2(to.y - from.y, to.x - from.x);
        
        // Draw arrow head
        this.ctx.beginPath();
        this.ctx.moveTo(to.x, to.y);
        this.ctx.lineTo(
            to.x - headSize * Math.cos(angle - Math.PI / 6),
            to.y - headSize * Math.sin(angle - Math.PI / 6)
        );
        this.ctx.lineTo(
            to.x - headSize * Math.cos(angle + Math.PI / 6),
            to.y - headSize * Math.sin(angle + Math.PI / 6)
        );
        this.ctx.closePath();
        this.ctx.fillStyle = '#666666';
        this.ctx.fill();
    }

    drawLegend(): void {
        const padding = 15;
        const itemHeight = 25;
        const legendWidth = 150;
        const legendHeight = Object.keys(this.eventColors).length * itemHeight + padding * 2;
        
        // Position legend in bottom left with fixed position
        const x = padding;
        const y = this.canvas.height - legendHeight - padding;
        
        // Legend background
        this.ctx.fillStyle = 'rgba(255, 255, 255, 0.8)';
        this.ctx.strokeStyle = '#666666';
        this.ctx.lineWidth = 1;
        
        // Draw rounded rectangle (fixed position)
        this.ctx.beginPath();
        this.ctx.moveTo(x + 5, y);
        this.ctx.lineTo(x + legendWidth - 5, y);
        this.ctx.quadraticCurveTo(x + legendWidth, y, x + legendWidth, y + 5);
        this.ctx.lineTo(x + legendWidth, y + legendHeight - 5);
        this.ctx.quadraticCurveTo(x + legendWidth, y + legendHeight, x + legendWidth - 5, y + legendHeight);
        this.ctx.lineTo(x + 5, y + legendHeight);
        this.ctx.quadraticCurveTo(x, y + legendHeight, x, y + legendHeight - 5);
        this.ctx.lineTo(x, y + 5);
        this.ctx.quadraticCurveTo(x, y, x + 5, y);
        this.ctx.closePath();
        this.ctx.fill();
        this.ctx.stroke();
        
        // Legend title
        this.ctx.fillStyle = '#333333';
        this.ctx.font = 'bold 14px Arial';
        this.ctx.textAlign = 'left';
        this.ctx.fillText(
            'Node States',
            x + padding,
            y + 25
        );
        
        // Legend items
        let itemY = y + 50;
        
        for (const [state, color] of Object.entries(this.eventColors)) {
            // Color box
            this.ctx.fillStyle = color;
            this.ctx.fillRect(
                x + padding,
                itemY - 10,
                15,
                15
            );
            this.ctx.strokeStyle = '#333333';
            this.ctx.strokeRect(
                x + padding,
                itemY - 10,
                15,
                15
            );
            
            // State label
            this.ctx.fillStyle = '#333333';
            this.ctx.font = '12px Arial';
            this.ctx.textAlign = 'left';
            this.ctx.fillText(
                state,
                x + padding + 25,
                itemY
            );
            
            itemY += itemHeight;
        }
    }

    roundRect(x: number, y: number, width: number, height: number, radius: number, 
              doFill: boolean = true, doStroke: boolean = true): void {
        this.ctx.beginPath();
        this.ctx.moveTo(x + radius, y);
        this.ctx.lineTo(x + width - radius, y);
        this.ctx.quadraticCurveTo(x + width, y, x + width, y + radius);
        this.ctx.lineTo(x + width, y + height - radius);
        this.ctx.quadraticCurveTo(x + width, y + height, x + width - radius, y + height);
        this.ctx.lineTo(x + radius, y + height);
        this.ctx.quadraticCurveTo(x, y + height, x, y + height - radius);
        this.ctx.lineTo(x, y + radius);
        this.ctx.quadraticCurveTo(x, y, x + radius, y);
        this.ctx.closePath();
        
        if (doFill) {
            this.ctx.fill();
        }
        
        if (doStroke) {
            this.ctx.stroke();
        }
    }

    resetCanvas(): void {
        this.initializeElements();
        this.setCanvasSize(this.parentContainer);
        this.calculateNodePositions();
        this.renderCanvas();
    }

    getNodeAtPosition(x: number, y: number): Node | null {
        if (!this.graphData) return null;
        
        // Adjust for pan and zoom
        x = (x - this.offsetX) / this.scale;
        y = (y - this.offsetY) / this.scale;
        
        for (const [id, node] of Object.entries(this.graphData.nodes)) {
            const pos = this.nodePositions.get(id);
            const size = this.nodeElements.get(id);
            
            if (!pos || !size) continue;
            
            // Check if point is inside node
            if (node.processingTime === 0) {
                // Circle
                const radius = Math.min(size.width, size.height) / 2;
                const dx = x - pos.x;
                const dy = y - pos.y;
                if (dx * dx + dy * dy <= radius * radius) {
                    return node;
                }
            } else {
                // Rectangle
                if (x >= pos.x - size.width / 2 &&
                    x <= pos.x + size.width / 2 &&
                    y >= pos.y - size.height / 2 &&
                    y <= pos.y + size.height / 2) {
                    return node;
                }
            }
        }
        
        return null;
    }

    addEventListeners(): void {
        this.canvas.addEventListener('mousedown', (e) => this.mouseDown(e));
        this.canvas.addEventListener('mouseup', () => this.mouseUp());
        this.canvas.addEventListener('mousemove', (e) => this.mouseMove(e));
        this.canvas.addEventListener('wheel', (e) => this.mouseWheel(e));
        this.homeButton.addEventListener('click', () => this.resetCanvas());
        window.onresize = () => this.resetCanvas();
    }

    mouseDown(e: MouseEvent): void {
        const mouseX = e.clientX - this.canvas.getBoundingClientRect().left;
        const mouseY = e.clientY - this.canvas.getBoundingClientRect().top;
        
        // Check if clicking on a node
        const node = this.getNodeAtPosition(mouseX, mouseY);
        if (node) {
            console.log('Clicked on node:', node);

            
            // Could add node selection behavior here
        } else {
            // Start panning
            this.isDragging = true;
            this.lastX = mouseX;
            this.lastY = mouseY;
            this.canvas.style.cursor = 'grabbing';
        }
    }

    mouseMove(e: MouseEvent): void {
        const mouseX = e.clientX - this.canvas.getBoundingClientRect().left;
        const mouseY = e.clientY - this.canvas.getBoundingClientRect().top;
        
        if (this.isDragging) {
            // Handle panning
            const dx = mouseX - this.lastX;
            const dy = mouseY - this.lastY;
            
            this.lastX = mouseX;
            this.lastY = mouseY;

            this.offsetX += dx;
            this.offsetY += dy;

            this.renderCanvas();
        } else {
            // Check for hover
            const node = this.getNodeAtPosition(mouseX, mouseY);
            
            if (node) {
                this.canvas.style.cursor = 'pointer';
                
                if (this.hoverNode !== node) {
                    this.hoverNode = node;
                    this.renderCanvas();
                }
                
                // Show tooltip
                this.tooltip.style.display = 'block';
                this.tooltip.style.left = `${e.clientX + 10}px`;
                this.tooltip.style.top = `${e.clientY + 10}px`;
                
                // Format tooltip content
                let tooltipContent = `
                    <strong>${node.id}</strong><br>
                    State: ${node.event}<br>
                    Processing time: ${node.processingTime}s<br>
                    ${node.nextNodes && node.nextNodes.length ? `Connections: ${node.nextNodes.join(', ')}` : ''}
                `;
                
                // Add queue contents if any
                const queueContents = this.nodeQueues.get(node.id);
                if (queueContents && queueContents.length > 0) {
                    tooltipContent += `<br><br>Queue: ${queueContents.join(', ')}`;
                }
                
                // List parts currently at this node
                const partsAtNode = Array.from(this.partStates.entries())
                    .filter(([_, state]) => state.nodeId === node.id)
                    .map(([partId, _]) => partId);
                    
                if (partsAtNode.length > 0) {
                    tooltipContent += `<br><br>Parts: ${partsAtNode.join(', ')}`;
                }
                
                this.tooltip.innerHTML = tooltipContent;
            } else {
                this.canvas.style.cursor = 'default';
                this.tooltip.style.display = 'none';
                
                if (this.hoverNode) {
                    this.hoverNode = null;
                    this.renderCanvas();
                }
            }
        }
    }

    mouseUp(): void {
        this.isDragging = false;
        this.canvas.style.cursor = 'default';
    }

    mouseWheel(e: WheelEvent): void {
        e.preventDefault();
        const zoomFactor = 0.1;
        const mouseX = e.offsetX;
        const mouseY = e.offsetY;
        
        if (e.deltaY < 0) {
            // Zoom in
            this.scale = Math.min(3, this.scale + zoomFactor);
        } else {
            // Zoom out
            this.scale = Math.max(0.5, this.scale - zoomFactor);
        }

        // Zoom centered on mouse position
        this.offsetX = mouseX - (mouseX - this.offsetX) * (this.scale / (this.scale + (e.deltaY < 0 ? zoomFactor : -zoomFactor)));
        this.offsetY = mouseY - (mouseY - this.offsetY) * (this.scale / (this.scale + (e.deltaY < 0 ? zoomFactor : -zoomFactor)));

        this.renderCanvas();
    }

    processLogMessage(logData: any): void {
        switch(logData.type) {
            case 'state':
                this.partStates.set(logData.partId, {
                    nodeId: logData.nodeId,
                    state: logData.state
                });
                break;
                
            case 'transition':
                // Start a new transition animation
                this.partsInTransit.set(logData.partId, {
                    from: logData.sourceNode,
                    to: logData.targetNode,
                    progress: 0 // 0 to 1 for animation
                });
                break;
                
            case 'queue':
                this.nodeQueues.set(logData.nodeId, logData.contents);
                break;
        }
        
        // Redraw to reflect changes
        this.renderCanvas();
    }
    
    startAnimationLoop(): void {
        const animate = () => {
            let needsRedraw = false;
            
            // Update transition animations
            for (const [partId, transit] of this.partsInTransit.entries()) {
                // Speed up/down based on number of active transitions
                const transitionSpeed = 0.01 + (0.01 / Math.max(1, this.partsInTransit.size));
                
                transit.progress += transitionSpeed;
                
                if (transit.progress >= 1) {
                    this.partsInTransit.delete(partId);
                    
                    // When transition completes, update part state
                    this.partStates.set(partId, {
                        nodeId: transit.to,
                        state: 'Arrived'
                    });
                }
                
                needsRedraw = true;
            }
            
            // Add subtle animation for processing nodes
            if (this.graphData) {
                for (const [id, node] of Object.entries(this.graphData.nodes)) {
                    if (node.event === 'Processing') {
                        needsRedraw = true;
                    }
                }
            }
            
            if (needsRedraw) {
                this.renderCanvas();
            }
            
            requestAnimationFrame(animate);
        };
        
        requestAnimationFrame(animate);
    }
}