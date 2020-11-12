INSERT INTO event_sequences (
  height,
  time,
  index,
  extrinsic_index,
  data,
  phase,
  method,
  section
)
VALUES @values

ON CONFLICT (height, index) DO UPDATE
SET
  extrinsic_index    = excluded.extrinsic_index,
  data               = excluded.data,
  phase              = excluded.phase,
  method             = excluded.method,
  section            = excluded.section
