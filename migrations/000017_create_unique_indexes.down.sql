DROP index IF EXISTS idx_account_era_sequences_accounts_era;

DROP index IF EXISTS idx_transaction_seq_height_index;

CREATE index idx_transaction_seq on transaction_sequences (height, index);

DROP index IF EXISTS idx_validator_sequences_height_stash;

DROP index IF EXISTS system_events_height_actor_kind;
