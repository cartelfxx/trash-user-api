import DiscordApiClient from './index.js';
import { logger } from './utils/logger.js';
import { config } from './config/index.js';

async function main() {
  logger.info('🚀 Discord API İstemci Örneği Başlatılıyor...', '🎯');


  const serverConfig = config.get();
  logger.info(`🔧 Sunucu URL'si: ${config.getServerUrl()}`, '⚙️');
  logger.info(`🔧 WebSocket URL'si: ${config.getWebSocketUrl()}`, '⚙️');


  const client = new DiscordApiClient({
    guildId: process.env['DISCORD_GUILD_ID'] || '1234567890123456789',
    userId: process.env['DISCORD_USER_ID'] || '1234567890123456789',
    autoReconnect: true,
    pingInterval: serverConfig.websocket.pingInterval,
    pongTimeout: serverConfig.websocket.pongTimeout,
  });


  client.on('connected', () => {
    logger.info('✅ WebSocket\'e başarıyla bağlanıldı', '🔗');
  });

  client.on('disconnected', () => {
    logger.info('❌ WebSocket bağlantısı kesildi', '🔌');
  });

  client.on('error', (error) => {
    logger.error(`❌ İstemci hatası: ${error.message}`, '💥');
  });

  client.on('websocket_event', (event) => {
    logger.info(`📡 WebSocket olayı: ${event.type}`, '📨');
  });

  client.on('cache_update', (_data) => {
    logger.info('💾 Önbellek güncellendi', '🔄');
  });

  client.on('guild_refreshed', (_data) => {
    logger.info('🔄 Sunucu güncellendi', '🏰');
  });

  client.on('members_refreshed', (_data) => {
    logger.info('👥 Üyeler güncellendi', '👤');
  });

  try {
    await client.connect();
    logger.info('✅ Başarıyla bağlanıldı', '🎉');

    const guildId = process.env['DISCORD_GUILD_ID'] || '1234567890123456789';
    client.subscribe(guildId);
    logger.info(`📡 Sunucuya abone olundu: ${guildId}`, '📋');

    const apiClient = client.getApiClient();

    logger.info('📊 Sunucu bilgisi alınıyor...', '🔍');
    const serverInfo = await apiClient.getServerInfo();
    logger.info(`📊 Sunucu bilgisi: ${JSON.stringify(serverInfo, null, 2)}`, '📈');

    logger.info('📈 Sunucu istatistikleri alınıyor...', '📊');
    const stats = await apiClient.getStats();
    logger.info(`📈 Sunucu istatistikleri: ${JSON.stringify(stats, null, 2)}`, '📊');

    logger.info('💾 Önbellek istatistikleri alınıyor...', '🗄️');
    const cacheStats = await apiClient.getCacheStats();
    logger.info(`💾 Önbellek istatistikleri: ${JSON.stringify(cacheStats, null, 2)}`, '🗄️');

    logger.info('🔌 WebSocket istatistikleri alınıyor...', '📡');
    const wsStats = await apiClient.getWebSocketStats();
    logger.info(`🔌 WebSocket istatistikleri: ${JSON.stringify(wsStats, null, 2)}`, '📡');

    logger.info('⏰ Bağlantı açık tutuluyor... Çıkmak için Ctrl+C', '⏳');
    
    process.on('SIGINT', () => {
      logger.info('🛑 Güvenli şekilde kapatılıyor...', '🔄');
      client.disconnect();
      process.exit(0);
    });

    process.on('SIGTERM', () => {
      logger.info('🛑 SIGTERM alındı, kapatılıyor...', '🔄');
      client.disconnect();
      process.exit(0);
    });

  } catch (error) {
    logger.error(`❌ İstemci başlatılamadı: ${error}`, '💥');
    process.exit(1);
  }
}


main().catch((error) => {
  logger.error(`❌ Yakalanmamış hata: ${error}`, '💥');
  process.exit(1);
}); 