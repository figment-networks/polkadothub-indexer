CREATE TABLE IF NOT EXISTS validator_aggregates
(
    id                         BIGSERIAL                NOT NULL,
    created_at                 TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                 TIMESTAMP WITH TIME ZONE NOT NULL,

    started_at_height          DECIMAL(65, 0)           NOT NULL,
    started_at                 TIMESTAMP WITH TIME ZONE NOT NULL,
    recent_at_height           DECIMAL(65, 0)           NOT NULL,
    recent_at                  TIMESTAMP WITH TIME ZONE NOT NULL,

    stash_account              TEXT,
    recent_as_validator_height DECIMAL(65, 0),
    accumulated_uptime         BIGINT,
    accumulated_uptime_count   BIGINT,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_validator_aggregates_stash_account on validator_aggregates (stash_account);