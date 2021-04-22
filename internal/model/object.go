package model

import "time"

// Object is an object :)
type Object struct {
	tableName struct{}  `pg:"objects"`
	ID        uint      `pg:"id,notnull,pk"`
	LastSeen  time.Time `pg:"last_seen,notnull"`
}
