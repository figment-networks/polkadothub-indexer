CREATE TABLE IF NOT EXISTS reports
(
    id            BIGSERIAL                NOT NULL,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL,

    kind          INT                      NOT NULL,
    index_version INT                      NOT NULL,
    start_height  DECIMAL(65, 0)           NOT NULL,
    end_height    DECIMAL(65, 0)           NOT NULL,
    success_count INT,
    error_count   INT,
    error_msg     TEXT,
    duration      BIGINT,
    completed_at  TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_reports_kind on reports (kind);
CREATE index idx_reports_index_version on reports (index_version);