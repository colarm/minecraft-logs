/// <reference types="node" />
import { defineConfig } from '@prisma/config';

// Prisma CLI loads .env automatically before evaluating the config file,
// but only after require.resolve resolves the module.
// We use process.env here so it gracefully falls through when the
// variable is unset (e.g. during prisma generate).
export default defineConfig({
  datasource: {
    url: process.env.DATABASE_URL ?? '',
  },
  schema: './prisma/schema.prisma',
});
