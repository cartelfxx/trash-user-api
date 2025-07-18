import { config } from '../config';
export var LogLevel;
(function (LogLevel) {
    LogLevel[LogLevel["DEBUG"] = 0] = "DEBUG";
    LogLevel[LogLevel["INFO"] = 1] = "INFO";
    LogLevel[LogLevel["WARN"] = 2] = "WARN";
    LogLevel[LogLevel["ERROR"] = 3] = "ERROR";
})(LogLevel || (LogLevel = {}));
export class Logger {
    constructor() {
        this.config = config.get().logging;
    }
    static getInstance() {
        if (!Logger.instance) {
            Logger.instance = new Logger();
        }
        return Logger.instance;
    }
    getTimestamp() {
        if (!this.config.enableTimestamps)
            return '';
        return `[${new Date().toISOString()}] `;
    }
    shouldLog(level) {
        const configLevel = LogLevel[this.config.level.toUpperCase()];
        return level >= configLevel;
    }
    formatMessage(level, message, emoji) {
        const timestamp = this.getTimestamp();
        const emojiStr = emoji ? `${emoji} ` : '';
        return `${timestamp}${level} ${emojiStr}${message}`;
    }
    debug(message, emoji) {
        if (this.shouldLog(LogLevel.DEBUG)) {
            console.debug(this.formatMessage('DEBUG', message, emoji));
        }
    }
    info(message, emoji) {
        if (this.shouldLog(LogLevel.INFO)) {
            console.info(this.formatMessage('INFO', message, emoji));
        }
    }
    warn(message, emoji) {
        if (this.shouldLog(LogLevel.WARN)) {
            console.warn(this.formatMessage('WARN', message, emoji));
        }
    }
    error(message, emoji) {
        if (this.shouldLog(LogLevel.ERROR)) {
            console.error(this.formatMessage('ERROR', message, emoji));
        }
    }
    logWebSocketEvent(event, data) {
        this.info(`WebSocket Event: ${event}`, '🔌');
        if (data) {
            this.debug(JSON.stringify(data, null, 2), '📄');
        }
    }
    logApiRequest(method, url) {
        this.info(`${method} ${url}`, '🌐');
    }
    logApiResponse(status, url, duration) {
        const durationStr = duration ? ` (${duration}ms)` : '';
        this.info(`${status} ${url}${durationStr}`, '✅');
    }
    logCacheHit(key) {
        this.debug(`Cache hit: ${key}`, '📤');
    }
    logCacheMiss(key) {
        this.debug(`Cache miss: ${key}`, '❌');
    }
    logReconnect(attempt, maxAttempts) {
        this.warn(`Reconnecting... (${attempt}/${maxAttempts})`, '🔄');
    }
    logConnectionEstablished() {
        this.info('WebSocket connection established', '🔗');
    }
    logConnectionClosed() {
        this.info('WebSocket connection closed', '🔌');
    }
    logConnectionError(error) {
        this.error(`WebSocket connection error: ${error}`, '❌');
    }
}
export const logger = Logger.getInstance();
//# sourceMappingURL=logger.js.map