import { Controller, Get, Query } from '@nestjs/common';
import { LogsService } from './logs.service';

@Controller('logs')
export class LogsController {
  constructor(private readonly logsService: LogsService) {}

  @Get()
  async getLogs(
    @Query('server_id') serverId: string,
    @Query('limit') limit = '100',
    @Query('offset') offset = '0',
    @Query('event_type') eventType?: string,
    @Query('player_name') playerName?: string,
  ) {
    return this.logsService.getLogs(serverId, parseInt(limit), parseInt(offset), eventType, playerName);
  }
}
