package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"discord-user-api/cache"
	"discord-user-api/config"
	"discord-user-api/discord"
	"discord-user-api/middleware"
	"discord-user-api/models"
	"discord-user-api/websocket"
)

type Server struct {
	config     *config.Config
	discord    *discord.Client
	cache      *cache.Cache
	rateLimiter *middleware.RateLimiter
	wsManager  *websocket.WebSocketManager
}

func NewServer(cfg *config.Config, discordClient *discord.Client, cache *cache.Cache) *Server {
	rateLimiter := middleware.CreateRateLimiter(cfg)
	wsManager := websocket.NewWebSocketManager()
	
	server := &Server{
		config:      cfg,
		discord:     discordClient,
		cache:       cache,
		rateLimiter: rateLimiter,
		wsManager:   wsManager,
	}

	cache.SetWebSocketManager(wsManager)
	
	go wsManager.Start()
	
	cache.StartAutoRefresh()

	log.Printf("🚀 HTTP Server başlatıldı")
	return server
}

func (s *Server) Start() error {
	middlewareChain := middleware.Compose(
		middleware.Recovery,
		middleware.Security,
		middleware.RequestID,
		middleware.Logging,
		middleware.CORS,
	)

	if s.rateLimiter != nil {
		middlewareChain = middleware.Compose(
			middleware.Recovery,
			middleware.Security,
			middleware.RequestID,
			middleware.Logging,
			middleware.CORS,
			middleware.RateLimit(s.rateLimiter),
		)
	}

	http.HandleFunc("/", middlewareChain(s.handleRoot))
	http.HandleFunc("/guilds", middlewareChain(s.handleGuilds))
	http.HandleFunc("/users", middlewareChain(s.handleUsers))
	http.HandleFunc("/guilds/", middlewareChain(s.handleGuildByID))
	http.HandleFunc("/guilds/members", middlewareChain(s.handleGuildMembers))
	http.HandleFunc("/guilds/refresh", middlewareChain(s.handleGuildRefresh))
	http.HandleFunc("/guilds/members/refresh", middlewareChain(s.handleGuildMembersRefresh))
	http.HandleFunc("/health", middlewareChain(s.handleHealth))
	http.HandleFunc("/stats", middlewareChain(s.handleStats))
	http.HandleFunc("/cache/clear", middlewareChain(s.handleCacheClear))
	http.HandleFunc("/cache/stats", middlewareChain(s.handleCacheStats))
	http.HandleFunc("/websocket", s.wsManager.HandleWebSocket)
	http.HandleFunc("/websocket/stats", middlewareChain(s.handleWebSocketStats))

	server := &http.Server{
		Addr:         s.config.Server.Host + ":" + s.config.Server.Port,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	log.Printf("🌐 Server başlatılıyor: http://%s:%s", s.config.Server.Host, s.config.Server.Port)
	log.Printf("📋 Endpoints:")
	log.Printf("   GET  /                    - API bilgileri")
	log.Printf("   GET  /guilds              - Tüm guild'ler")
	log.Printf("   GET  /guilds?id=<id>      - Belirli guild")
	log.Printf("   GET  /users?id=<id>       - Kullanıcı profili")
	log.Printf("   GET  /guilds/members?guild_id=<id>&limit=<limit> - Guild üyeleri")
	log.Printf("   POST /guilds/refresh?guild_id=<id> - Guild'i yenile")
	log.Printf("   POST /guilds/members/refresh?guild_id=<id>&limit=<limit> - Üyeleri yenile")
	log.Printf("   GET  /health              - Sağlık kontrolü")
	log.Printf("   GET  /stats               - İstatistikler")
	log.Printf("   POST /cache/clear         - Cache temizle")
	log.Printf("   GET  /cache/stats         - Cache istatistikleri")
	log.Printf("   WS   /websocket           - WebSocket bağlantısı")
	log.Printf("   GET  /websocket/stats     - WebSocket istatistikleri")

	return server.ListenAndServe()
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"name":        "Discord API Server - Professional Edition",
			"version":     "2.0.0",
			"description": "Profesyonel Discord API sunucusu - Gerçek zamanlı güncellemeler ile",
			"features": []string{
				"🔐 Güvenli API",
				"💾 Akıllı Cache",
				"🔄 Retry Logic",
				"📊 İstatistikler",
				"🛡️ Middleware",
				"⚡ Performans",
				"🎯 Türkçe Loglar",
				"🔌 WebSocket Desteği",
				"🔄 Otomatik Yenileme",
			},
			"endpoints": map[string]string{
				"guilds":         "/guilds",
				"guilds_filter":  "/guilds?id=<guild_id>",
				"users":          "/users?id=<user_id>",
				"guild_members":  "/guilds/members?guild_id=<guild_id>&limit=<limit>",
				"guild_refresh":  "/guilds/refresh?guild_id=<guild_id>",
				"members_refresh": "/guilds/members/refresh?guild_id=<guild_id>&limit=<limit>",
				"health":         "/health",
				"stats":          "/stats",
				"cache_clear":    "/cache/clear",
				"cache_stats":    "/cache/stats",
				"websocket":      "/websocket",
				"websocket_stats": "/websocket/stats",
			},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleGuilds(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guildID := r.URL.Query().Get("id")

	if guildID != "" {
		guild, err := s.discord.GetGuild(guildID)
		if err != nil {
			log.Printf("❌ Guild getirme hatası: %v", err)
			s.sendError(w, fmt.Sprintf("Guild bulunamadı: %v", err), http.StatusNotFound)
			return
		}

		response := models.APIResponse{
			Success:   true,
			Data:      guild,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		s.sendJSONResponse(w, response, http.StatusOK)
		return
	}

	guilds, err := s.discord.GetGuilds()
	if err != nil {
		log.Printf("❌ Guild'ler getirme hatası: %v", err)
		s.sendError(w, fmt.Sprintf("Guild'ler getirilemedi: %v", err), http.StatusInternalServerError)
		return
	}

	response := models.APIResponse{
		Success:   true,
		Data:      guilds,
		Count:     len(guilds),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RateLimit: s.discord.GetRateLimitInfo(),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		s.sendError(w, "User ID required (id parameter)", http.StatusBadRequest)
		return
	}

	if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
		s.sendError(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	profile, err := s.discord.GetUser(userID)
	if err != nil {
		log.Printf("❌ Kullanıcı getirme hatası: %v", err)
		s.sendError(w, fmt.Sprintf("Kullanıcı bulunamadı: %v", err), http.StatusNotFound)
		return
	}

	response := models.APIResponse{
		Success:   true,
		Data:      profile,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RateLimit: s.discord.GetRateLimitInfo(),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleGuildByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	guildID := path[len("/guilds/"):]

	if guildID == "" {
		s.sendError(w, "Guild ID required", http.StatusBadRequest)
		return
	}

	guild, err := s.discord.GetGuild(guildID)
	if err != nil {
		log.Printf("❌ Guild getirme hatası: %v", err)
		s.sendError(w, fmt.Sprintf("Guild bulunamadı: %v", err), http.StatusNotFound)
		return
	}

	response := models.APIResponse{
		Success:   true,
		Data:      guild,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RateLimit: s.discord.GetRateLimitInfo(),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleGuildMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		s.sendError(w, "Guild ID required (guild_id parameter)", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 1000
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	members, err := s.discord.GetGuildMembers(guildID, limit)
	if err != nil {
		log.Printf("❌ Guild üyeleri getirme hatası: %v", err)
		s.sendError(w, fmt.Sprintf("Guild üyeleri getirilemedi: %v", err), http.StatusNotFound)
		return
	}

	response := models.APIResponse{
		Success:   true,
		Data:      members,
		Count:     len(members),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RateLimit: s.discord.GetRateLimitInfo(),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleGuildRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		s.sendError(w, "Guild ID required (guild_id parameter)", http.StatusBadRequest)
		return
	}

	err := s.discord.RefreshGuild(guildID)
	if err != nil {
		log.Printf("❌ Guild yenileme hatası: %v", err)
		s.sendError(w, fmt.Sprintf("Guild yenilenemedi: %v", err), http.StatusInternalServerError)
		return
	}

	s.wsManager.BroadcastToGuild(guildID, "guild_refreshed", map[string]interface{}{
		"guild_id": guildID,
		"message":  "Guild başarıyla yenilendi",
	})

	response := models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Guild başarıyla yenilendi: %s", guildID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleGuildMembersRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	guildID := r.URL.Query().Get("guild_id")
	if guildID == "" {
		s.sendError(w, "Guild ID required (guild_id parameter)", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 1000
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	err := s.discord.RefreshGuildMembers(guildID, limit)
	if err != nil {
		log.Printf("❌ Guild üyeleri yenileme hatası: %v", err)
		s.sendError(w, fmt.Sprintf("Guild üyeleri yenilenemedi: %v", err), http.StatusInternalServerError)
		return
	}

	s.wsManager.BroadcastToGuild(guildID, "members_refreshed", map[string]interface{}{
		"guild_id": guildID,
		"limit":    limit,
		"message":  "Guild üyeleri başarıyla yenilendi",
	})

	response := models.APIResponse{
		Success:   true,
		Message:   fmt.Sprintf("Guild üyeleri başarıyla yenilendi: %s", guildID),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"uptime":    time.Since(time.Now()).String(),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cacheStats := s.discord.GetCacheStats()
	rateLimitInfo := s.discord.GetRateLimitInfo()
	wsStats := s.wsManager.GetConnectedClientsInfo()

	response := models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"cache": map[string]interface{}{
				"hits":        cacheStats.Hits,
				"misses":      cacheStats.Misses,
				"evictions":   cacheStats.Evictions,
				"refreshes":   cacheStats.Refreshes,
				"size":        cacheStats.Size,
				"last_cleanup": cacheStats.LastCleanup.Format(time.RFC3339),
			},
			"rate_limit": rateLimitInfo,
			"websocket": map[string]interface{}{
				"connected_clients": s.wsManager.GetConnectedClientsCount(),
				"clients_info":      wsStats,
			},
			"server": map[string]interface{}{
				"start_time": time.Now().UTC().Format(time.RFC3339),
			},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleCacheClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.discord.ClearCache()

	response := models.APIResponse{
		Success:   true,
		Message:   "Cache başarıyla temizlendi",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cacheStats := s.discord.GetCacheStats()

	response := models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"hits":         cacheStats.Hits,
			"misses":       cacheStats.Misses,
			"evictions":    cacheStats.Evictions,
			"refreshes":    cacheStats.Refreshes,
			"size":         cacheStats.Size,
			"last_cleanup": cacheStats.LastCleanup.Format(time.RFC3339),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) handleWebSocketStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		s.sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"connected_clients": s.wsManager.GetConnectedClientsCount(),
			"clients_info":      s.wsManager.GetConnectedClientsInfo(),
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, http.StatusOK)
}

func (s *Server) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	response := models.APIResponse{
		Success:   false,
		Error:     message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	s.sendJSONResponse(w, response, statusCode)
} 