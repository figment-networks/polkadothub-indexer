package psql

import "fmt"

func getFindAllByHeightAndIndexQuery(table string) string {
	return fmt.Sprintf(`
	SELECT *
	FROM %s as t
	WHERE t.height = ? AND t.index IN (?)
	`, table)
}
