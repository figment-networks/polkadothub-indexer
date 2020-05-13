CREATE TABLE IF NOT EXISTS syncables
(
    id            BIGSERIAL                NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL,

    sequence_type TEXT                     NOT NULL,
    sequence_id   DOUBLE PRECISION         NOT NULL,

    report_id     NUMERIC,
    type          VARCHAR(100),
    data          JSONB,
    processed_at  TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY (id)
);

-- Hypertable

-- Indexes
CREATE index idx_syncables_sequence_type on syncables (sequence_type);
CREATE index idx_syncables_sequence_id on syncables (sequence_id);
CREATE index idx_syncables_report_id on syncables (report_id);
CREATE index idx_syncables_processed_at on syncables (processed_at);