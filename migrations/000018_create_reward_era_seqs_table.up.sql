CREATE TABLE IF NOT EXISTS reward_era_sequences
(
    id                           BIGSERIAL                NOT NULL,

    era                DECIMAL(65, 0)           NOT NULL,
    start_height       DECIMAL(65, 0)           NOT NULL,
    end_height         DECIMAL(65, 0)           NOT NULL,
    time               TIMESTAMP WITH TIME ZONE NOT NULL,

    stash_account                TEXT                     NOT NULL,
    validator_stash_account      TEXT                     NOT NULL,
    amount                       DECIMAL(65, 0)           NOT NULL,
    kind                         TEXT                     NOT NULL,
    claimed                      BOOLEAN                  NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_rewards_era on reward_era_sequences (era);
CREATE index idx_rewards_stash_account on reward_era_sequences (stash_account);
CREATE UNIQUE INDEX idx_rewards_accounts_kind
    ON reward_era_sequences(era, stash_account, validator_stash_account, kind);