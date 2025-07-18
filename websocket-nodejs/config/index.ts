import dotenv from 'dotenv';
import path from 'path';

dotenv.config({ path: path.resolve(process.cwd(), '.env') });

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

export class Config {
  private static instance: Config;
  private config: ClientConfig;

  private constructor() {
    this.config = {
      server: {
        host: process.env['SERVER_HOST'] || 'localhost',
        port: parseInt(process.env['SERVER_PORT'] || '8080', 10),
        protocol: (process.env['SERVER_PROTOCOL'] as 'http' | 'https') || 'http',
        wsProtocol: (process.env['WS_PROTOCOL'] as 'ws' | 'wss') || 'ws',
      },
      websocket: {
        reconnectInterval: parseInt(process.env['WS_RECONNECT_INTERVAL'] || '5000', 10),
        maxReconnectAttempts: parseInt(process.env['WS_MAX_RECONNECT_ATTEMPTS'] || '10', 10),
        pingInterval: parseInt(process.env['WS_PING_INTERVAL'] || '30000', 10),
        pongTimeout: parseInt(process.env['WS_PONG_TIMEOUT'] || '10000', 10),
      },
      api: {
        timeout: parseInt(process.env['API_TIMEOUT'] || '30000', 10),
        retries: parseInt(process.env['API_RETRIES'] || '3', 10),
        retryDelay: parseInt(process.env['API_RETRY_DELAY'] || '1000', 10),
      },
      logging: {
        level: (process.env['LOG_LEVEL'] as 'debug' | 'info' | 'warn' | 'error') || 'info',
        enableColors: process.env['LOG_ENABLE_COLORS'] !== 'false',
        enableTimestamps: process.env['LOG_ENABLE_TIMESTAMPS'] !== 'false',
      },
    };
  }

  public static getInstance(): Config {
    if (!Config.instance) {
      Config.instance = new Config();
    }
    return Config.instance;
  }

  public get(): ClientConfig {
    return this.config;
  }

  public getServerUrl(): string {
    const { protocol, host, port } = this.config.server;
    return `${protocol}://${host}:${port}`;
  }

  public getWebSocketUrl(): string {
    const { wsProtocol, host, port } = this.config.server;
    return `${wsProtocol}://${host}:${port}/websocket`;
  }

  public getWebSocketUrlWithParams(params: Record<string, string>): string {
    const baseUrl = this.getWebSocketUrl();
    const searchParams = new URLSearchParams(params);
    return `${baseUrl}?${searchParams.toString()}`;
  }
}

export const config = Config.getInstance(); 