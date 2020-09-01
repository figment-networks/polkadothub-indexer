CREATE TABLE IF NOT EXISTS account_era_sequences
(
    id                           BIGSERIAL                NOT NULL,

    era                          DECIMAL(65, 0)           NOT NULL,
    start_height                 DECIMAL(65, 0)           NOT NULL,
    end_height                   DECIMAL(65, 0)           NOT NULL,
    time                         TIMESTAMP WITH TIME ZONE NOT NULL,

    stash_account                TEXT                     NOT NULL,
    controller_account           TEXT                     NOT NULL,
    validator_stash_account      TEXT                     NOT NULL,
    validator_controller_account TEXT                     NOT NULL,
    stake                        DECIMAL(65, 0)           NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_account_era_sequences_era on account_era_sequences (era);
CREATE index idx_account_era_sequences_heights on account_era_sequences (start_height, end_height);
CREATE index idx_account_era_sequences_time on account_era_sequences (time);
CREATE index idx_account_era_sequences_stash_account on account_era_sequences (stash_account);
CREATE index idx_account_era_sequences_validator_stash_account on account_era_sequences (validator_stash_account);
