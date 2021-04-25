package model

import "time"

// DBObject is an object in PostgreSQL.
type DBObject struct {
	tableName struct{}  `pg:"objects"`
	ID        uint      `pg:"id,notnull,pk"`
	LastSeen  time.Time `pg:"last_seen,notnull"`
}
