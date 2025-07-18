import { EventEmitter } from 'events';
import { ApiClient } from './api/client';
import { WebSocketClient, WebSocketClientOptions } from './websocket/client';
import { config } from './config';
import { logger } from './utils/logger';
import { WebSocketEvent } from './types';

export interface DiscordApiClientOptions {
  guildId?: string;
  userId?: string;
  autoReconnect?: boolean;
  pingInterval?: number;
  pongTimeout?: number;
}

export class DiscordApiClient extends EventEmitter {
  private apiClient: ApiClient;
  private wsClient: WebSocketClient;
  private options: Required<DiscordApiClientOptions>;

  constructor(options: DiscordApiClientOptions = {}) {
    super();
    this.options = {
      guildId: options.guildId || '',
      userId: options.userId || '',
      autoReconnect: options.autoReconnect ?? true,
      pingInterval: options.pingInterval || config.get().websocket.pingInterval,
      pongTimeout: options.pongTimeout || config.get().websocket.pongTimeout,
    };

    this.apiClient = new ApiClient();
    
    const wsOptions: WebSocketClientOptions = {
      guildId: this.options.guildId,
      userId: this.options.userId,
      autoReconnect: this.options.autoReconnect,
      pingInterval: this.options.pingInterval,
      pongTimeout: this.options.pongTimeout,
    };
    
    this.wsClient = new WebSocketClient(wsOptions);
    this.setupWebSocketEvents();
  }

  private setupWebSocketEvents(): void {
    this.wsClient.on('connected', () => {
      logger.info('Discord API Client connected', '🔗');
    });

    this.wsClient.on('disconnected', () => {
      logger.info('Discord API Client disconnected', '🔌');
    });

    this.wsClient.on('error', (error: Error) => {
      logger.error(`WebSocket error: ${error.message}`, '❌');
    });

    this.wsClient.on('message', (event: WebSocketEvent) => {
      this.handleWebSocketEvent(event);
    });

    this.wsClient.on('cache_update', (data: any) => {
      logger.info('Cache updated', '💾');
      this.emit('cache_update', data);
    });

    this.wsClient.on('guild_refreshed', (data: any) => {
      logger.info('Guild refreshed', '🔄');
      this.emit('guild_refreshed', data);
    });

    this.wsClient.on('members_refreshed', (data: any) => {
      logger.info('Members refreshed', '👥');
      this.emit('members_refreshed', data);
    });

    this.wsClient.on('cache_refresh', (data: any) => {
      logger.info('Cache auto-refresh', '🔄');
      this.emit('cache_refresh', data);
    });

    this.wsClient.on('cache_delete', (data: any) => {
      logger.info('Cache item deleted', '🗑️');
      this.emit('cache_delete', data);
    });

    this.wsClient.on('cache_clear', (data: any) => {
      logger.info('Cache cleared', '🧹');
      this.emit('cache_clear', data);
    });
  }

  private handleWebSocketEvent(event: WebSocketEvent): void {
    this.emit('websocket_event', event);
  }

  public async connect(): Promise<void> {
    try {
      await this.wsClient.connect();
      logger.info('Discord API Client started successfully', '🚀');
    } catch (error) {
      logger.error(`Failed to connect: ${error}`, '❌');
      throw error;
    }
  }

  public disconnect(): void {
    this.wsClient.disconnect();
    logger.info('Discord API Client stopped', '🛑');
  }

  public isConnected(): boolean {
    return this.wsClient.isConnected();
  }

  public subscribe(guildId: string): void {
    this.wsClient.subscribe(guildId);
  }

  public unsubscribe(): void {
    this.wsClient.unsubscribe();
  }

  public ping(): void {
    this.wsClient.ping();
  }

  public getApiClient(): ApiClient {
    return this.apiClient;
  }

  public getWebSocketClient(): WebSocketClient {
    return this.wsClient;
  }
}

export { ApiClient } from './api/client';
export { WebSocketClient } from './websocket/client';
export { config } from './config';
export { logger } from './utils/logger';
export * from './types';

export default DiscordApiClient; 