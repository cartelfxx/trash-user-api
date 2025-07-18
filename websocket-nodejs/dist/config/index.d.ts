export interface ClientConfig {
    server: {
        host: string;
        port: number;
        protocol: 'http' | 'https';
        wsProtocol: 'ws' | 'wss';
    };
    websocket: {
        reconnectInterval: number;
        maxReconnectAttempts: number;
        pingInterval: number;
        pongTimeout: number;
    };
    api: {
        timeout: number;
        retries: number;
        retryDelay: number;
    };
    logging: {
        level: 'debug' | 'info' | 'warn' | 'error';
        enableColors: boolean;
        enableTimestamps: boolean;
    };
}
export declare class Config {
    private static instance;
    private config;
    private constructor();
    static getInstance(): Config;
    get(): ClientConfig;
    getServerUrl(): string;
    getWebSocketUrl(): string;
    getWebSocketUrlWithParams(params: Record<string, string>): string;
}
export declare const config: Config;
//# sourceMappingURL=index.d.ts.map