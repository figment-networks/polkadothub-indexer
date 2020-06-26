CREATE TABLE IF NOT EXISTS syncables
(
    id              BIGSERIAL                NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL,

    spec_version    DECIMAL(65, 0)           NOT NULL,
    chain_uid       TEXT                     NOT NULL,
    height          DECIMAL(65, 0)           NOT NULL,
    session         DECIMAL(65, 0)           NOT NULL,
    era             DECIMAL(65, 0)           NOT NULL,
    last_in_session BOOLEAN                  NOT NULL,
    last_in_era     BOOLEAN                  NOT NULL,
    time            TIMESTAMP WITH TIME ZONE NOT NULL,

    index_version   INT                      NOT NULL,
    status          SMALLINT DEFAULT 0,
    report_id       BIGINT,
    started_at      TIMESTAMP WITH TIME ZONE,
    processed_at    TIMESTAMP WITH TIME ZONE,
    duration        BIGINT,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_syncables_report_id on syncables (report_id);
CREATE index idx_syncables_height on syncables (height);
CREATE index idx_syncables_session on syncables (session);
CREATE index idx_syncables_era on syncables (era);
CREATE index idx_syncables_index_version on syncables (index_version);
CREATE index idx_syncables_processed_at on syncables (processed_at);
