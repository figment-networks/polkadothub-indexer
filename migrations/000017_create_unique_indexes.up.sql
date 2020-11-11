CREATE UNIQUE INDEX idx_account_era_sequences_accounts_era
  ON account_era_sequences(stash_account, validator_stash_account, era);