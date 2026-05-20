import { Injectable } from '@nestjs/common';
import { PrismaService } from '../prisma/prisma.service';

@Injectable()
export class PlayersService {
  constructor(private readonly prisma: PrismaService) {}

  async getPlayers(serverId: string, limit: number, offset: number, search?: string) {
    const searchClause = search ? `AND p.name ILIKE '%${search.replace(/'/g, "''")}%'` : '';

    const rows = await this.prisma.$queryRawUnsafe(`
      SELECT p.*,
        EXISTS (
          SELECT 1 FROM player_sessions ps
          WHERE ps.server_id = p.server_id
            AND ps.player_name = p.name
            AND ps.left_at IS NULL
        ) as is_online
      FROM players p
      WHERE p.server_id = $1 ${searchClause}
      ORDER BY p.last_seen DESC
      LIMIT $2 OFFSET $3
    `, serverId, limit, offset);

    const totalRows = await this.prisma.$queryRawUnsafe(
      `SELECT COUNT(*)::int as total FROM players WHERE server_id = $1 ${searchClause}`,
      serverId,
    );

    return {
      players: rows,
      total: (totalRows as any[])[0]?.total || 0,
    };
  }
}
