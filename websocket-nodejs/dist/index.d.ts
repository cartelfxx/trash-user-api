import { EventEmitter } from 'events';
import { ApiClient } from './api/client';
import { WebSocketClient } from './websocket/client';
export interface DiscordApiClientOptions {
    guildId?: string;
    userId?: string;
    autoReconnect?: boolean;
    pingInterval?: number;
    pongTimeout?: number;
}
export declare class DiscordApiClient extends EventEmitter {
    private apiClient;
    private wsClient;
    private options;
    constructor(options?: DiscordApiClientOptions);
    private setupWebSocketEvents;
    private handleWebSocketEvent;
    connect(): Promise<void>;
    disconnect(): void;
    isConnected(): boolean;
    subscribe(guildId: string): void;
    unsubscribe(): void;
    ping(): void;
    getApiClient(): ApiClient;
    getWebSocketClient(): WebSocketClient;
}
export { ApiClient } from './api/client';
export { WebSocketClient } from './websocket/client';
export { config } from './config';
export { logger } from './utils/logger';
export * from './types';
export default DiscordApiClient;
//# sourceMappingURL=index.d.ts.map