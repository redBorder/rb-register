// Copyright (C) 2016 Eneo Tecnologia S.L.
// Diego Fernández Barrera <bigomby@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"database/sql"

	"github.com/sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqlCreateTable   = "CREATE TABLE IF NOT EXISTS Devices (Hash varchar(255) PRIMARY KEY, Uuid varchar(255))"
	sqlInsertEntry   = "INSERT INTO Devices (Hash, Uuid) values (?, ?)"
	sqlSelectDevices = "SELECT * FROM Devices"
)

// Database handles the connection with a SQL Database
type Database struct {
	config DatabaseConfig
}

// NewDatabase creates a new instance of a database
func NewDatabase(config DatabaseConfig) *Database {
	db := &Database{
		config: config,
	}

	if db.config.Logger == nil {
		db.config.Logger = logrus.New()
	}
	logger := db.config.Logger

	if len(db.config.dbFile) <= 0 {
		return nil
	}

	var err error
	db.config.sqldb, err = sql.Open("sqlite3", db.config.dbFile)
	if err != nil {
		logger.Fatal(err)
	}

	// Ping db to check if db is available
	if err := db.config.sqldb.Ping(); err != nil {
		logger.Error(err)
		return nil
	}

	// Init connection with db
	if _, err := db.config.sqldb.Begin(); err != nil {
		logger.Error(err)
		return nil
	}

	// Create table if no exists
	if _, err := db.config.sqldb.Exec(sqlCreateTable); err != nil {
		logger.Error(err)
		return nil
	}

	return db
}

// LoadUUID loads from the database the UUID used along with a previous HASH
func (db *Database) LoadUUID(hash string) (uuid string, err error) {
	logger := db.config.Logger

	// Get all the rows and load data
	rows, err := db.config.sqldb.Query(sqlSelectDevices)
	if err != nil {
		return
	}

	for rows.Next() {
		var devHash sql.NullString
		var devUUID sql.NullString

		if err = rows.Scan(&devHash, &devUUID); err != nil {
			return
		}

		// If hash is in database load the associated UUID
		if devHash.String == hash {
			uuid = devUUID.String
			break
		}
	}

	if len(uuid) > 0 {
		logger.Debugf("Loaded UUID from DB: %s", uuid)
	}

	return
}

// StoreUUID save the UUD to a database
func (db *Database) StoreUUID(hash, uuid string) error {
	logger := db.config.Logger

	// Insert data into DB
	if _, err := db.config.sqldb.Exec(sqlInsertEntry, hash, uuid); err != nil {
		return err
	}

	logger.Infof("Stored UUID on DB: %s", uuid)

	return nil
}

// Close closes the connection with the database
func (db *Database) Close() {
	db.config.sqldb.Close()
}
