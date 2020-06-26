CREATE TABLE IF NOT EXISTS block_sequences
(
    id                        BIGSERIAL                NOT NULL,

    height                    DOUBLE PRECISION         NOT NULL,
    time                      TIMESTAMP WITH TIME ZONE NOT NULL,

    extrinsics_count          DOUBLE PRECISION,
    signed_extrinsics_count   DOUBLE PRECISION,
    unsigned_extrinsics_count DOUBLE PRECISION,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_block_sequences_height on block_sequences (height);
CREATE index idx_block_sequences_time on block_sequences (time);
