import {Connection} from "./handleSocket.js";
import {Graph} from "./graph.js";

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
graph.renderCanvas();

const conn = new Connection();
conn.connectToTopic("logs", (msg: string) => {
    // Try to parse message as JSON
    try {
        const logData = typeof msg === 'string' ? JSON.parse(msg) : msg;
        
        // Update the graph with the log event
        graph.processLogMessage(logData);
        
        // Also update the text log
        updateTextLog(msg);
    } catch (e) {
        // If not valid JSON, just update the text log
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