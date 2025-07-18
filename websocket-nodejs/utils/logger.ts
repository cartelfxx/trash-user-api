import { config } from '../config';

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

export class Logger {
  private static instance: Logger;
  private config = config.get().logging;

  private constructor() {}

  public static getInstance(): Logger {
    if (!Logger.instance) {
      Logger.instance = new Logger();
    }
    return Logger.instance;
  }

  private getTimestamp(): string {
    if (!this.config.enableTimestamps) return '';
    return `[${new Date().toISOString()}] `;
  }

  private shouldLog(level: LogLevel): boolean {
    const configLevel = LogLevel[this.config.level.toUpperCase() as keyof typeof LogLevel];
    return level >= configLevel;
  }

  private formatMessage(level: string, message: string, emoji?: string): string {
    const timestamp = this.getTimestamp();
    const emojiStr = emoji ? `${emoji} ` : '';
    return `${timestamp}${level} ${emojiStr}${message}`;
  }

  public debug(message: string, emoji?: string): void {
    if (this.shouldLog(LogLevel.DEBUG)) {
      console.debug(this.formatMessage('DEBUG', message, emoji));
    }
  }

  public info(message: string, emoji?: string): void {
    if (this.shouldLog(LogLevel.INFO)) {
      console.info(this.formatMessage('INFO', message, emoji));
    }
  }

  public warn(message: string, emoji?: string): void {
    if (this.shouldLog(LogLevel.WARN)) {
      console.warn(this.formatMessage('WARN', message, emoji));
    }
  }

  public error(message: string, emoji?: string): void {
    if (this.shouldLog(LogLevel.ERROR)) {
      console.error(this.formatMessage('ERROR', message, emoji));
    }
  }

  public logWebSocketEvent(event: string, data?: any): void {
    this.info(`WebSocket Event: ${event}`, '🔌');
    if (data) {
      this.debug(JSON.stringify(data, null, 2), '📄');
    }
  }

  public logApiRequest(method: string, url: string): void {
    this.info(`${method} ${url}`, '🌐');
  }

  public logApiResponse(status: number, url: string, duration?: number): void {
    const durationStr = duration ? ` (${duration}ms)` : '';
    this.info(`${status} ${url}${durationStr}`, '✅');
  }

  public logCacheHit(key: string): void {
    this.debug(`Cache hit: ${key}`, '📤');
  }

  public logCacheMiss(key: string): void {
    this.debug(`Cache miss: ${key}`, '❌');
  }

  public logReconnect(attempt: number, maxAttempts: number): void {
    this.warn(`Reconnecting... (${attempt}/${maxAttempts})`, '🔄');
  }

  public logConnectionEstablished(): void {
    this.info('WebSocket connection established', '🔗');
  }

  public logConnectionClosed(): void {
    this.info('WebSocket connection closed', '🔌');
  }

  public logConnectionError(error: string): void {
    this.error(`WebSocket connection error: ${error}`, '❌');
  }
}

export const logger = Logger.getInstance(); 