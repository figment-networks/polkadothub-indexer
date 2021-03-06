SELECT * FROM account_era_sequences
  WHERE stash_account = ?
  AND era = (
	SELECT era 
		FROM account_era_sequences 
		WHERE stash_account = ? 
		GROUP BY era 
		ORDER BY era LIMIT 1);