package discord

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"discord-user-api/cache"
	"discord-user-api/config"
	"discord-user-api/models"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client
	cache      *cache.Cache
	rateLimiter *RateLimiter
}

type RateLimiter struct {
	limit     int
	remaining int
	reset     int64
	resetTime time.Time
}

func NewClient(cfg *config.Config, cache *cache.Cache) *Client {
	client := &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Discord.RequestTimeout,
		},
		cache:      cache,
		rateLimiter: &RateLimiter{},
	}

	log.Printf("🤖 Discord Client başlatıldı")
	return client
}

func (c *Client) GetGuilds() ([]models.DiscordGuild, error) {
	cacheKey := "guilds"
	
	if c.config.Cache.Enabled {
		if cached, exists := c.cache.Get(cacheKey); exists {
			if guilds, ok := cached.([]models.DiscordGuild); ok {
				log.Printf("📤 Cache'den guild'ler getirildi: %d adet", len(guilds))
				return guilds, nil
			}
		}
	}

	url := fmt.Sprintf("%s/%s/users/@me/guilds", c.config.Discord.APIURL, c.config.Discord.APIVersion)
	
	guilds, err := c.makeRequest("GET", url, nil, func(body io.Reader) (interface{}, error) {
		var guilds []models.DiscordGuild
		err := json.NewDecoder(body).Decode(&guilds)
		return guilds, err
	})

	if err != nil {
		return nil, fmt.Errorf("guild'ler getirilemedi: %v", err)
	}

	guildsList := guilds.([]models.DiscordGuild)
	
	if c.config.Cache.Enabled {
		c.cache.SetWithAutoRefresh(cacheKey, guildsList, c.config.Cache.TTL, true, 5*time.Minute)
	}

	log.Printf("✅ Guild'ler başarıyla getirildi: %d adet", len(guildsList))
	return guildsList, nil
}

func (c *Client) GetUser(userID string) (*models.DiscordProfile, error) {
	cacheKey := fmt.Sprintf("user_%s", userID)
	
	if c.config.Cache.Enabled {
		if cached, exists := c.cache.Get(cacheKey); exists {
			if profile, ok := cached.(*models.DiscordProfile); ok {
				log.Printf("📤 Cache'den kullanıcı getirildi: %s (%s)", userID, profile.User.Username)
				return profile, nil
			}
		}
	}

	url := fmt.Sprintf("%s/%s/users/%s/profile", c.config.Discord.APIURL, c.config.Discord.APIVersion, userID)
	
	profile, err := c.makeRequest("GET", url, nil, func(body io.Reader) (interface{}, error) {
		var profile models.DiscordProfile
		err := json.NewDecoder(body).Decode(&profile)
		return &profile, err
	})

	if err != nil {
		return nil, fmt.Errorf("kullanıcı profili getirilemedi: %v", err)
	}

	profileData := profile.(*models.DiscordProfile)
	
	if c.config.Cache.Enabled {
		c.cache.SetWithAutoRefresh(cacheKey, profileData, c.config.Cache.TTL, true, 10*time.Minute)
	}

	log.Printf("✅ Kullanıcı profili başarıyla getirildi: %s (%s)", userID, profileData.User.Username)
	return profileData, nil
}

func (c *Client) GetGuild(guildID string) (*models.DiscordGuild, error) {
	cacheKey := fmt.Sprintf("guild_%s", guildID)
	
	if c.config.Cache.Enabled {
		if cached, exists := c.cache.Get(cacheKey); exists {
			if guild, ok := cached.(*models.DiscordGuild); ok {
				log.Printf("📤 Cache'den guild getirildi: %s (%s)", guildID, guild.Name)
				return guild, nil
			}
		}
	}

	url := fmt.Sprintf("%s/%s/guilds/%s", c.config.Discord.APIURL, c.config.Discord.APIVersion, guildID)
	
	guild, err := c.makeRequest("GET", url, nil, func(body io.Reader) (interface{}, error) {
		var guild models.DiscordGuild
		err := json.NewDecoder(body).Decode(&guild)
		return &guild, err
	})

	if err != nil {
		return nil, fmt.Errorf("guild getirilemedi: %v", err)
	}

	guildData := guild.(*models.DiscordGuild)
	
	if c.config.Cache.Enabled {
		c.cache.SetWithAutoRefresh(cacheKey, guildData, c.config.Cache.TTL, true, 2*time.Minute)
	}

	log.Printf("✅ Guild başarıyla getirildi: %s (%s) - %d rol, %d emoji", 
		guildID, guildData.Name, len(guildData.Roles), len(guildData.Emojis))
	return guildData, nil
}

func (c *Client) GetGuildMembers(guildID string, limit int) ([]models.DiscordGuildMember, error) {
	cacheKey := fmt.Sprintf("guild_members_%s_%d", guildID, limit)
	
	if c.config.Cache.Enabled {
		if cached, exists := c.cache.Get(cacheKey); exists {
			if members, ok := cached.([]models.DiscordGuildMember); ok {
				log.Printf("📤 Cache'den guild üyeleri getirildi: %s (%d adet)", guildID, len(members))
				return members, nil
			}
		}
	}

	if limit <= 0 {
		limit = 1000
	}

	url := fmt.Sprintf("%s/%s/guilds/%s/members?limit=%d", c.config.Discord.APIURL, c.config.Discord.APIVersion, guildID, limit)
	
	members, err := c.makeRequest("GET", url, nil, func(body io.Reader) (interface{}, error) {
		var members []models.DiscordGuildMember
		err := json.NewDecoder(body).Decode(&members)
		return members, err
	})

	if err != nil {
		return nil, fmt.Errorf("guild üyeleri getirilemedi: %v", err)
	}

	membersList := members.([]models.DiscordGuildMember)
	
	if c.config.Cache.Enabled {
		c.cache.SetWithAutoRefresh(cacheKey, membersList, c.config.Cache.TTL, true, 3*time.Minute)
	}

	log.Printf("✅ Guild üyeleri başarıyla getirildi: %s (%d adet)", guildID, len(membersList))
	return membersList, nil
}

func (c *Client) RefreshGuild(guildID string) error {
	cacheKey := fmt.Sprintf("guild_%s", guildID)
	
	if c.config.Cache.Enabled {
		c.cache.Delete(cacheKey)
	}
	
	_, err := c.GetGuild(guildID)
	if err != nil {
		return fmt.Errorf("guild yenilenemedi: %v", err)
	}
	
	log.Printf("🔄 Guild yenilendi: %s", guildID)
	return nil
}

func (c *Client) RefreshGuildMembers(guildID string, limit int) error {
	cacheKey := fmt.Sprintf("guild_members_%s_%d", guildID, limit)
	
	if c.config.Cache.Enabled {
		c.cache.Delete(cacheKey)
	}
	
	_, err := c.GetGuildMembers(guildID, limit)
	if err != nil {
		return fmt.Errorf("guild üyeleri yenilenemedi: %v", err)
	}
	
	log.Printf("🔄 Guild üyeleri yenilendi: %s", guildID)
	return nil
}

func (c *Client) makeRequest(method, url string, body io.Reader, decoder func(io.Reader) (interface{}, error)) (interface{}, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.Discord.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("🔄 Yeniden deneme %d/%d", attempt, c.config.Discord.MaxRetries)
			time.Sleep(c.config.Discord.RetryDelay * time.Duration(attempt))
		}

		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("request oluşturulamadı: %v", err)
		}

		req.Header.Set("Authorization", c.config.Discord.Token)
		req.Header.Set("User-Agent", "Discord-API-Client/1.0")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		if c.rateLimiter.remaining <= 0 && c.rateLimiter.resetTime.After(time.Now()) {
			waitTime := time.Until(c.rateLimiter.resetTime)
			log.Printf("⏰ Rate limit bekleme: %v", waitTime)
			time.Sleep(waitTime)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP isteği başarısız: %v", err)
			continue
		}

		c.parseRateLimitHeaders(resp)

		switch resp.StatusCode {
		case http.StatusOK:
			result, err := decoder(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("response decode hatası: %v", err)
			}
			return result, nil

		case http.StatusTooManyRequests:
			resetTime := c.rateLimiter.resetTime
			if resetTime.After(time.Now()) {
				waitTime := time.Until(resetTime)
				log.Printf("⏰ Rate limit aşıldı, bekleme: %v", waitTime)
				time.Sleep(waitTime)
			}
			resp.Body.Close()
			continue

		case http.StatusUnauthorized:
			resp.Body.Close()
			return nil, fmt.Errorf("yetkilendirme hatası: geçersiz token")

		case http.StatusForbidden:
			resp.Body.Close()
			return nil, fmt.Errorf("erişim reddedildi: yetersiz izinler")

		case http.StatusNotFound:
			resp.Body.Close()
			return nil, fmt.Errorf("kaynak bulunamadı")

		default:
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("API hatası: %d - %s", resp.StatusCode, string(bodyBytes))
			continue
		}
	}

	return nil, fmt.Errorf("maksimum deneme sayısı aşıldı, son hata: %v", lastErr)
}

func (c *Client) parseRateLimitHeaders(resp *http.Response) {
	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			c.rateLimiter.limit = val
		}
	}

	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			c.rateLimiter.remaining = val
		}
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			c.rateLimiter.reset = val
			c.rateLimiter.resetTime = time.Unix(val, 0)
		}
	}
}

func (c *Client) GetRateLimitInfo() *models.RateLimit {
	return &models.RateLimit{
		Limit:     c.rateLimiter.limit,
		Remaining: c.rateLimiter.remaining,
		Reset:     c.rateLimiter.reset,
		ResetTime: c.rateLimiter.resetTime.Format(time.RFC3339),
	}
}

func (c *Client) ClearCache() {
	if c.cache != nil {
		c.cache.Clear()
		log.Printf("🧹 Discord client cache temizlendi")
	}
}

func (c *Client) GetCacheStats() *cache.CacheStats {
	if c.cache != nil {
		return c.cache.GetStats()
	}
	return &cache.CacheStats{}
} 