import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';

@Injectable()
export class LogsService {
  constructor(private readonly prisma: PrismaService) {}

  async getLogs(
    serverId: string,
    limit: number,
    offset: number,
    eventType?: string,
    playerName?: string,
  ) {
    const conditions: string[] = ['e.server_id = $1'];
    const params: any[] = [serverId];
    let paramIdx = 2;

    if (eventType) {
      conditions.push(`e.event_type = $${paramIdx}`);
      params.push(eventType);
      paramIdx++;
    }
    if (playerName) {
      conditions.push(`e.player_name = $${paramIdx}`);
      params.push(playerName);
      paramIdx++;
    }

    const whereClause = conditions.join(' AND ');

    const rows = await this.prisma.$queryRawUnsafe(
      `SELECT id, event_type, player_name, message, metadata, "timestamp"
       FROM events e
       WHERE ${whereClause}
       ORDER BY e."timestamp" DESC
       LIMIT $${paramIdx} OFFSET $${paramIdx + 1}`,
      ...params,
      limit,
      offset,
    );

    const totalRows = await this.prisma.$queryRawUnsafe(
      `SELECT COUNT(*)::int as total FROM events e WHERE ${whereClause}`,
      ...params,
    );

    return {
      events: rows,
      total: (totalRows as any[])[0]?.total || 0,
      has_more: (rows as any[]).length >= limit,
    };
  }
}
