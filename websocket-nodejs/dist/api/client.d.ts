import { APIResponse, DiscordGuild, DiscordGuildMember, DiscordProfile, ServerStats, CacheStats, WebSocketStats } from '../types';
export declare class ApiClient {
    private client;
    private config;
    constructor();
    private setupInterceptors;
    private makeRequest;
    getGuilds(): Promise<APIResponse<DiscordGuild[]>>;
    getGuild(guildId: string): Promise<APIResponse<DiscordGuild>>;
    getGuildById(guildId: string): Promise<APIResponse<DiscordGuild>>;
    getGuildMembers(guildId: string, limit?: number): Promise<APIResponse<DiscordGuildMember[]>>;
    getUser(userId: string): Promise<APIResponse<DiscordProfile>>;
    refreshGuild(guildId: string): Promise<APIResponse<void>>;
    refreshGuildMembers(guildId: string, limit?: number): Promise<APIResponse<void>>;
    getStats(): Promise<APIResponse<ServerStats>>;
    getCacheStats(): Promise<APIResponse<CacheStats>>;
    getWebSocketStats(): Promise<APIResponse<WebSocketStats>>;
    clearCache(): Promise<APIResponse<void>>;
    healthCheck(): Promise<APIResponse<any>>;
    getServerInfo(): Promise<APIResponse<any>>;
}
//# sourceMappingURL=client.d.ts.map