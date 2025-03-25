import {Connection} from "./handleSocket.js";
import {Graph,Node} from "./graph.js";

declare global {
  interface Window {
    appData: {
      allNodes: string;
      count: string;
    };
  }
}
console.log("testData.ts loaded");

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

const saveEdit = document.getElementById("edit-save") as HTMLButtonElement

saveEdit.onclick = SetNode

function SetNode() {
    const nodeTextArea = document.getElementById("edit-node") as HTMLTextAreaElement;
    const nodeData = JSON.parse(nodeTextArea.value);
    const statusData = document.getElementById("edit-status") as HTMLSpanElement;
    fetch(`/api/simdata/set_node`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(nodeData)
    })
    .then(response =>{ 
        response.json()
        statusData.innerHTML = "Node data updated"
    })
    .catch(error => {
        console.error('Error setting node data:', error);
        statusData.innerHTML = "Error updating node data";
    });

}


graph.renderCanvas();

const conn = new Connection();
conn.connectToTopic("logs", (msg: string) => {
    try {
        const logData = typeof msg === 'string' ? JSON.parse(msg) : msg;
        
        graph.processLogMessage(logData);
        
        updateTextLog(msg);
    } catch (e) {
        updateTextLog(msg);
        console.error("Failed to parse log message:", e);
    }
});

function updateTextLog(msg: any) {
    const logTextArea = document.getElementById("logs") as HTMLTextAreaElement;
    if (!logTextArea) return;
    
    // Make sure msg is a string
    let textMsg = typeof msg === 'string' ? msg : JSON.stringify(msg, null, 2);
    
    var messages = textMsg.split('\n');
    for (var i = 0; i < messages.length; i++) {
        if (messages[i].trim() !== "") {
            logTextArea.value += messages[i] + "\n";
            logTextArea.scrollTop = logTextArea.scrollHeight;
        }
    }
    
    const messageLength = logTextArea.value.length;
    const cutOff = 2000;
    if (messageLength > cutOff) {
        logTextArea.value = logTextArea.value.substring(messageLength - cutOff);
    }
}

try {
    if (window.appData && window.appData.allNodes) {
        const nodes = JSON.parse(window.appData.allNodes);
        graph.setData({ nodes });
    }
} catch (error) {
    console.error('Error loading static node data:', error);
}
