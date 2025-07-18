import { EventEmitter } from 'events';
import { WebSocketMessage } from '../types';
export interface WebSocketClientOptions {
    guildId?: string;
    userId?: string;
    autoReconnect?: boolean;
    pingInterval?: number;
    pongTimeout?: number;
}
export declare class WebSocketClient extends EventEmitter {
    private ws;
    private url;
    private options;
    private reconnectAttempts;
    private pingTimer;
    private pongTimer;
    private isConnecting;
    private isReconnecting;
    constructor(options?: WebSocketClientOptions);
    connect(): Promise<void>;
    disconnect(): void;
    send(message: WebSocketMessage): void;
    subscribe(guildId: string): void;
    unsubscribe(): void;
    ping(): void;
    isConnected(): boolean;
    getReadyState(): number;
    private handleMessage;
    private handleClose;
    private scheduleReconnect;
    private startPingTimer;
    private stopPingTimer;
    private startPongTimer;
    private stopPongTimer;
    private handlePong;
}
//# sourceMappingURL=client.d.ts.map