package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Хранилище метрик в базе данных
type DBStorage struct {
	*sql.DB
}

func NewDBStorage() *DBStorage {
	return &DBStorage{}
}

func (dbs *DBStorage) Open(driverName string, dataSourceName string) (err error) {
	dbs.DB, err = sql.Open(driverName, dataSourceName)
	if err != nil {
		return err
	}
	return nil
}
