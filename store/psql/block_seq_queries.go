package store

const (
	blockTimesForRecentBlocksQuery = `
SELECT 
  MIN(height) start_height, 
  MAX(height) end_height, 
  MIN(time) start_time,
  MAX(time) end_time,
  COUNT(*) count, 
  EXTRACT(EPOCH FROM MAX(time) - MIN(time)) AS diff, 
  EXTRACT(EPOCH FROM ((MAX(time) - MIN(time)) / COUNT(*))) AS avg
  FROM ( 
    SELECT * FROM block_sequences
    ORDER BY height DESC
    LIMIT ?
  ) t;
`

	summarizeBlocksQuerySelect = `
    DATE_TRUNC(?, time) AS time_bucket,
    COUNT(*) AS count,
    EXTRACT(EPOCH FROM (MAX(time) - MIN(time)) / COUNT(*)) AS block_time_avg
`
)
