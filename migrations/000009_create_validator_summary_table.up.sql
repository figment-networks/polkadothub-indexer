CREATE TABLE IF NOT EXISTS validator_summary
(
    id                BIGSERIAL                NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL,

    time_interval     VARCHAR                  NOT NULL,
    time_bucket       TIMESTAMP WITH TIME ZONE NOT NULL,
    index_version     INT                      NOT NULL,

    stash_account     TEXT                     NOT NULL,
    total_stake_avg   DECIMAL(65, 0)           NOT NULL,
    total_stake_max   DECIMAL(65, 0)           NOT NULL,
    total_stake_min   DECIMAL(65, 0)           NOT NULL,
    own_stake_avg     DECIMAL(65, 0)           NOT NULL,
    own_stake_max     DECIMAL(65, 0)           NOT NULL,
    own_stake_min     DECIMAL(65, 0)           NOT NULL,
    stakers_stake_avg DECIMAL(65, 0)           NOT NULL,
    stakers_stake_max DECIMAL(65, 0)           NOT NULL,
    stakers_stake_min DECIMAL(65, 0)           NOT NULL,
    reward_points_avg  DECIMAL(65, 0)           NOT NULL,
    reward_points_max  DECIMAL(65, 0)           NOT NULL,
    reward_points_min  DECIMAL(65, 0)           NOT NULL,
    commission_avg    DECIMAL                  NOT NULL,
    commission_min    BIGINT                   NOT NULL,
    commission_max    BIGINT                   NOT NULL,
    stakers_count_avg DECIMAL                  NOT NULL,
    stakers_count_min BIGINT                   NOT NULL,
    stakers_count_max BIGINT                   NOT NULL,

    uptime_avg        DECIMAL                  NOT NULL,
    uptime_min        BIGINT                   NOT NULL,
    uptime_max        BIGINT                   NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_validator_summary_time on validator_summary (time_interval, time_bucket);
CREATE index idx_validator_summary_index_version on validator_summary (index_version);
CREATE index idx_validator_summary_stash_account on validator_summary (stash_account);