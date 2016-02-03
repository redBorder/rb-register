package database

import (
	"database/sql"

	"github.com/Sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
)

const (
	CREATE_TABLE   = "CREATE TABLE IF NOT EXISTS Devices (Hash varchar(255) PRIMARY KEY, Uuid varchar(255))"
	INSERT_ENTRY   = "Insert into Devices (Hash, Uuid) values (?, ?)"
	SELECT_DEVICES = "SELECT * FROM Devices"
)

var log *logrus.Logger

type Database struct {
	sqldb *sql.DB
}

func NewDatabase(sqldb *sql.DB) *Database {
	log = logrus.New()

	// Check if db is not nil
	if sqldb == nil {
		return nil
	}

	db := &Database{
		sqldb: sqldb,
	}

	// Ping db to check if db is available
	if err := db.sqldb.Ping(); err != nil {
		log.Error(err)
		return nil
	}

	// Init connection with db
	if _, err := db.sqldb.Begin(); err != nil {
		log.Error(err)
		return nil
	}

	// Create table if no exists
	if _, err := db.sqldb.Exec(CREATE_TABLE); err != nil {
		log.Error(err)
		return nil
	}

	return db
}

func (db *Database) LoadUuid(hash string) (uuid string, err error) {

	// Get all the rows and load data
	rows, err := db.sqldb.Query(SELECT_DEVICES)
	if err != nil {
		return
	}

	for rows.Next() {
		var devHash sql.NullString
		var devUuid sql.NullString

		if err = rows.Scan(&devHash, &devUuid); err != nil {
			return
		}

		// If hash is in database load the associated UUID
		if devHash.String == hash {
			uuid = devUuid.String
			break
		}
	}

	if len(uuid) > 0 {
		log.Debugf("Loaded UUID from DB: %s", uuid)
	}

	return
}

func (db *Database) StoreUuid(hash, uuid string) error {

	// Insert data into DB
	if _, err := db.sqldb.Exec(INSERT_ENTRY, hash, uuid); err != nil {
		return err
	}

	log.Infof("Stored UUID on DB: %s", uuid)

	return nil
}
