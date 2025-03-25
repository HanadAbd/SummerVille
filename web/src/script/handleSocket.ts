export class Connection {
    private socket: WebSocket | null;
    private topic: string | null;
    private reconnectAttempts: number = 0;
    private maxReconnectAttempts: number = 10;
    private reconnectInterval: number = 3000; 
    private messageCallback: ((msg: string) => void) | null = null;

    constructor() {
        this.socket = null;
        this.topic = null;
    }
    
    connectToTopic(topic: string, callback: (msg: string) => void) {
        this.topic = topic;
        this.messageCallback = callback;

        this.socket = this.establishConnection();
        if (this.socket) {
            console.log('Connected to topic:', topic);
        }
    }

    private establishConnection(): WebSocket {
        this.socket = new WebSocket(`ws://${window.location.host}/ws?topic=${this.topic}`);
        
        this.socket.onopen = () => {
            if (this.topic) {
                this.socket?.send(`subscribe:${this.topic}`);
                console.log(`Subscribed to ${this.topic}`);
            }
        }
        
        this.socket.onmessage = (event) => {
            if (this.messageCallback) {
                const logData = this.parseLogMessage(event.data);
                this.messageCallback(logData);
            }

            // Log raw message to console
        };

        this.socket.onclose = () => {
            this.handleDisconnection();
        };

        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.handleDisconnection();
        };

        return this.socket;
    }
    
    private parseLogMessage(message: string): any {
        // Split the message by newlines and process each line separately
        const lines = message.split('\n').filter(line => line.trim() !== '');
        
        // If there are multiple lines, process each line and return the first valid result
        if (lines.length > 1) {
            for (const line of lines) {
                const result = this.parseLogLine(line);
                if (result.type !== 'raw') {
                    return result;
                }
            }
        }
        
        // If no valid result was found or there's only one line, parse it normally
        return this.parseLogLine(lines[0] || message);
    }

    private parseLogLine(line: string): any {
        const trimmedLine = line.trim();
        
        if (trimmedLine.startsWith('node=')) {
            // Handle queue message format: "node=nodeId;queue=count"
            const parts = trimmedLine.split(';');
            if (parts.length >= 2 && parts[1].startsWith('queue=')) {
                const nodeId = parts[0].substring(5); // Remove 'node=' prefix
                const queueCount = parseInt(parts[1].substring(6)); // Remove 'queue=' prefix
                
                return {
                    type: 'queue',
                    nodeId: nodeId,
                    queueSize: queueCount,
                    contents: new Array(queueCount).fill('item')
                };
            }
        }
        
        else if (trimmedLine.startsWith('part=')) {
            // Handle state message format: "part=partId;state=stateValue;node=nodeId"
            const parts = trimmedLine.split(';');
            if (parts.length >= 3) {
                const partIdRaw = parts[0].substring(5); // Remove 'part=' prefix
                const partId = partIdRaw || 'node_state'; // Use a placeholder for empty part IDs
                const state = parts[1].substring(6); // Remove 'state=' prefix
                const nodeId = parts[2].substring(5); // Remove 'node=' prefix
                
                // If the part ID is empty, this is a node state update, not a part update
                if (!partIdRaw) {
                    return {
                        type: 'nodeState',
                        state: state,
                        nodeId: nodeId
                    };
                } else {
                    return {
                        type: 'partState',
                        partId: partId,
                        state: state,
                        nodeId: nodeId
                    };
                }
            }
        }
        
        // Handle transition message format: "partId;transition;sourceNode;targetNode"
        else {
            const parts = trimmedLine.split(';');
            if (parts.length >= 4 && parts[1] === 'transition') {
                return {
                    type: 'transition',
                    partId: parts[0],
                    sourceNode: parts[2],
                    targetNode: parts[3]
                };
            }
        }
        
        // Return raw message for unrecognized formats
        return { type: 'raw', message: trimmedLine };
    }

    private handleDisconnection() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            console.log(`Connection lost. Attempting to reconnect... (${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})`);
            setTimeout(() => {
                this.reconnectAttempts++;
                this.establishConnection();
            }, this.reconnectInterval);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }

    send(message: string): void {
        if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
            console.error('Socket is not connected');
            this.handleDisconnection();
            return;
        }
        this.socket.send(message);
    }

    async sendAPIRequest(endpoint: string, method: string = 'GET', body?: any): Promise<any> {
        try {
            const response = await fetch(endpoint, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: body ? JSON.stringify(body) : undefined,
            });
            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    close(): void {
        console.log('Closing connection to topic:', this.topic);
        this.topic = null;
        this.messageCallback = null;
        this.reconnectAttempts = 0;
        
        if (this.socket) {
            this.socket.close();
            this.socket = null;
        }
    }

    isConnected(): boolean {
        return this.socket !== null && this.socket.readyState === WebSocket.OPEN;
    }
}
