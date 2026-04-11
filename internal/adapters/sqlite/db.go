package sqlite

import (
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func Open(path string) (*bun.DB, error) {
	sqldb, err := sql.Open(sqliteshim.ShimName, "file:"+path+"?cache=shared")
	if err != nil {
		return nil, err
	}
	return bun.NewDB(sqldb, sqlitedialect.New()), nil
}
