import { Injectable, Logger } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';
import { MinecraftEvent } from './events.types';

@Injectable()
export class EventsService {
  private readonly logger = new Logger(EventsService.name);

  constructor(private readonly prisma: PrismaService) {}

  async ingest(event: MinecraftEvent): Promise<void> {
    try {
      await this.prisma.$executeRaw`
        INSERT INTO events (server_id, event_type, player_name, message, metadata, "timestamp")
        VALUES (${event.server_id}, ${event.event_type}, ${event.player_name || null},
                ${event.message || null}, ${JSON.stringify(event.metadata || {})}, ${event.timestamp}::timestamptz)
      `;

      switch (event.event_type) {
        case 'player_join':
          await this.handlePlayerJoin(event);
          break;
        case 'player_leave':
          await this.handlePlayerLeave(event);
          break;
        case 'tps':
          await this.handleTps(event);
          break;
      }
    } catch (err) {
      this.logger.error(`Failed to ingest event: ${JSON.stringify(event)}`, err);
    }
  }

  private async handlePlayerJoin(event: MinecraftEvent): Promise<void> {
    const ts = new Date(event.timestamp);
    await this.prisma.$executeRaw`
      INSERT INTO players (server_id, name, first_seen, last_seen, join_count)
      VALUES (${event.server_id}, ${event.player_name}, ${ts}, ${ts}, 1)
      ON CONFLICT (server_id, name) DO UPDATE
        SET last_seen = ${ts}, join_count = players.join_count + 1
    `;

    await this.prisma.$executeRaw`
      INSERT INTO player_sessions (server_id, player_name, joined_at)
      VALUES (${event.server_id}, ${event.player_name}, ${ts})
    `;

    await this.prisma.$executeRaw`
      INSERT INTO server_status (server_id, online_count, last_updated)
      VALUES (${event.server_id}, 1, ${ts})
      ON CONFLICT (server_id) DO UPDATE
        SET online_count = server_status.online_count + 1, last_updated = ${ts}
    `;

    await this.prisma.$executeRaw`
      UPDATE server_status SET peak_online = GREATEST(peak_online, online_count)
      WHERE server_id = ${event.server_id}
    `;
  }

  private async handlePlayerLeave(event: MinecraftEvent): Promise<void> {
    const ts = new Date(event.timestamp);
    await this.prisma.$executeRaw`
      UPDATE player_sessions
      SET left_at = ${ts}, duration = ${ts} - joined_at
      WHERE server_id = ${event.server_id}
        AND player_name = ${event.player_name}
        AND left_at IS NULL
      ORDER BY joined_at DESC
      LIMIT 1
    `;

    await this.prisma.$executeRaw`
      UPDATE players
      SET total_play_time = total_play_time + COALESCE(
        (SELECT EXTRACT(EPOCH FROM (${ts} - joined_at))::BIGINT
         FROM player_sessions
         WHERE server_id = ${event.server_id}
           AND player_name = ${event.player_name}
           AND left_at = ${ts}
         ORDER BY joined_at DESC LIMIT 1), 0)
      WHERE server_id = ${event.server_id} AND name = ${event.player_name}
    `;

    await this.prisma.$executeRaw`
      UPDATE server_status
      SET online_count = GREATEST(0, online_count - 1), last_updated = ${ts}
      WHERE server_id = ${event.server_id}
    `;
  }

  private async handleTps(event: MinecraftEvent): Promise<void> {
    const tps = (event.metadata?.tps as number) || null;
    const ts = new Date(event.timestamp);
    await this.prisma.$executeRaw`
      UPDATE server_status
      SET current_tps = ${tps}, last_updated = ${ts}
      WHERE server_id = ${event.server_id}
    `;
  }
}
