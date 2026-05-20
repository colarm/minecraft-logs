import { Controller, Get, Query } from '@nestjs/common';
import { PlayersService } from './players.service';

@Controller('players')
export class PlayersController {
  constructor(private readonly playersService: PlayersService) {}

  @Get()
  async getPlayers(
    @Query('server_id') serverId: string,
    @Query('limit') limit = '50',
    @Query('offset') offset = '0',
    @Query('search') search?: string,
  ) {
    return this.playersService.getPlayers(serverId, parseInt(limit), parseInt(offset), search);
  }
}
