import axios from 'axios';
import { config } from '../config';
import { logger } from '../utils/logger';
export class ApiClient {
    constructor() {
        this.config = config.get();
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
    setupInterceptors() {
        this.client.interceptors.request.use((config) => {
            logger.logApiRequest(config.method?.toUpperCase() || 'GET', config.url || '');
            return config;
        }, (error) => {
            logger.error(`Request error: ${error.message}`, '❌');
            return Promise.reject(error);
        });
        this.client.interceptors.response.use((response) => {
            return response;
        }, (error) => {
            if (error.response) {
                logger.error(`Response error: ${error.response.status} - ${error.response.statusText}`, '❌');
            }
            else {
                logger.error(`Network error: ${error.message}`, '❌');
            }
            return Promise.reject(error);
        });
    }
    async makeRequest(method, url, data) {
        const startTime = Date.now();
        try {
            const response = await this.client.request({
                method,
                url,
                data,
            });
            const duration = Date.now() - startTime;
            logger.logApiResponse(response.status, url, duration);
            return response.data;
        }
        catch (error) {
            if (error.response?.data) {
                return error.response.data;
            }
            throw error;
        }
    }
    async getGuilds() {
        return this.makeRequest('GET', '/guilds');
    }
    async getGuild(guildId) {
        return this.makeRequest('GET', `/guilds?id=${guildId}`);
    }
    async getGuildById(guildId) {
        return this.makeRequest('GET', `/guilds/${guildId}`);
    }
    async getGuildMembers(guildId, limit = 1000) {
        return this.makeRequest('GET', `/guilds/members?guild_id=${guildId}&limit=${limit}`);
    }
    async getUser(userId) {
        return this.makeRequest('GET', `/users?id=${userId}`);
    }
    async refreshGuild(guildId) {
        return this.makeRequest('POST', `/guilds/refresh?guild_id=${guildId}`);
    }
    async refreshGuildMembers(guildId, limit = 1000) {
        return this.makeRequest('POST', `/guilds/members/refresh?guild_id=${guildId}&limit=${limit}`);
    }
    async getStats() {
        return this.makeRequest('GET', '/stats');
    }
    async getCacheStats() {
        return this.makeRequest('GET', '/cache/stats');
    }
    async getWebSocketStats() {
        return this.makeRequest('GET', '/websocket/stats');
    }
    async clearCache() {
        return this.makeRequest('POST', '/cache/clear');
    }
    async healthCheck() {
        return this.makeRequest('GET', '/health');
    }
    async getServerInfo() {
        return this.makeRequest('GET', '/');
    }
}
//# sourceMappingURL=client.js.map