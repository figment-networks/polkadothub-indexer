CREATE TABLE IF NOT EXISTS validator_era_sequences
(
    id                 BIGSERIAL        NOT NULL,

    era                DOUBLE PRECISION NOT NULL,
    start_height       DOUBLE PRECISION NOT NULL,
    end_height         DOUBLE PRECISION NOT NULL,

    index              BIGINT           NOT NULL,
    stash_account      TEXT             NOT NULL,
    controller_account TEXT             NOT NULL,
    session_accounts   VARCHAR[],
    total_stake        DECIMAL(65, 0)   NOT NULL,
    own_stake          DECIMAL(65, 0)   NOT NULL,
    stakers_stake      DECIMAL(65, 0)   NOT NULL,
    reward_points      DECIMAL(65, 0)   NOT NULL,
    commission         BIGINT           NOT NULL,
    stakers_count      INT              NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_validator_era_sequences_era on validator_era_sequences (era);
CREATE index idx_validator_era_sequences_heights on validator_era_sequences (start_height, end_height);
CREATE index idx_validator_era_sequences_validator_stash_account on validator_era_sequences (stash_account);
