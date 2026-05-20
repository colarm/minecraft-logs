import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';

@Injectable()
export class HistoryService {
  constructor(private readonly prisma: PrismaService) {}

  async getHistory(serverId: string, start: string, end: string, interval: string) {
    const validIntervals = ['1m', '5m', '15m', '1h', '6h', '1d'];
    if (!validIntervals.includes(interval)) {
      interval = '1h';
    }

    const rows = await this.prisma.$queryRawUnsafe(
      `SELECT
        time_bucket($2, "timestamp") AS bucket,
        AVG((metadata->>'tps')::DOUBLE PRECISION) AS avg_tps,
        COUNT(*) FILTER (WHERE event_type = 'player_join') AS joins,
        COUNT(*) FILTER (WHERE event_type = 'player_leave') AS leaves,
        COUNT(*) FILTER (WHERE event_type = 'chat') AS chat_messages,
        COUNT(*) FILTER (WHERE event_type = 'death') AS deaths,
        COUNT(*) AS total_events
       FROM events
       WHERE server_id = $1
         AND "timestamp" >= $3::timestamptz
         AND "timestamp" < $4::timestamptz
       GROUP BY bucket
       ORDER BY bucket ASC`,
      serverId,
      interval,
      start,
      end,
    );

    return { buckets: rows };
  }
}
