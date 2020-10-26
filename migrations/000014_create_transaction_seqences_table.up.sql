CREATE TABLE IF NOT EXISTS transaction_sequences
(
    id             BIGSERIAL                NOT NULL,

    height          DECIMAL(65, 0)           NOT NULL,
    time            TIMESTAMP WITH TIME ZONE NOT NULL,

    index           BIGINT                   NOT NULL,
    hash            TEXT                     NOT NULL,
    method          TEXT                     NOT NULL,
    section         TEXT                     NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_transaction_seq on transaction_sequences (height, index);
