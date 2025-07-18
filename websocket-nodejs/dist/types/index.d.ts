export interface DiscordUser {
    id: string;
    username: string;
    global_name: string;
    discriminator: string;
    avatar: string;
    avatar_decoration_data: any;
    collectibles: any;
    verified: boolean;
    mfa_enabled: boolean;
    premium_type?: number;
    public_flags: number;
    flags: number;
    banner: string;
    banner_color: string;
    accent_color: number;
    bio: string;
    primary_guild: {
        identity_guild_id: string;
        identity_enabled: any;
        tag: any;
        badge: any;
    };
    clan: {
        identity_guild_id: string;
        identity_enabled: any;
        tag: any;
        badge: any;
    };
}
export interface DiscordProfile {
    user: DiscordUser;
    connected_accounts: Array<{
        id: string;
        name: string;
        type: string;
    }>;
    premium_type: number;
    premium_since: string;
    premium_guild_since: string;
    profile_themes_experiment_bucket: number;
    user_profile: {
        bio: string;
        accent_color: number;
        pronouns: string;
    };
    badges: Array<{
        id: string;
        description: string;
        icon: string;
        link: string;
    }>;
    guild_badges: any[];
    mutual_guilds: Array<{
        id: string;
        nick: string;
    }>;
}
export interface DiscordRole {
    id: string;
    name: string;
    description: string;
    permissions: string;
    position: number;
    color: number;
    colors: any;
    hoist: boolean;
    managed: boolean;
    mentionable: boolean;
    icon: string;
    unicode_emoji: string;
    flags: number;
}
export interface DiscordEmoji {
    id: string;
    name: string;
    roles: string[];
    require_colons: boolean;
    managed: boolean;
    animated: boolean;
    available: boolean;
}
export interface DiscordGuild {
    id: string;
    name: string;
    icon: string;
    description: string;
    home_header: string;
    splash: string;
    discovery_splash: string;
    features: string[];
    banner: string;
    owner_id: string;
    application_id: string;
    region: string;
    afk_channel_id: string;
    afk_timeout: number;
    system_channel_id: string;
    system_channel_flags: number;
    widget_enabled: boolean;
    widget_channel_id: string;
    verification_level: number;
    roles: DiscordRole[];
    default_message_notifications: number;
    mfa_level: number;
    explicit_content_filter: number;
    max_presences: number;
    max_members: number;
    max_stage_video_channel_users: number;
    max_video_channel_users: number;
    vanity_url_code: string;
    premium_tier: number;
    premium_subscription_count: number;
    preferred_locale: string;
    rules_channel_id: string;
    safety_alerts_channel_id: string;
    public_updates_channel_id: string;
    hub_type: string;
    premium_progress_bar_enabled: boolean;
    latest_onboarding_question_id: string;
    nsfw: boolean;
    nsfw_level: number;
    owner_configured_content_level: number;
    emojis: DiscordEmoji[];
    stickers: any[];
    incidents_data: any;
    inventory_settings: any;
    embed_enabled: boolean;
    embed_channel_id: string;
    owner: boolean;
    permissions: string;
    approximate_member_count?: number;
    approximate_presence_count?: number;
}
export interface DiscordGuildMember {
    user: DiscordUser;
    nick: string;
    roles: string[];
    joined_at: string;
    premium_since?: string;
    avatar?: string;
    communication_disabled_until?: string;
}
export interface APIResponse<T = any> {
    success: boolean;
    data?: T;
    error?: string;
    message?: string;
    timestamp: string;
    count?: number;
    rate_limit?: RateLimit;
}
export interface RateLimit {
    limit: number;
    remaining: number;
    reset: number;
    reset_time: string;
}
export interface WebSocketEvent {
    type: string;
    data: any;
    timestamp: string;
    guild_id?: string;
    user_id?: string;
}
export interface CacheUpdateEvent {
    type: string;
    key: string;
    timestamp: string;
    data: any;
}
export interface WebSocketMessage {
    type: 'subscribe' | 'unsubscribe' | 'ping';
    guild_id?: string;
    user_id?: string;
}
export interface WebSocketStats {
    connected_clients: number;
    clients_info: Array<{
        user_id: string;
        guild_id: string;
        address: string;
    }>;
}
export interface CacheStats {
    hits: number;
    misses: number;
    evictions: number;
    refreshes: number;
    size: number;
    last_cleanup: string;
}
export interface ServerStats {
    cache: CacheStats;
    rate_limit: RateLimit;
    websocket: WebSocketStats;
    server: {
        start_time: string;
    };
}
//# sourceMappingURL=index.d.ts.map