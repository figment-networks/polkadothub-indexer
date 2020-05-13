package blockseqrepo

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
	blockTimesForIntervalQuery = `
SELECT
  time_bucket($1, time) AS time_interval,
  COUNT(*) AS count,
  EXTRACT(EPOCH FROM (last(time, time) - first(time, time)) / COUNT(*)) AS avg
FROM block_sequences
  WHERE (
    SELECT time
    FROM block_sequences
    ORDER BY time DESC
    LIMIT 1
  ) - $2::INTERVAL < time
GROUP BY time_interval
ORDER BY time_interval ASC;
`
)
