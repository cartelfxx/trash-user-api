export declare enum LogLevel {
    DEBUG = 0,
    INFO = 1,
    WARN = 2,
    ERROR = 3
}
export declare class Logger {
    private static instance;
    private config;
    private constructor();
    static getInstance(): Logger;
    private getTimestamp;
    private shouldLog;
    private formatMessage;
    debug(message: string, emoji?: string): void;
    info(message: string, emoji?: string): void;
    warn(message: string, emoji?: string): void;
    error(message: string, emoji?: string): void;
    logWebSocketEvent(event: string, data?: any): void;
    logApiRequest(method: string, url: string): void;
    logApiResponse(status: number, url: string, duration?: number): void;
    logCacheHit(key: string): void;
    logCacheMiss(key: string): void;
    logReconnect(attempt: number, maxAttempts: number): void;
    logConnectionEstablished(): void;
    logConnectionClosed(): void;
    logConnectionError(error: string): void;
}
export declare const logger: Logger;
//# sourceMappingURL=logger.d.ts.map