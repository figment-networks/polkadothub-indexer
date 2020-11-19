SELECT * 
FROM block_summary 
WHERE time_bucket >= (
	SELECT time_bucket 
	FROM block_summary 
	WHERE time_interval = ?
	ORDER BY time_bucket DESC
	LIMIT 1
) - ?::INTERVAL AND time_interval = ?
ORDER BY time_bucket