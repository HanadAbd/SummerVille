type CanvasMap ={
    element:HTMLCanvasElement;
    ctx:CanvasRenderingContext2D;
    offsetX:number;
    offsetY:number;
    scale:number;
    isDragging:boolean;
    lastX:number;
    lastY:number;
}

type Nodes = {
    id: string;
    name: string;
    type: string;
    state: string;
    connects: string[];
    contains: string[];
    parameters?: Record<string, any>;
    position?: Pos;
};

type Pos ={
    x:number;
    y:number;
    z:number;  
}

document.addEventListener('DOMContentLoaded', function () {
    intialiseFactoryFloor();
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
  
    const canvasMap:CanvasMap = {
        element:factoryFloor,
        ctx:ctx,
        offsetX:0,
        offsetY:0,
        scale:1,
        isDragging:false,
        lastX:0,
        lastY:0
    }


    setCanvasSize(canvasMap, container);
    renderCanvas(canvasMap);
    intialiseEvents(canvasMap);
}

function setCanvasSize (canvasMap:CanvasMap, container:HTMLElement) :void {
    const width = container.clientWidth;
    const height = container.clientHeight;
    canvasMap.element.width = width;
    canvasMap.element.height = height;
    canvasMap.ctx.fillStyle = '#DFE0EF';

    canvasMap.ctx.fillRect(0, 0, width, height);
}
function getAllNodes(): Nodes[] {
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

function renderCanvas(canvasMap: CanvasMap):void {
    const allNodes = getAllNodes();
    const nodePositions = calculatePos(allNodes);
}

function calculatePos(nodes:Nodes[],):Record<string, Pos>{
    const nodePositions:Record<string, Pos> = {};
    const start = nodes[0];
    return nodePositions;
    
}

function intialiseEvents(canvasMaps:CanvasMap):void{
    canvasMaps.element.addEventListener('mousedown', (e:MouseEvent) => HandleMouseDown(e, canvasMaps));
    canvasMaps.element.addEventListener('mousemove', (e:MouseEvent) => HandleMouseMove(e, canvasMaps));
    canvasMaps.element.addEventListener('mouseup', (e:MouseEvent) => HandleMouseUp(e, canvasMaps));
    canvasMaps.element.addEventListener('mouseleave', (e:MouseEvent) => HandleMouseUp(e, canvasMaps));
    canvasMaps.element.addEventListener('wheel', (e:WheelEvent) => HandleMouseWheel(e, canvasMaps));
}

function HandleMouseDown(e:MouseEvent, canvasMap:CanvasMap):void{
    const mouseX = e.clientX - canvasMap.element.offsetLeft;
    const mouseY = e.clientY - canvasMap.element.offsetTop;
    canvasMap.isDragging = true;
    canvasMap.lastX = mouseX;
    canvasMap.lastY = mouseY;
    canvasMap.element.style.cursor = 'grabbing';
}

function HandleMouseMove(e:MouseEvent, canvasMap:CanvasMap):void{
    if(!canvasMap.isDragging){
        return;
    }
    const mouseX = e.clientX - canvasMap.element.offsetLeft;
    const mouseY = e.clientY - canvasMap.element.offsetTop;
    const dx = mouseX - canvasMap.lastX;
    const dy = mouseY - canvasMap.lastY;
    canvasMap.lastX = mouseX;
    canvasMap.lastY = mouseY;
    canvasMap.offsetX += dx;
    canvasMap.offsetY += dy;

    renderCanvas(canvasMap);
}

function HandleMouseUp(e:MouseEvent, canvasMap:CanvasMap):void{
    canvasMap.isDragging = false;
}

function HandleMouseWheel(e:WheelEvent, canvasMap:CanvasMap):void{
    e.preventDefault();
    const zoomFactor = 0.1;
    const mouseX = e.offsetX;
    const mouseY = e.offsetY;
    
    if (e.deltaY < 0) {
        canvasMap.scale += zoomFactor;
    } else {
        canvasMap.scale -= zoomFactor;
        if (canvasMap.scale < 0.1) canvasMap.scale = 0.1; 
    }

    canvasMap.offsetX = mouseX - (mouseX - canvasMap.offsetX) * (canvasMap.scale / (canvasMap.scale + zoomFactor));
    canvasMap.offsetY = mouseY - (mouseY - canvasMap.offsetY) * (canvasMap.scale / (canvasMap.scale + zoomFactor));

    renderCanvas(canvasMap);
}

