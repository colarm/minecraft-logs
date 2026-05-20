export interface MinecraftEvent {
  server_id: string;
  event_type: string;
  player_name?: string;
  message?: string;
  metadata?: Record<string, unknown>;
  timestamp: string;
}
