type node = {
    id: string;
    name: string;
    type: string;
    state: string;
    connects: string[];
    contains: string[];
    parameters?: Record<string, any>;
};

type pos ={
    x:number;
    y:number;
    z:number;  
}

document.addEventListener('DOMContentLoaded', function () {
    intialiseFactoryFloor();
    console.log('Document loaded');
});

window.addEventListener("resize", function () {
    clearCanvas();
    intialiseFactoryFloor();
});
function clearCanvas(){
    const canvas = document.getElementById('factoryFloor') as HTMLCanvasElement;
    const ctx = canvas.getContext('2d') as CanvasRenderingContext2D;
    ctx.clearRect(0, 0, canvas.width, canvas.height);
}


function intialiseFactoryFloor(){
    const factoryFloor = document.getElementById('factoryFloor') as HTMLCanvasElement;
    const ctx = factoryFloor.getContext('2d') as CanvasRenderingContext2D;
    const container = factoryFloor.parentElement as HTMLElement;
    const allNodes = getAllNodes()


    setCanvasSize(factoryFloor, container,ctx);
    populateCanvas(ctx,allNodes);
}

function setCanvasSize (canvas:HTMLCanvasElement, container:HTMLElement,ctx:CanvasRenderingContext2D) :void {
    const width = container.clientWidth;
    const height = container.clientHeight;
    canvas.width = width;
    canvas.height = height;
    ctx.fillStyle = '#DFE0EF';

    ctx.fillRect(0, 0, width, height);
}
function getAllNodes(): node[] {
    return [
        {
            "id": "start",
            "name": "Start",
            "type": "start",
            "state": "idle",
            "connects": ["station1"],
            "contains": []
        },
        {
            "id": "complete",
            "name": "Complete",
            "type": "end",
            "state": "idle",
            "connects": [],
            "contains": [],
            "parameters": {
                "amount": 23
            }
        },
        {
            "id":"reject",
            "name":"Reject",
            "type":"end",
            "state":"idle",
            "connects":[],
            "contains":[],
            "parameters":{
                "amount": 2
            }
        },
        {
            "id": "reject",
            "name": "Reject",
            "type": "end",
            "state": "idle",
            "connects": [],
            "contains": []
        },
        {
            "id": "station1",
            "name": "Cutting Station",
            "type": "station",
            "state": "idle",
            "connects": ["station2"],
            "contains": ["cutting1", "cutting2", "cutting3"],
            "parameters": {
                "failureRate": 0.1,
                "failureType": "cutting"
            }
        },
        {
            "id": "station2",
            "name": "Sensor Station",
            "type": "station",
            "state": "idle",
            "connects": ["complete","reject"],
            "contains": ["sensor1", "sensor2", "sensor3"]
        },
        {
            "id": "cutting1",
            "name": "Cutting 1",
            "type": "cuttingMachine",
            "state": "idle",
            "connects": ["station2"],
            "contains": []
        },
        {
            "id": "cutting2",
            "name": "Cutting 2",
            "type": "cuttingMachine",
            "state": "idle",
            "connects": ["station2"],
            "contains": []
        },
        {
            "id": "cutting3",
            "name": "Cutting 3",
            "type": "cuttingMachine",
            "state": "idle",
            "connects": ["station2"],
            "contains": []
        },
        {
            "id": "sensor1",
            "name": "Sensor 1",
            "type": "sensorMachine",
            "state": "idle",
            "connects": ["complete", "reject"],
            "contains": []
        },
        {
            "id": "sensor2",
            "name": "Sensor 2",
            "type": "sensorMachine",
            "state": "idle",
            "connects": ["complete", "reject"],
            "contains": []
        },
        {
            "id": "sensor3",
            "name": "Sensor 3",
            "type": "sensorMachine",
            "state": "idle",
            "connects": ["complete", "reject"],
            "contains": []
        }
    
    ]
}

function populateCanvas(ctx : CanvasRenderingContext2D, nodes: node[]):void {
    const nodePositions = calculateNodePositions(nodes, ctx);
    
    nodes.forEach(node => {
        if (node.connects) {
            node.connects.forEach(targetId => {
                const targetPos = nodePositions[targetId];
                const sourcePos = nodePositions[node.id];
                if (targetPos && sourcePos) {
                    const targetNode = nodes.find(n => n.id === targetId);
                    const isBidirectional = targetNode && targetNode.connects && targetNode.connects.includes(node.id);
                    drawArrow(ctx, sourcePos.x, sourcePos.y, targetPos.x, targetPos.y, isBidirectional);
                }
            });
        }
    });

    const sortedNodes = [...nodes].sort((a, b) => {
        if ((a.contains?.length || 0) > 0 && (b.contains?.length || 0) === 0) return -1;
        if ((a.contains?.length || 0) === 0 && (b.contains?.length || 0) > 0) return 1;
        return 0;
    });

    sortedNodes.forEach(node => {
        const pos = nodePositions[node.id];
        if (pos) {
            drawNode(ctx, node, pos.x, pos.y);
        }
    });

    setupNodeInteractivity(ctx, nodes, nodePositions);
}

function calculateNodePositions(nodes:node[], ctx:CanvasRenderingContext2D) : Record<string, pos>{
    const positions: Record<string,pos>= {};
    const levels :Record<number,node[]>= {};
    
    let startNode = nodes.find(n => n.type === 'start');
    if (!startNode) return positions

    assignLevels(startNode, 0);
    
    function assignLevels(node:node, level:number) {
        if (!node) return;
        if (!levels[level]) levels[level] = [];
        levels[level].push(node);
        
        if (node.connects) {
            node.connects.forEach(targetId => {
                const targetNode = nodes.find(n => n.id === targetId);
                if (targetNode && !levels[level + 1]?.includes(targetNode)) {
                    assignLevels(targetNode, level + 1);
                }
            });
        }
    }
    
    const canvas = ctx.canvas;
    const levelWidth = canvas.width / (Object.keys(levels).length + 1);
    
    Object.entries(levels).forEach(([level, nodesInLevel]) => {
        const levelY = canvas.height / (nodesInLevel.length + 1);
        nodesInLevel.forEach((node, index) => {
            positions[node.id] = {
                x: levelWidth * (parseInt(level) + 1),
                y: levelY * (index + 1),
                z:1
            };
        });
    });
    
    return positions;
}

function drawNode(ctx: CanvasRenderingContext2D, node: node, x: number, y: number) {
    // Base radius for nodes
    let radius = 30;
    
    // Make container nodes larger
    if (node.contains && node.contains.length > 0) {
        radius = 45; // Larger radius for container nodes
    }
    
    ctx.fillStyle = getNodeColor(node.type);
    ctx.strokeStyle = '#000';
    ctx.lineWidth = 2;
    
    ctx.beginPath();
    ctx.arc(x, y, radius, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();
    
    if (node.contains && node.contains.length > 0) {
        ctx.beginPath();
        ctx.arc(x, y, radius + 5, 0, Math.PI * 2);
        ctx.strokeStyle = '#666';
        ctx.stroke();
        
        // Add indicator that this is a container
        ctx.fillStyle = '#000';
        ctx.font = '10px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(`Contains: ${node.contains.length}`, x, y + radius - 10);
    }
    
    ctx.fillStyle = '#000';
    ctx.font = '12px Arial';
    ctx.textAlign = 'center';
    ctx.fillText(node.name, x, y + 4);
}

function drawArrow(ctx: CanvasRenderingContext2D, fromX:number, fromY:number, toX:number, toY:number, isBidirectional = false) {
    const headLength = 10;
    const dx = toX - fromX;
    const dy = toY - fromY;
    const angle = Math.atan2(dy, dx);
    
    const distance = Math.sqrt(dx * dx + dy * dy);
    const radius = 30;
    
    const startX = fromX + (radius * Math.cos(angle));
    const startY = fromY + (radius * Math.sin(angle));
    const endX = toX - (radius * Math.cos(angle));
    const endY = toY - (radius * Math.sin(angle));
    
    ctx.beginPath();
    
    if (isBidirectional) {
        const midX = (fromX + toX) / 2;
        const midY = (fromY + toY) / 2;
        const offsetX = -dy * 0.2; 
        const offsetY = dx * 0.2;
        
        ctx.moveTo(startX, startY);
        ctx.quadraticCurveTo(midX + offsetX, midY + offsetY, endX, endY);
    } else {
        ctx.moveTo(startX, startY);
        ctx.lineTo(endX, endY);
    }
    
    ctx.lineTo(endX - headLength * Math.cos(angle - Math.PI/6), endY - headLength * Math.sin(angle - Math.PI/6));
    ctx.moveTo(endX, endY);
    ctx.lineTo(endX - headLength * Math.cos(angle + Math.PI/6), endY - headLength * Math.sin(angle + Math.PI/6));
    
    ctx.strokeStyle = '#666';
    ctx.lineWidth = 2;
    ctx.stroke();
}

function getNodeColor(type:string) : string {
    const colors = {
        'start': '#90EE90',
        'end': '#FFB6C1',
        'station': '#ADD8E6',
        'cuttingMachine': '#DDA0DD',
        'sensorMachine': '#F0E68C'
    } as Record<string, string>;
    return colors[type] || '#FFFFFF';
}

function setupNodeInteractivity(ctx: CanvasRenderingContext2D, nodes: node[], nodePositions: Record<string, pos>) :void{
    const canvas = ctx.canvas;
    const tooltip = document.getElementById('tooltip') as HTMLElement;
    const editTextArea = document.getElementById('edit-node') as HTMLTextAreaElement;
    
    function isOverNode(x: number, y:number, nodeX: number, nodeY:number) {
        const radius = 30;
        return Math.sqrt((x - nodeX) * (x - nodeX) + (y - nodeY) * (y - nodeY)) <= radius;
    }
    
    canvas.addEventListener('click', function(e) {
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        for (const node of nodes) {
            const pos = nodePositions[node.id];
            if (pos && isOverNode(x, y, pos.x, pos.y)) {
                let parameterText = '';
                if (node.parameters) {
                    parameterText = JSON.stringify(node.parameters, null, 2);
                } else {
                    parameterText = `Node ID: ${node.id}\nName: ${node.name}\nType: ${node.type}\nState: ${node.state}`;
                }
                editTextArea.value = parameterText;
                return;
            }
        }
    });
    
    canvas.addEventListener('mousemove', function(e) {
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        let hoverNode = null;
        
        for (const node of nodes) {
            const pos = nodePositions[node.id];
            if (pos && isOverNode(x, y, pos.x, pos.y)) {
                hoverNode = node;
                break;
            }
        }
        
        if (hoverNode) {
            let tooltipContent = `<b>${hoverNode.name}</b><br>Type: ${hoverNode.type}<br>State: ${hoverNode.state}`;
            
            if (hoverNode.parameters) {
                tooltipContent += '<br><b>Parameters:</b><br>';
                for (const [key, value] of Object.entries(hoverNode.parameters)) {
                    tooltipContent += `${key}: ${value}<br>`;
                }
            }
            
            tooltip.innerHTML = tooltipContent;
            tooltip.style.display = 'block';
            tooltip.style.left = (e.clientX + 10) + 'px';
            tooltip.style.top = (e.clientY + 10) + 'px';
        } else {
            tooltip.style.display = 'none';
        }
    });
    
    canvas.addEventListener('mouseleave', function() {
        tooltip.style.display = 'none';
    });
}
