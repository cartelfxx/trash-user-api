package models

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	AvatarDecorationData interface{} `json:"avatar_decoration_data"`
	Collectibles  interface{} `json:"collectibles"`
	Verified      bool   `json:"verified"`
	MFAEnabled    bool   `json:"mfa_enabled"`
	PremiumType   int    `json:"premium_type,omitempty"`
	PublicFlags   int    `json:"public_flags"`
	Flags         int    `json:"flags"`
	Banner        string `json:"banner"`
	BannerColor   string `json:"banner_color"`
	AccentColor   int    `json:"accent_color"`
	Bio           string `json:"bio"`
	PrimaryGuild  struct {
		IdentityGuildID string `json:"identity_guild_id"`
		IdentityEnabled interface{} `json:"identity_enabled"`
		Tag            interface{} `json:"tag"`
		Badge          interface{} `json:"badge"`
	} `json:"primary_guild"`
	Clan struct {
		IdentityGuildID string `json:"identity_guild_id"`
		IdentityEnabled interface{} `json:"identity_enabled"`
		Tag            interface{} `json:"tag"`
		Badge          interface{} `json:"badge"`
	} `json:"clan"`
}

type DiscordProfile struct {
	User         DiscordUser `json:"user"`
	ConnectedAccounts []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"connected_accounts"`
	PremiumType   int `json:"premium_type"`
	PremiumSince  string `json:"premium_since"`
	PremiumGuildSince string `json:"premium_guild_since"`
	ProfileThemesExperimentBucket int `json:"profile_themes_experiment_bucket"`
	UserProfile   struct {
		Bio         string `json:"bio"`
		AccentColor int    `json:"accent_color"`
		Pronouns    string `json:"pronouns"`
	} `json:"user_profile"`
	Badges []struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		Link        string `json:"link"`
	} `json:"badges"`
	GuildBadges []interface{} `json:"guild_badges"`
	MutualGuilds []struct {
		ID   string `json:"id"`
		Nick string `json:"nick"`
	} `json:"mutual_guilds"`
}

type DiscordRole struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Permissions   string `json:"permissions"`
	Position      int    `json:"position"`
	Color         int    `json:"color"`
	Colors        interface{} `json:"colors"`
	Hoist         bool   `json:"hoist"`
	Managed       bool   `json:"managed"`
	Mentionable   bool   `json:"mentionable"`
	Icon          string `json:"icon"`
	UnicodeEmoji  string `json:"unicode_emoji"`
	Flags         int    `json:"flags"`
}

type DiscordEmoji struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Roles          []string `json:"roles"`
	RequireColons  bool     `json:"require_colons"`
	Managed        bool     `json:"managed"`
	Animated       bool     `json:"animated"`
	Available      bool     `json:"available"`
}

type DiscordGuild struct {
	ID                          string         `json:"id"`
	Name                        string         `json:"name"`
	Icon                        string         `json:"icon"`
	Description                 string         `json:"description"`
	HomeHeader                  string         `json:"home_header"`
	Splash                      string         `json:"splash"`
	DiscoverySplash             string         `json:"discovery_splash"`
	Features                    []string       `json:"features"`
	Banner                      string         `json:"banner"`
	OwnerID                     string         `json:"owner_id"`
	ApplicationID               string         `json:"application_id"`
	Region                      string         `json:"region"`
	AFKChannelID                string         `json:"afk_channel_id"`
	AFKTimeout                  int            `json:"afk_timeout"`
	SystemChannelID             string         `json:"system_channel_id"`
	SystemChannelFlags          int            `json:"system_channel_flags"`
	WidgetEnabled               bool           `json:"widget_enabled"`
	WidgetChannelID             string         `json:"widget_channel_id"`
	VerificationLevel           int            `json:"verification_level"`
	Roles                       []DiscordRole  `json:"roles"`
	DefaultMessageNotifications int            `json:"default_message_notifications"`
	MFALevel                    int            `json:"mfa_level"`
	ExplicitContentFilter       int            `json:"explicit_content_filter"`
	MaxPresences                int            `json:"max_presences"`
	MaxMembers                  int            `json:"max_members"`
	MaxStageVideoChannelUsers   int            `json:"max_stage_video_channel_users"`
	MaxVideoChannelUsers        int            `json:"max_video_channel_users"`
	VanityURLCode               string         `json:"vanity_url_code"`
	PremiumTier                 int            `json:"premium_tier"`
	PremiumSubscriptionCount    int            `json:"premium_subscription_count"`
	PreferredLocale             string         `json:"preferred_locale"`
	RulesChannelID              string         `json:"rules_channel_id"`
	SafetyAlertsChannelID       string         `json:"safety_alerts_channel_id"`
	PublicUpdatesChannelID      string         `json:"public_updates_channel_id"`
	HubType                     string         `json:"hub_type"`
	PremiumProgressBarEnabled   bool           `json:"premium_progress_bar_enabled"`
	LatestOnboardingQuestionID  string         `json:"latest_onboarding_question_id"`
	NSFW                        bool           `json:"nsfw"`
	NSFWLevel                   int            `json:"nsfw_level"`
	OwnerConfiguredContentLevel int            `json:"owner_configured_content_level"`
	Emojis                      []DiscordEmoji `json:"emojis"`
	Stickers                    []interface{}  `json:"stickers"`
	IncidentsData               interface{}    `json:"incidents_data"`
	InventorySettings           interface{}    `json:"inventory_settings"`
	EmbedEnabled                bool           `json:"embed_enabled"`
	EmbedChannelID              string         `json:"embed_channel_id"`
	
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
	ApproximateMemberCount int `json:"approximate_member_count,omitempty"`
	ApproximatePresenceCount int `json:"approximate_presence_count,omitempty"`
}

type DiscordGuildMember struct {
	User      DiscordUser `json:"user"`
	Nick      string      `json:"nick"`
	Roles     []string    `json:"roles"`
	JoinedAt  string      `json:"joined_at"`
	PremiumSince string   `json:"premium_since,omitempty"`
	Avatar    string      `json:"avatar,omitempty"`
	CommunicationDisabledUntil string `json:"communication_disabled_until,omitempty"`
}

type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp string      `json:"timestamp"`
	Count     int         `json:"count,omitempty"`
	RateLimit *RateLimit  `json:"rate_limit,omitempty"`
}

type RateLimit struct {
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	Reset     int64  `json:"reset"`
	ResetTime string `json:"reset_time"`
}

type WebSocketEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
	GuildID   string      `json:"guild_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
}

type CacheUpdateEvent struct {
	Type      string `json:"type"`
	Key       string `json:"key"`
	Timestamp string `json:"timestamp"`
	Data      interface{} `json:"data"`
} 