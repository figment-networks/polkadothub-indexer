CREATE TABLE IF NOT EXISTS event_sequences
(
    id              BIGSERIAL                NOT NULL,

    height          DECIMAL(65, 0)           NOT NULL,
    time            TIMESTAMP WITH TIME ZONE NOT NULL,

    index           BIGINT                   NOT NULL,
    extrinsic_index BIGINT                   NOT NULL,
    data            JSONB                    NOT NULL,
    phase           TEXT                     NOT NULL,
    method          TEXT                     NOT NULL,
    section         TEXT                     NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_event_sequences_height on event_sequences (height);
CREATE index idx_event_sequences_type on event_sequences (method, section);
