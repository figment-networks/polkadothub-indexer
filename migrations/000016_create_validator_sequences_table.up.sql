CREATE TABLE IF NOT EXISTS validator_sequences
(
    id                      BIGSERIAL                NOT NULL,

    height                  DECIMAL(65, 0)           NOT NULL,
    time                    TIMESTAMP WITH TIME ZONE NOT NULL,

    stash_account           TEXT                     NOT NULL,
    active_balance          DECIMAL(65, 0)           NOT NULL,
    commission              DECIMAL(65, 0)           NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_validator_sequences_height on validator_sequences (height);
