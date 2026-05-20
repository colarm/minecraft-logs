import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';

@Injectable()
export class StatsService {
  constructor(private readonly prisma: PrismaService) {}

  async getStats(serverId: string) {
    const [status, tpsHistory] = await Promise.all([
      this.prisma.$queryRaw`SELECT * FROM server_status WHERE server_id = ${serverId}`,
      this.prisma.$queryRaw`
        SELECT bucket, avg_tps, min_tps, max_tps
        FROM tps_per_minute
        WHERE server_id = ${serverId}
        AND bucket > NOW() - INTERVAL '24 hours'
        ORDER BY bucket ASC
      `,
    ]);

    const current = (status as any[])[0] || {
      server_id: serverId,
      current_tps: null,
      online_count: 0,
      peak_online: 0,
      last_updated: new Date(),
    };

    return {
      server_id: current.server_id,
      current_tps: current.current_tps,
      online_count: current.online_count,
      peak_online: current.peak_online,
      last_updated: current.last_updated,
      tps_history: tpsHistory,
    };
  }
}
