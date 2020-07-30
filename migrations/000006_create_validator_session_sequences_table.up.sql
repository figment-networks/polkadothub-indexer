CREATE TABLE IF NOT EXISTS validator_session_sequences
(
    id            BIGSERIAL                NOT NULL,

    session       DECIMAL(65, 0)           NOT NULL,
    start_height  DECIMAL(65, 0)           NOT NULL,
    end_height    DECIMAL(65, 0)           NOT NULL,
    time          TIMESTAMP WITH TIME ZONE NOT NULL,

    stash_account TEXT                     NOT NULL,
    online        BOOLEAN,

    PRIMARY KEY (id)
);

-- Indexes
CREATE index idx_validator_session_sequences_session on validator_session_sequences (session);
CREATE index idx_validator_session_sequences_heights on validator_session_sequences (start_height, end_height);
CREATE index idx_validator_session_sequences_time on validator_session_sequences (time);
CREATE index idx_validator_session_sequences_stash_account on validator_session_sequences (stash_account);
