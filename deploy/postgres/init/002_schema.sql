-- ============================================
-- events table (TimescaleDB hypertable)
-- ============================================
CREATE TABLE events (
    id              BIGSERIAL,
    server_id       VARCHAR(64) NOT NULL,
    event_type      VARCHAR(32) NOT NULL,
    player_name     VARCHAR(64),
    message         TEXT,
    metadata        JSONB DEFAULT '{}',
    "timestamp"     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

SELECT create_hypertable('events', 'timestamp', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX idx_events_server_id ON events (server_id);
CREATE INDEX idx_events_event_type ON events (event_type);
CREATE INDEX idx_events_player_name ON events (player_name);
CREATE INDEX idx_events_timestamp_desc ON events ("timestamp" DESC);

SELECT add_compression_policy('events', INTERVAL '30 days');
SELECT add_retention_policy('events', INTERVAL '365 days');

-- ============================================
-- players table
-- ============================================
CREATE TABLE players (
    id              SERIAL PRIMARY KEY,
    uuid            UUID NOT NULL DEFAULT gen_random_uuid(),
    server_id       VARCHAR(64) NOT NULL,
    name            VARCHAR(64) NOT NULL,
    first_seen      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_play_time BIGINT NOT NULL DEFAULT 0,
    join_count      INTEGER NOT NULL DEFAULT 0,
    UNIQUE(server_id, name)
);

CREATE INDEX idx_players_server_id ON players (server_id);
CREATE INDEX idx_players_name ON players (name);

-- ============================================
-- server_status table
-- ============================================
CREATE TABLE server_status (
    server_id       VARCHAR(64) PRIMARY KEY,
    current_tps     DOUBLE PRECISION,
    online_count    INTEGER NOT NULL DEFAULT 0,
    peak_online     INTEGER NOT NULL DEFAULT 0,
    last_updated    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================
-- player_sessions table
-- ============================================
CREATE TABLE player_sessions (
    id              BIGSERIAL,
    server_id       VARCHAR(64) NOT NULL,
    player_name     VARCHAR(64) NOT NULL,
    joined_at       TIMESTAMPTZ NOT NULL,
    left_at         TIMESTAMPTZ,
    duration        INTERVAL
);

CREATE INDEX idx_sessions_server_player ON player_sessions (server_id, player_name);
CREATE INDEX idx_sessions_joined_at ON player_sessions (joined_at);

-- ============================================
-- Continuous aggregates
-- ============================================
CREATE MATERIALIZED VIEW tps_per_minute
WITH (timescaledb.continuous) AS
SELECT
    server_id,
    time_bucket('1 minute', "timestamp") AS bucket,
    AVG((metadata->>'tps')::DOUBLE PRECISION) AS avg_tps,
    MIN((metadata->>'tps')::DOUBLE PRECISION) AS min_tps,
    MAX((metadata->>'tps')::DOUBLE PRECISION) AS max_tps
FROM events
WHERE event_type = 'tps'
GROUP BY server_id, bucket;

CREATE MATERIALIZED VIEW player_activity_per_hour
WITH (timescaledb.continuous) AS
SELECT
    server_id,
    time_bucket('1 hour', "timestamp") AS bucket,
    COUNT(*) FILTER (WHERE event_type = 'player_join') AS joins,
    COUNT(*) FILTER (WHERE event_type = 'player_leave') AS leaves,
    COUNT(*) FILTER (WHERE event_type = 'chat') AS chat_messages,
    COUNT(*) FILTER (WHERE event_type = 'death') AS deaths
FROM events
GROUP BY server_id, bucket;

SELECT add_continuous_aggregate_policy('tps_per_minute',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '5 minutes');

SELECT add_continuous_aggregate_policy('player_activity_per_hour',
    start_offset => INTERVAL '2 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '10 minutes');
