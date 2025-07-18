import { EventEmitter } from 'events';
import { ApiClient } from './api/client';
import { WebSocketClient } from './websocket/client';
import { config } from './config';
import { logger } from './utils/logger';
export class DiscordApiClient extends EventEmitter {
    constructor(options = {}) {
        super();
        this.options = {
            guildId: options.guildId || '',
            userId: options.userId || '',
            autoReconnect: options.autoReconnect ?? true,
            pingInterval: options.pingInterval || config.get().websocket.pingInterval,
            pongTimeout: options.pongTimeout || config.get().websocket.pongTimeout,
        };
        this.apiClient = new ApiClient();
        const wsOptions = {
            guildId: this.options.guildId,
            userId: this.options.userId,
            autoReconnect: this.options.autoReconnect,
            pingInterval: this.options.pingInterval,
            pongTimeout: this.options.pongTimeout,
        };
        this.wsClient = new WebSocketClient(wsOptions);
        this.setupWebSocketEvents();
    }
    setupWebSocketEvents() {
        this.wsClient.on('connected', () => {
            logger.info('Discord API Client connected', '🔗');
        });
        this.wsClient.on('disconnected', () => {
            logger.info('Discord API Client disconnected', '🔌');
        });
        this.wsClient.on('error', (error) => {
            logger.error(`WebSocket error: ${error.message}`, '❌');
        });
        this.wsClient.on('message', (event) => {
            this.handleWebSocketEvent(event);
        });
        this.wsClient.on('cache_update', (data) => {
            logger.info('Cache updated', '💾');
            this.emit('cache_update', data);
        });
        this.wsClient.on('guild_refreshed', (data) => {
            logger.info('Guild refreshed', '🔄');
            this.emit('guild_refreshed', data);
        });
        this.wsClient.on('members_refreshed', (data) => {
            logger.info('Members refreshed', '👥');
            this.emit('members_refreshed', data);
        });
        this.wsClient.on('cache_refresh', (data) => {
            logger.info('Cache auto-refresh', '🔄');
            this.emit('cache_refresh', data);
        });
        this.wsClient.on('cache_delete', (data) => {
            logger.info('Cache item deleted', '🗑️');
            this.emit('cache_delete', data);
        });
        this.wsClient.on('cache_clear', (data) => {
            logger.info('Cache cleared', '🧹');
            this.emit('cache_clear', data);
        });
    }
    handleWebSocketEvent(event) {
        this.emit('websocket_event', event);
    }
    async connect() {
        try {
            await this.wsClient.connect();
            logger.info('Discord API Client started successfully', '🚀');
        }
        catch (error) {
            logger.error(`Failed to connect: ${error}`, '❌');
            throw error;
        }
    }
    disconnect() {
        this.wsClient.disconnect();
        logger.info('Discord API Client stopped', '🛑');
    }
    isConnected() {
        return this.wsClient.isConnected();
    }
    subscribe(guildId) {
        this.wsClient.subscribe(guildId);
    }
    unsubscribe() {
        this.wsClient.unsubscribe();
    }
    ping() {
        this.wsClient.ping();
    }
    getApiClient() {
        return this.apiClient;
    }
    getWebSocketClient() {
        return this.wsClient;
    }
}
export { ApiClient } from './api/client';
export { WebSocketClient } from './websocket/client';
export { config } from './config';
export { logger } from './utils/logger';
export * from './types';
export default DiscordApiClient;
//# sourceMappingURL=index.js.map