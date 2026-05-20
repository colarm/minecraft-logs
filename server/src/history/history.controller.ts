import { Controller, Get, Query } from '@nestjs/common';
import { HistoryService } from './history.service';

@Controller('history')
export class HistoryController {
  constructor(private readonly historyService: HistoryService) {}

  @Get()
  async getHistory(
    @Query('server_id') serverId: string,
    @Query('start') start: string,
    @Query('end') end: string,
    @Query('interval') interval = '1h',
  ) {
    return this.historyService.getHistory(serverId, start, end, interval);
  }
}
