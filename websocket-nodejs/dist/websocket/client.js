import WebSocket from 'ws';
import { EventEmitter } from 'events';
import { config } from '../config';
import { logger } from '../utils/logger';
export class WebSocketClient extends EventEmitter {
    constructor(options = {}) {
        super();
        this.ws = null;
        this.reconnectAttempts = 0;
        this.pingTimer = null;
        this.pongTimer = null;
        this.isConnecting = false;
        this.isReconnecting = false;
        this.options = {
            guildId: options.guildId || '',
            userId: options.userId || '',
            autoReconnect: options.autoReconnect ?? true,
            pingInterval: options.pingInterval || config.get().websocket.pingInterval,
            pongTimeout: options.pongTimeout || config.get().websocket.pongTimeout,
        };
        const params = {};
        if (this.options.guildId)
            params['guild_id'] = this.options.guildId;
        if (this.options.userId)
            params['user_id'] = this.options.userId;
        this.url = config.getWebSocketUrlWithParams(params);
    }
    connect() {
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
                this.ws.on('message', (data) => {
                    try {
                        const event = JSON.parse(data.toString());
                        this.handleMessage(event);
                    }
                    catch (error) {
                        logger.error(`Failed to parse WebSocket message: ${error}`, '❌');
                    }
                });
                this.ws.on('close', (code, reason) => {
                    this.handleClose(code, reason.toString());
                });
                this.ws.on('error', (error) => {
                    this.isConnecting = false;
                    logger.logConnectionError(error.message);
                    this.emit('error', error);
                    reject(error);
                });
                this.ws.on('pong', () => {
                    this.handlePong();
                });
            }
            catch (error) {
                this.isConnecting = false;
                reject(error);
            }
        });
    }
    disconnect() {
        this.stopPingTimer();
        this.stopPongTimer();
        if (this.ws) {
            this.ws.close(1000, 'Client disconnect');
            this.ws = null;
        }
        logger.logConnectionClosed();
        this.emit('disconnected');
    }
    send(message) {
        if (this.ws?.readyState === WebSocket.OPEN) {
            const messageStr = JSON.stringify(message);
            this.ws.send(messageStr);
            logger.debug(`Sent message: ${messageStr}`, '📤');
        }
        else {
            logger.warn('WebSocket is not connected, cannot send message', '⚠️');
        }
    }
    subscribe(guildId) {
        this.send({
            type: 'subscribe',
            guild_id: guildId,
        });
        logger.info(`Subscribed to guild: ${guildId}`, '📡');
    }
    unsubscribe() {
        this.send({
            type: 'unsubscribe',
        });
        logger.info('Unsubscribed from all guilds', '📡');
    }
    ping() {
        this.send({
            type: 'ping',
        });
    }
    isConnected() {
        return this.ws?.readyState === WebSocket.OPEN;
    }
    getReadyState() {
        return this.ws?.readyState || WebSocket.CLOSED;
    }
    handleMessage(event) {
        logger.logWebSocketEvent(event.type, event.data);
        this.emit('message', event);
        this.emit(event.type, event.data, event);
    }
    handleClose(code, reason) {
        this.stopPingTimer();
        this.stopPongTimer();
        logger.info(`WebSocket closed: ${code} - ${reason}`, '🔌');
        this.emit('closed', code, reason);
        if (this.options.autoReconnect && !this.isReconnecting) {
            this.scheduleReconnect();
        }
    }
    scheduleReconnect() {
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
    startPingTimer() {
        this.stopPingTimer();
        this.pingTimer = setInterval(() => {
            if (this.isConnected()) {
                this.ping();
                this.startPongTimer();
            }
        }, this.options.pingInterval);
    }
    stopPingTimer() {
        if (this.pingTimer) {
            clearInterval(this.pingTimer);
            this.pingTimer = null;
        }
    }
    startPongTimer() {
        this.stopPongTimer();
        this.pongTimer = setTimeout(() => {
            logger.warn('Pong timeout, closing connection', '⏰');
            this.disconnect();
        }, this.options.pongTimeout);
    }
    stopPongTimer() {
        if (this.pongTimer) {
            clearTimeout(this.pongTimer);
            this.pongTimer = null;
        }
    }
    handlePong() {
        this.stopPongTimer();
        logger.debug('Received pong', '🏓');
    }
}
//# sourceMappingURL=client.js.map