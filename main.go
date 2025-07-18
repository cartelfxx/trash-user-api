package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"discord-user-api/cache"
	"discord-user-api/config"
	"discord-user-api/discord"
	"discord-user-api/server"
)

func main() {
	log.Printf("🚀 Discord API Server başlatılıyor...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Konfigürasyon yüklenemedi: %v", err)
	}

	cache := cache.NewCache(
		cfg.Cache.MaxSize,
		cfg.Cache.TTL,
		cfg.Cache.CleanupInterval,
	)

	discordClient := discord.NewClient(cfg, cache)

	server := server.NewServer(cfg, discordClient, cache)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("❌ Server başlatılamadı: %v", err)
		}
	}()

	log.Printf("✅ Server başarıyla başlatıldı: http://%s:%s", cfg.Server.Host, cfg.Server.Port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Printf("🛑 Server kapatılıyor...")

	cache.Stop()
	log.Printf("👋 Server kapatıldı")
}
