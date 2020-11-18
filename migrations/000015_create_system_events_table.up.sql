CREATE TABLE IF NOT EXISTS system_events
(
    id         BIGSERIAL                NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    height     DECIMAL(65, 0)           NOT NULL,
    time       TIMESTAMP WITH TIME ZONE NOT NULL,

    actor      TEXT,
    kind       TEXT                     NOT NULL,
    data       JSONB                    NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_system_events_height on system_events (height);
CREATE index idx_system_events_actor on system_events (actor);
CREATE index idx_system_events_kind on system_events (kind);