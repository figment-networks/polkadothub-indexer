CREATE TABLE IF NOT EXISTS extrinsic_sequences
(
    id               BIGSERIAL                NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL,

    spec_version_uid DOUBLE PRECISION         NOT NULL,
    chain_uid        TEXT                     NOT NULL,
    height           DOUBLE PRECISION         NOT NULL,
    time             TIMESTAMP WITH TIME ZONE NOT NULL,

    index            DOUBLE PRECISION         NOT NULL,
    signature        TEXT,
    signer           TEXT,
    nonce            NUMERIC                  NOT NULL,
    method           TEXT                     NOT NULL,
    section          TEXT                     NOT NULL,
    args             TEXT,
    is_signed        BOOLEAN,

    PRIMARY KEY (time, id)
);

-- Hypertable
SELECT create_hypertable('extrinsic_sequences', 'time', if_not_exists => TRUE);

-- Indexes
CREATE index idx_extrinsic_sequences_hash on extrinsic_sequences (spec_version_uid, time DESC);
CREATE index idx_extrinsic_sequences_chain_id on extrinsic_sequences (chain_uid, time DESC);
CREATE index idx_extrinsic_sequences_height on extrinsic_sequences (height, time DESC);
CREATE index idx_extrinsic_sequences_public_key on extrinsic_sequences (index, time DESC);
