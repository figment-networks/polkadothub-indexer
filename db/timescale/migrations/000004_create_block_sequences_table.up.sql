CREATE TABLE IF NOT EXISTS block_sequences
(
    id                        BIGSERIAL                NOT NULL,
    created_at                TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                TIMESTAMP WITH TIME ZONE NOT NULL,

    spec_version_uid          DOUBLE PRECISION         NOT NULL,
    chain_uid                 TEXT                     NOT NULL,
    height                    DOUBLE PRECISION         NOT NULL,
    time                      TIMESTAMP WITH TIME ZONE NOT NULL,

    parent_hash               TEXT                     NOT NULL,
    state_root                TEXT                     NOT NULL,
    extrinsics_root           TEXT                     NOT NULL,
    extrinsics_count          DOUBLE PRECISION,
    signed_extrinsics_count   DOUBLE PRECISION,
    unsigned_extrinsics_count DOUBLE PRECISION,

    PRIMARY KEY (time, id)
);

-- Hypertable
SELECT create_hypertable('block_sequences', 'time', if_not_exists => TRUE);

-- Indexes
CREATE index idx_block_sequences_chain_id on block_sequences (chain_uid, time DESC);
CREATE index idx_block_sequences_height on block_sequences (height, time DESC);
CREATE index idx_block_sequences_app_version on block_sequences (spec_version_uid, time DESC);
