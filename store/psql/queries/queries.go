
package queries

const (
	
	// store/psql/queries/account_era_seq_find_last_by_stash.sql
	AccountEraSeqFindLastByStash = `SELECT * FROM account_era_sequences   WHERE stash_account = ?   AND era = ( 	SELECT era  		FROM account_era_sequences  		WHERE stash_account = ?  		GROUP BY era  		ORDER BY era LIMIT 1);`
	
	// store/psql/queries/account_era_seq_find_last_by_validator_stash.sql
	AccountEraSeqFindLastByValidatorStash = `SELECT * FROM account_era_sequences   WHERE validator_stash_account = ?   AND era = ( 	SELECT era  		FROM account_era_sequences  		WHERE validator_stash_account = ?  		GROUP BY era  		ORDER BY era LIMIT 1);`
	
	// store/psql/queries/account_era_seq_insert.sql
	AccountEraSeqInsert = `INSERT INTO account_era_sequences (   era,   start_height,   end_height,   time,   stash_account,   controller_account,   validator_stash_account,   validator_controller_account,   stake ) VALUES @values  ON CONFLICT (era, stash_account, validator_stash_account) DO UPDATE SET   controller_account                 = excluded.controller_account,   validator_controller_account       = excluded.validator_controller_account,   stake                              = excluded.stake `
	
	// store/psql/queries/block_seq_summarize.sql
	BlockSeqSummarize = `DATE_TRUNC(?, time) AS time_bucket, COUNT(*) AS count, EXTRACT(EPOCH FROM (MAX(time) - MIN(time)) / COUNT(*)) AS block_time_avg`
	
	// store/psql/queries/block_seq_times.sql
	BlockSeqTimes = `SELECT    MIN(height) start_height,    MAX(height) end_height,    MIN(time) start_time,   MAX(time) end_time,   COUNT(*) count,    EXTRACT(EPOCH FROM MAX(time) - MIN(time)) AS diff,    EXTRACT(EPOCH FROM ((MAX(time) - MIN(time)) / COUNT(*))) AS avg   FROM (      SELECT * FROM block_sequences     ORDER BY height DESC     LIMIT ?   ) t;`
	
	// store/psql/queries/block_summary_activity_periods.sql
	BlockSummaryActivityPeriods = `WITH cte AS (     SELECT       time_bucket,       sum(CASE WHEN diff IS NULL OR diff > ? :: INTERVAL         THEN 1           ELSE NULL END)       OVER (         ORDER BY time_bucket ) AS period     FROM (            SELECT              time_bucket,              time_bucket - lag(time_bucket, 1)              OVER (                ORDER BY time_bucket ) AS diff            FROM block_summary            WHERE time_interval = ? AND index_version = ?          ) AS x ) SELECT   period,   MIN(time_bucket),   MAX(time_bucket) FROM cte GROUP BY period ORDER BY period`
	
	// store/psql/queries/block_summary_for_interval.sql
	BlockSummaryForInterval = `SELECT *  FROM block_summary  WHERE time_bucket >= ( 	SELECT time_bucket  	FROM block_summary  	WHERE time_interval = ? 	ORDER BY time_bucket DESC 	LIMIT 1 ) - ?::INTERVAL AND time_interval = ? ORDER BY time_bucket`
	
	// store/psql/queries/event_seq_insert.sql
	EventSeqInsert = `INSERT INTO event_sequences (   height,   time,   index,   extrinsic_index,   data,   phase,   method,   section ) VALUES @values  ON CONFLICT (height, index) DO UPDATE SET   extrinsic_index    = excluded.extrinsic_index,   data               = excluded.data,   phase              = excluded.phase,   method             = excluded.method,   section            = excluded.section `
	
	// store/psql/queries/event_seq_with_tx_hash_for_src.sql
	EventSeqWithTxHashForSrc = `	SELECT 		e.height, 		e.method, 		e.section, 		e.data, 		t.hash 	FROM event_sequences AS e 	INNER JOIN transaction_sequences as t 		ON t.height = e.height AND t.index = e.extrinsic_index 	WHERE e.section = ? AND e.method = ? AND e.data->0->>'value' = ?`
	
	// store/psql/queries/event_seq_with_tx_hash_for_src_and_target.sql
	EventSeqWithTxHashForSrcAndTarget = `	SELECT 		e.height, 		e.method, 		e.section, 		e.data, 		t.hash 	FROM event_sequences AS e 	INNER JOIN transaction_sequences as t 		ON t.height = e.height AND t.index = e.extrinsic_index 	WHERE e.section = ? AND e.method = ? AND (e.data->0->>'value' = ? OR e.data->1->>'value' = ?)`
	
	// store/psql/queries/system_event_insert.sql
	SystemEventInsert = `INSERT INTO system_events (   created_at,   updated_at,   height,   time,   actor,   kind,   data ) VALUES @values  ON CONFLICT (height, actor, kind) DO UPDATE SET   updated_at   = excluded.updated_at,   data         = excluded.data `
	
	// store/psql/queries/transaction_seq_insert.sql
	TransactionSeqInsert = `INSERT INTO transaction_sequences (   height,   time,   index,   hash,   method,   section ) VALUES @values  ON CONFLICT (height, index) DO UPDATE SET   hash     = excluded.hash,   method   = excluded.method,   section  = excluded.section `
	
	// store/psql/queries/validator_era_seq_insert.sql
	ValidatorEraSeqInsert = `INSERT INTO validator_era_sequences (   era,   start_height,   end_height,   time,   stash_account,   controller_account,   session_accounts,   index,   total_stake,   own_stake,   stakers_stake,   reward_points,   commission,   stakers_count ) VALUES @values  ON CONFLICT (era, stash_account) DO UPDATE SET   b.controller_account = m.controller_account,   b.session_accounts = m.session_accounts,   b.index = m.index,   b.total_stake = m.total_stake,   b.own_stake = m.own_stake,   b.stakers_stake = m.stakers_stake,   b.reward_points = m.reward_points,   b.commission = m.commission,   b.stakers_count = m.stakers_count `
	
	// store/psql/queries/validator_era_seq_summarize_select.sql
	ValidatorEraSeqSummarizeSelect = `	stash_account, 	DATE_TRUNC(?, time) AS time_bucket,    	AVG(total_stake) AS total_stake_avg,    	MAX(total_stake) AS total_stake_max,    	MIN(total_stake) AS total_stake_min, 	AVG(own_stake) AS own_stake_avg,    	MAX(own_stake) AS own_stake_max,    	MIN(own_stake) AS own_stake_min, 	AVG(stakers_stake) AS stakers_stake_avg,    	MAX(stakers_stake) AS stakers_stake_max,    	MIN(stakers_stake) AS stakers_stake_min, 	AVG(reward_points) AS reward_points_avg,    	MAX(reward_points) AS reward_points_max,    	MIN(reward_points) AS reward_points_min, 	AVG(commission) AS commission_avg,    	MAX(commission) AS commission_max,    	MIN(commission) AS commission_min, 	AVG(stakers_count) AS stakers_count_avg,    	MAX(stakers_count) AS stakers_count_max,    	MIN(stakers_count) AS stakers_count_min`
	
	// store/psql/queries/validator_seq_insert.sql
	ValidatorSeqInsert = `INSERT INTO validator_sequences (   height,   time,   stash_account,   active_balance,   commission ) VALUES @values  ON CONFLICT (height, stash_account) DO UPDATE SET   active_balance   = excluded.active_balance,   commission       = excluded.commission `
	
	// store/psql/queries/validator_session_seq_get_counts.sql
	ValidatorSessionSeqGetCounts = `	stash_account, 	COUNT(*) AS count`
	
	// store/psql/queries/validator_session_seq_insert.sql
	ValidatorSessionSeqInsert = `INSERT INTO validator_session_sequences (   session,   start_height,   end_height,   time,   stash_account,   online ) VALUES @values  ON CONFLICT (session, stash_account) DO UPDATE SET   b.online = m.online `
	
	// store/psql/queries/validator_session_seq_summarize_select.sql
	ValidatorSessionSeqSummarizeSelect = `	stash_account, 	DATE_TRUNC(?, time) AS time_bucket,    	AVG(online::INT) AS uptime_avg,    	MAX(online::INT) AS uptime_max,    	MIN(online::INT) AS uptime_min`
	
	// store/psql/queries/validator_summary_activity_periods.sql
	ValidatorSummaryActivityPeriods = `WITH cte AS (     SELECT       time_bucket,       sum(CASE WHEN diff IS NULL OR diff > ? :: INTERVAL         THEN 1           ELSE NULL END)       OVER (         ORDER BY time_bucket ) AS period     FROM (            SELECT              time_bucket,              time_bucket - lag(time_bucket, 1)              OVER (                ORDER BY time_bucket ) AS diff            FROM validator_summary            WHERE time_interval = ? AND index_version = ?          ) AS x ) SELECT   period,   MIN(time_bucket),   MAX(time_bucket) FROM cte GROUP BY period ORDER BY period`
	
	// store/psql/queries/validator_summary_for_interval.sql
	ValidatorSummaryForInterval = `SELECT   time_bucket,   time_interval,    AVG(total_stake_avg) AS total_stake_avg,   MAX(total_stake_max) AS total_stake_max,   MIN(total_stake_min) AS total_stake_min,   AVG(own_stake_avg) AS own_stake_avg,   MAX(own_stake_max) AS own_stake_max,   MIN(own_stake_min) AS own_stake_min,   AVG(stakers_stake_avg) AS stakers_stake_avg,   MAX(stakers_stake_max) AS stakers_stake_max,   MIN(stakers_stake_min) AS stakers_stake_min,   AVG(reward_points_avg) AS reward_points_avg,   MAX(reward_points_max) AS reward_points_max,   MIN(reward_points_min) AS reward_points_min,   AVG(commission_avg) AS commission_avg,   MIN(commission_min) AS commission_min,   MAX(commission_max) AS commission_max,   AVG(stakers_count_avg) AS stakers_count_avg,   MIN(stakers_count_min) AS stakers_count_min,   MAX(stakers_count_max) AS stakers_count_max,   AVG(uptime_avg) AS uptime_avg FROM validator_summary WHERE time_bucket >= ( 	SELECT time_bucket  	FROM validator_summary  	WHERE time_interval = ? 	ORDER BY time_bucket DESC  	LIMIT 1 ) - ?::INTERVAL 	AND time_interval = ? GROUP BY time_bucket, time_interval ORDER BY time_bucket`
	
	// store/psql/queries/validator_summary_for_interval_and_stash.sql
	ValidatorSummaryForIntervalAndStash = `SELECT *  FROM validator_summary  WHERE time_bucket >= ( 	SELECT time_bucket  	FROM validator_summary  	WHERE time_interval = ? 	ORDER BY time_bucket DESC 	LIMIT 1 ) - ?::INTERVAL 	AND stash_account = ? AND time_interval = ? ORDER BY time_bucket`
	
)
	