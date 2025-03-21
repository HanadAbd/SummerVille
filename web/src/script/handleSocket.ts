export class Connection {
    private socket: WebSocket | null;
    private topic: string | null;
    private reconnectAttempts: number = 0;
    private maxReconnectAttempts: number = 5;
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
                // Parse structured log format
                const logData = this.parseLogMessage(event.data);
                this.messageCallback(JSON.stringify(logData));
            }
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
        const trimmedMessage = message.trim();
        const parts = trimmedMessage.split(';');
        
        if (parts.length < 3) {
            return { type: 'raw', message: message };
        }
        
        // Handle different message types
        switch(parts[1]) {
            case 'state':
                return {
                    type: 'state',
                    partId: parts[0],
                    state: parts[2],
                    nodeId: parts[3]
                };
            case 'transition':
                return {
                    type: 'transition',
                    partId: parts[0],
                    sourceNode: parts[2],
                    targetNode: parts[3]
                };
            case 'queue':
                return {
                    type: 'queue',
                    nodeId: parts[0],
                    contents: parts[2].split(',')
                };
            default:
                return { type: 'raw', message: message };
        }
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
