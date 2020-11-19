INSERT INTO validator_session_sequences (
  session,
  start_height,
  end_height,
  time,
  stash_account,
  online
)
VALUES @values

ON CONFLICT (session, stash_account) DO UPDATE
SET
  b.online = m.online
