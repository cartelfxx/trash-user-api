import axios, { AxiosInstance, AxiosResponse } from 'axios';
import { config } from '../config';
import { logger } from '../utils/logger';
import {
  APIResponse,
  DiscordGuild,
  DiscordGuildMember,
  DiscordProfile,
  ServerStats,
  CacheStats,
  WebSocketStats,
} from '../types';

export class ApiClient {
  private client: AxiosInstance;
  private config = config.get();

  constructor() {
    this.client = axios.create({
      baseURL: config.getServerUrl(),
      timeout: this.config.api.timeout,
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'Discord-API-Client/1.0',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    this.client.interceptors.request.use(
      (config) => {
        logger.logApiRequest(config.method?.toUpperCase() || 'GET', config.url || '');
        return config;
      },
      (error) => {
        logger.error(`Request error: ${error.message}`, '❌');
        return Promise.reject(error);
      }
    );

    this.client.interceptors.response.use(
      (response: AxiosResponse) => {
        return response;
      },
      (error) => {
        if (error.response) {
          logger.error(`Response error: ${error.response.status} - ${error.response.statusText}`, '❌');
        } else {
          logger.error(`Network error: ${error.message}`, '❌');
        }
        return Promise.reject(error);
      }
    );
  }

  private async makeRequest<T>(method: string, url: string, data?: any): Promise<APIResponse<T>> {
    const startTime = Date.now();

    try {
      const response = await this.client.request<APIResponse<T>>({
        method,
        url,
        data,
      });

      const duration = Date.now() - startTime;
      logger.logApiResponse(response.status, url, duration);

      return response.data;
    } catch (error: any) {
      if (error.response?.data) {
        return error.response.data;
      }
      throw error;
    }
  }

  public async getGuilds(): Promise<APIResponse<DiscordGuild[]>> {
    return this.makeRequest<DiscordGuild[]>('GET', '/guilds');
  }

  public async getGuild(guildId: string): Promise<APIResponse<DiscordGuild>> {
    return this.makeRequest<DiscordGuild>('GET', `/guilds?id=${guildId}`);
  }

  public async getGuildById(guildId: string): Promise<APIResponse<DiscordGuild>> {
    return this.makeRequest<DiscordGuild>('GET', `/guilds/${guildId}`);
  }

  public async getGuildMembers(guildId: string, limit: number = 1000): Promise<APIResponse<DiscordGuildMember[]>> {
    return this.makeRequest<DiscordGuildMember[]>('GET', `/guilds/members?guild_id=${guildId}&limit=${limit}`);
  }

  public async getUser(userId: string): Promise<APIResponse<DiscordProfile>> {
    return this.makeRequest<DiscordProfile>('GET', `/users?id=${userId}`);
  }

  public async refreshGuild(guildId: string): Promise<APIResponse<void>> {
    return this.makeRequest<void>('POST', `/guilds/refresh?guild_id=${guildId}`);
  }

  public async refreshGuildMembers(guildId: string, limit: number = 1000): Promise<APIResponse<void>> {
    return this.makeRequest<void>('POST', `/guilds/members/refresh?guild_id=${guildId}&limit=${limit}`);
  }

  public async getStats(): Promise<APIResponse<ServerStats>> {
    return this.makeRequest<ServerStats>('GET', '/stats');
  }

  public async getCacheStats(): Promise<APIResponse<CacheStats>> {
    return this.makeRequest<CacheStats>('GET', '/cache/stats');
  }

  public async getWebSocketStats(): Promise<APIResponse<WebSocketStats>> {
    return this.makeRequest<WebSocketStats>('GET', '/websocket/stats');
  }

  public async clearCache(): Promise<APIResponse<void>> {
    return this.makeRequest<void>('POST', '/cache/clear');
  }

  public async healthCheck(): Promise<APIResponse<any>> {
    return this.makeRequest<any>('GET', '/health');
  }

  public async getServerInfo(): Promise<APIResponse<any>> {
    return this.makeRequest<any>('GET', '/');
  }
} 