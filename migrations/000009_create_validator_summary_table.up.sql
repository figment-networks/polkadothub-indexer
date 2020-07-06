CREATE TABLE IF NOT EXISTS validator_summary
(
    id                BIGSERIAL                NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL,

    time_interval     VARCHAR                  NOT NULL,
    time_bucket       TIMESTAMP WITH TIME ZONE NOT NULL,
    index_version     INT                      NOT NULL,

    address        TEXT                     NOT NULL,
    voting_power_avg  DECIMAL(65, 0)           NOT NULL,
    voting_power_max  DECIMAL(65, 0)           NOT NULL,
    voting_power_min  DECIMAL(65, 0)           NOT NULL,
    total_shares_avg  DECIMAL(65, 0)           NOT NULL,
    total_shares_max  DECIMAL(65, 0)           NOT NULL,
    total_shares_min  DECIMAL(65, 0)           NOT NULL,
    uptime_avg        DECIMAL                  NOT NULL,
    validated_sum     BIGINT                   NOT NULL,
    not_validated_sum BIGINT                   NOT NULL,
    proposed_sum      BIGINT                   NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_validator_summary_time on validator_summary (time_interval, time_bucket);
CREATE index idx_validator_summary_index_version on validator_summary (index_version);
CREATE index idx_validator_summary_address on validator_summary (address);