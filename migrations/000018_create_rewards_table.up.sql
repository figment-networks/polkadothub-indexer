CREATE TABLE IF NOT EXISTS rewards
(
    id                           BIGSERIAL                NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    era                          DECIMAL(65, 0)           NOT NULL,

    stash_account                TEXT                     NOT NULL,
    validator_stash_account      TEXT                     NOT NULL,
    amount                       DECIMAL(65, 0)           NOT NULL,
    kind                         TEXT                     NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_rewards_era on rewards (era);
CREATE index idx_rewards_stash_account on rewards (stash_account);
CREATE UNIQUE INDEX idx_rewards_accounts_kind
    ON rewards(era, stash_account, validator_stash_account, kind);