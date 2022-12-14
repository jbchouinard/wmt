package history

import (
	"database/sql"
	"time"

	"github.com/jbchouinard/wmt/config"
	"github.com/jbchouinard/wmt/database"
	"github.com/jbchouinard/wmt/errored"
)

var enabled bool

func init() {
	config.ValidValues["history"] = map[string]bool{"yes": true, "no": true}
	config.DefaultValues["history"] = "yes"
	enabled = config.Get("history") == "yes"
	_, err := database.TxExec("CREATE TABLE IF NOT EXISTS history (ts TIMESTAMP, key TEXT, value TEXT)")
	errored.Check(err, "init db.history")
}

type Entry struct {
	Timestamp time.Time
	Key       string
	Value     string
}

func Add(k string, v string) {
	if !enabled {
		return
	}
	ts := time.Now().UTC()
	_, err := database.TxExec(
		"INSERT INTO history (ts, key, value) VALUES (?, ?, ?)",
		ts, k, v,
	)
	errored.Check(err, "history add")
}

func Purge(asOf time.Time) {
	_, err := database.TxExec(
		"DELETE FROM history WHERE ts < ?",
		asOf,
	)
	errored.Check(err, "history purge")
}

func GetLast(k string, n int) []*Entry {
	values := make([]*Entry, 0)

	err := database.TxQuery(
		"SELECT ts, key, value FROM history WHERE key=? ORDER BY ts DESC LIMIT ?",
		k, n,
	)(func(r *sql.Rows) error {
		var ts time.Time
		var key string
		var value string
		if err := r.Scan(&ts, &key, &value); err != nil {
			return err
		}
		values = append(values, &Entry{ts, key, value})
		return nil
	})
	errored.Check(err, "history get last")
	return values
}
