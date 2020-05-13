CREATE TABLE IF NOT EXISTS reports
(
    id                BIGSERIAL                NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL,

    sequence_type     TEXT                     NOT NULL,
    start_sequence_id DOUBLE PRECISION         NOT NULL,
    end_sequence_id   DOUBLE PRECISION         NOT NULL,

    success_count     NUMERIC,
    error_count       NUMERIC,
    error_msg         TEXT,
    duration          NUMERIC,
    details           JSONB,
    completed_at      TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY (id)
);

-- Hypertable

-- Indexes
CREATE index idx_report_sequence_type on reports (sequence_type);