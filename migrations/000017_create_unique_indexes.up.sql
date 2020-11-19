CREATE UNIQUE INDEX idx_account_era_sequences_accounts_era
  ON account_era_sequences(stash_account, validator_stash_account, era);


DROP index IF EXISTS idx_transaction_seq;

CREATE UNIQUE INDEX idx_transaction_seq_height_index
  ON transaction_sequences(height, index);

CREATE UNIQUE INDEX idx_validator_sequences_height_stash
    ON validator_sequences(height, stash_account);

CREATE UNIQUE INDEX idx_system_events_height_actor_kind
    ON system_events(height, actor, kind);

CREATE UNIQUE INDEX idx_event_sequences_height_index
    ON event_sequences(height, index);