INSERT INTO transaction_sequences (
  height,
  time,
  index,
  hash,
  method,
  section
)
VALUES @values

ON CONFLICT (height, index) DO UPDATE
SET
  hash     = excluded.hash,
  method   = excluded.method,
  section  = excluded.section
