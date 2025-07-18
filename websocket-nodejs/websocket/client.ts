import WebSocket from 'ws';
import { EventEmitter } from 'events';
import { config } from '../config';
import { logger } from '../utils/logger';
import { WebSocketEvent, WebSocketMessage } from '../types';

export interface WebSocketClientOptions {
  guildId?: string;
  userId?: string;
  autoReconnect?: boolean;
  pingInterval?: number;
  pongTimeout?: number;
}

export class WebSocketClient extends EventEmitter {
  private ws: WebSocket | null = null;
  private url: string;
  private options: Required<WebSocketClientOptions>;
  private reconnectAttempts = 0;
  private pingTimer: NodeJS.Timeout | null = null;
  private pongTimer: NodeJS.Timeout | null = null;
  private isConnecting = false;
  private isReconnecting = false;

  constructor(options: WebSocketClientOptions = {}) {
    super();
    
    this.options = {
      guildId: options.guildId || '',
      userId: options.userId || '',
      autoReconnect: options.autoReconnect ?? true,
      pingInterval: options.pingInterval || config.get().websocket.pingInterval,
      pongTimeout: options.pongTimeout || config.get().websocket.pongTimeout,
    };

    const params: Record<string, string> = {};
    if (this.options.guildId) params['guild_id'] = this.options.guildId;
    if (this.options.userId) params['user_id'] = this.options.userId;

    this.url = config.getWebSocketUrlWithParams(params);
  }

  public connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        resolve();
        return;
      }

      if (this.isConnecting) {
        reject(new Error('Connection already in progress'));
        return;
      }

      this.isConnecting = true;
      logger.info(`Connecting to WebSocket: ${this.url}`, '🔗');

      try {
        this.ws = new WebSocket(this.url);

        this.ws.on('open', () => {
          this.isConnecting = false;
          this.isReconnecting = false;
          this.reconnectAttempts = 0;
          this.startPingTimer();
          logger.logConnectionEstablished();
          this.emit('connected');
          resolve();
        });

        this.ws.on('message', (data: WebSocket.Data) => {
          try {
            const event: WebSocketEvent = JSON.parse(data.toString());
            this.handleMessage(event);
          } catch (error) {
            logger.error(`Failed to parse WebSocket message: ${error}`, '❌');
          }
        });

        this.ws.on('close', (code: number, reason: Buffer) => {
          this.handleClose(code, reason.toString());
        });

        this.ws.on('error', (error: Error) => {
          this.isConnecting = false;
          logger.logConnectionError(error.message);
          this.emit('error', error);
          reject(error);
        });

        this.ws.on('pong', () => {
          this.handlePong();
        });

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  public disconnect(): void {
    this.stopPingTimer();
    this.stopPongTimer();
    
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    
    logger.logConnectionClosed();
    this.emit('disconnected');
  }

  public send(message: WebSocketMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const messageStr = JSON.stringify(message);
      this.ws.send(messageStr);
      logger.debug(`Sent message: ${messageStr}`, '📤');
    } else {
      logger.warn('WebSocket is not connected, cannot send message', '⚠️');
    }
  }

  public subscribe(guildId: string): void {
    this.send({
      type: 'subscribe',
      guild_id: guildId,
    });
    logger.info(`Subscribed to guild: ${guildId}`, '📡');
  }

  public unsubscribe(): void {
    this.send({
      type: 'unsubscribe',
    });
    logger.info('Unsubscribed from all guilds', '📡');
  }

  public ping(): void {
    this.send({
      type: 'ping',
    });
  }

  public isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  public getReadyState(): number {
    return this.ws?.readyState || WebSocket.CLOSED;
  }

  private handleMessage(event: WebSocketEvent): void {
    logger.logWebSocketEvent(event.type, event.data);
    this.emit('message', event);
    this.emit(event.type, event.data, event);
  }

  private handleClose(code: number, reason: string): void {
    this.stopPingTimer();
    this.stopPongTimer();
    
    logger.info(`WebSocket closed: ${code} - ${reason}`, '🔌');
    this.emit('closed', code, reason);

    if (this.options.autoReconnect && !this.isReconnecting) {
      this.scheduleReconnect();
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= config.get().websocket.maxReconnectAttempts) {
      logger.error('Max reconnection attempts reached', '❌');
      this.emit('maxReconnectAttemptsReached');
      return;
    }

    this.isReconnecting = true;
    this.reconnectAttempts++;
    
    const delay = config.get().websocket.reconnectInterval * this.reconnectAttempts;
    logger.logReconnect(this.reconnectAttempts, config.get().websocket.maxReconnectAttempts);
    
    setTimeout(() => {
      this.connect().catch((error) => {
        logger.error(`Reconnection failed: ${error.message}`, '❌');
        this.scheduleReconnect();
      });
    }, delay);
  }

  private startPingTimer(): void {
    this.stopPingTimer();
    
    this.pingTimer = setInterval(() => {
      if (this.isConnected()) {
        this.ping();
        this.startPongTimer();
      }
    }, this.options.pingInterval);
  }

  private stopPingTimer(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
  }

  private startPongTimer(): void {
    this.stopPongTimer();
    
    this.pongTimer = setTimeout(() => {
      logger.warn('Pong timeout, closing connection', '⏰');
      this.disconnect();
    }, this.options.pongTimeout);
  }

  private stopPongTimer(): void {
    if (this.pongTimer) {
      clearTimeout(this.pongTimer);
      this.pongTimer = null;
    }
  }

  private handlePong(): void {
    this.stopPongTimer();
    logger.debug('Received pong', '🏓');
  }
} 