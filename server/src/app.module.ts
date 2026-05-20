import { Module } from '@nestjs/common';
import { NatsModule } from './nats/nats.module';
import { EventsModule } from './events/events.module';
import { StatsModule } from './stats/stats.module';
import { PlayersModule } from './players/players.module';
import { LogsModule } from './logs/logs.module';
import { HistoryModule } from './history/history.module';
import { PrismaModule } from './prisma/prisma.module';

@Module({
  imports: [
    PrismaModule,
    NatsModule,
    EventsModule,
    StatsModule,
    PlayersModule,
    LogsModule,
    HistoryModule,
  ],
})
export class AppModule {}
