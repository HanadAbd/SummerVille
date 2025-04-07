import {Connection} from "./handleSocket.js";
import {Graph,Node} from "./graph.js";
import { Grid } from "gridjs";


declare global {
  interface Window {
    appData: {
      allNodes: string;
      count: string;
    };
  }
}


const canvas = document.getElementById("map") as HTMLCanvasElement;
const homeButton = document.getElementById("homeButton") as HTMLButtonElement;
const tooltip = document.getElementById("tooltip") as HTMLElement;

const graph = new Graph(canvas, homeButton, tooltip);

graph.nodeClickCallback = (node: Node) => {
    const nodeTextArea = document.getElementById("edit-node") as HTMLTextAreaElement;

    fetch(`/api/simdata/get_node?node_id=${node.id}`)
        .then(response => response.json())
        .then(data => {
            nodeTextArea.value = JSON.stringify(data["payload"], null, 2);
        })
        .catch(error => {
            console.error('Error fetching node data:', error);
        });
        
    console.log("Node clicked:", node);
}


graph.renderCanvas();

const conn = new Connection();

conn.connectToTopic("logs", (msg: string) => {
    try {
        const logData = typeof msg === 'string' ? JSON.parse(msg) : msg;
        
        graph.processLogMessage(logData);
        
    } catch (e) {
        console.error("Failed to parse log message:", e);
    }
});
try {
    if (window.appData && window.appData.allNodes) {
        const nodes = JSON.parse(window.appData.allNodes);
        graph.setData({ nodes });
    }
} catch (error) {
    console.error('Error loading static node data:', error);
}