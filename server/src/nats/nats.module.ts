import { Module } from '@nestjs/common';
import { EventsModule } from '../events/events.module';
import { NatsConsumer } from './nats.consumer';

@Module({
  imports: [EventsModule],
  providers: [NatsConsumer],
})
export class NatsModule {}
