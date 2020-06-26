CREATE TABLE IF NOT EXISTS block_summary
(
    id                            BIGSERIAL                NOT NULL,
    created_at                    TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at                    TIMESTAMP WITH TIME ZONE NOT NULL,

    time_interval                 VARCHAR                  NOT NULL,
    time_bucket                   TIMESTAMP WITH TIME ZONE NOT NULL,
    index_version                 INT                      NOT NULL,

    count                         BIGINT                   NOT NULL,
    block_time_avg                DECIMAL                  NOT NULL,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_block_summary_time on block_summary (time_interval, time_bucket);
CREATE index idx_block_summary_index_version on block_summary (index_version);