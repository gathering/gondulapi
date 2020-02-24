package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var connstr string
var DB *sql.DB

func init() {
	var err error
	connstr = "user=kly password=lolkek dbname=klytest sslmode=disable"
	DB, err = sql.Open("postgres", connstr)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
	err = DB.Ping()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
}

/*

func (sq *SQL) Send(c *skogul.Container) error {
	sq.once.Do(func() {
		sq.init()
	})
	if sq.initErr != nil {
		sqlLog.WithError(sq.initErr).Error("Database initialization failed")
		return sq.initErr
	}
	txn, err := sq.db.Begin()
	if err != nil {
		sqlLog.WithError(err).Error("Acquiring database transaction failed")
		return err
	}
	defer func() {
		if err != nil {
			sqlLog.WithError(skogul.Error{Source: "sql sender", Reason: "failed to send", Next: err}).Error("Failed to send")
			txn.Rollback()
		}
	}()

	stmt, err := txn.Prepare(sq.q)
	if err != nil {
		return err
	}

	for _, m := range c.Metrics {
		if err = sq.exec(stmt, m); err != nil {
			return err
		}
	}

	if err = stmt.Close(); err != nil {
		return err
	}

	if err = txn.Commit(); err != nil {
		return err
	}
	return nil
}

// Verify ensures options are set, but currently doesn't check very well,
// since it is disallowed from connecting to a database and such.
func (sq *SQL) Verify() error {
	if sq.ConnStr == "" {
		return skogul.Error{Source: "sql sender", Reason: "ConnStr is empty"}
	}
	if sq.Query == "" {
		return skogul.Error{Source: "sql sender", Reason: "Query is empty"}
	}
	if sq.Driver == "" {
		return skogul.Error{Source: "sql sender", Reason: "Driver is empty"}
	}
	if sq.Driver != "mysql" && sq.Driver != "postgres" {
		return skogul.Error{Source: "sql sender", Reason: fmt.Sprintf("unsuported database driver %s - must be `mysql' or `postgres'", sq.Driver)}
	}
	return nil
}
*/
