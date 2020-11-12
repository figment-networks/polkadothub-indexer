INSERT INTO system_events (
  created_at,
  updated_at,
  height,
  time,
  actor,
  kind,
  data
)
VALUES @values

ON CONFLICT (height, actor, kind) DO UPDATE
SET
  updated_at   = excluded.updated_at,
  data         = excluded.data
