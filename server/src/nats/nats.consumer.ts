import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { connect, JSONCodec } from 'nats';
import { EventsService } from '../events/events.service';
import { MinecraftEvent } from '../events/events.types';

@Injectable()
export class NatsConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(NatsConsumer.name);

  constructor(private readonly eventsService: EventsService) {}

  async onModuleInit() {
    const url = process.env.NATS_URL || 'nats://localhost:4222';
    const token = process.env.NATS_TOKEN || '';

    this.logger.log(`Connecting to NATS at ${url}`);

    const nc = await connect({
      servers: url,
      token: token || undefined,
      reconnectTimeWait: 5000,
      maxReconnectAttempts: -1,
    });

    this.logger.log(`Connected to NATS at ${nc.getServer()}`);

    const jc = JSONCodec<MinecraftEvent>();
    const sub = nc.subscribe('minecraft.events.>', { queue: 'mc-logs-worker' });

    (async () => {
      for await (const msg of sub) {
        try {
          const event = jc.decode(msg.data);
          await this.eventsService.ingest(event);
        } catch (err) {
          this.logger.error(`Failed to process message: ${err}`);
        }
      }
    })();
  }

  async onModuleDestroy() {
    this.logger.log('Shutting down NATS consumer');
  }
}
