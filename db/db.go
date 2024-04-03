/*
Gondul GO API, database integration
Copyright 2020, Kristian Lyngstøl <kly@kly.no>

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

// Package db integrates with generic databases, so far it doesn't do much,
// but it's supposed to do more.
package db

import (
	"database/sql"

	gapi "github.com/gathering/gondulapi"
	_ "github.com/lib/pq" // for postgres support
	_ "github.com/go-sql-driver/mysql" // Imported for side effect/mysql support
	"github.com/gathering/gondulapi/log"
)

// DB is the main database handle used throughout the API
var DB *sql.DB

// Ping is a wrapper for DB.Ping: it checks that the database is alive.
// It's provided to add standard gondulapi-logging and error-types that can
// be exposed to users.
func Ping() error {
	if DB == nil {
		log.Printf("Ping() issued without a valid DB. Use Connect() first.")
		return gapi.Error{500, "Failed to communicate with the database"}
	}
	err := DB.Ping()
	if err != nil {
		log.Printf("Failed to ping the database: %v", err)
		return gapi.Error{500, "Failed to communicate with the database"}
	}
	log.Tracef("Ping() of db successful")
	return nil
}

// Connect sets up the database connection, using the configured
// ConnectionString, and ensures it is working.
func Connect() error {
	var err error
	// Mainly allowed because testing can easily trigger multiple
	// connects.
	if DB != nil {
		log.Printf("Got superfluous db.Connect(). Running a ping and hoping for the best.")
		return Ping()
	}

	if gapi.Config.ConnectionString == "" {
		log.Printf("Using default connection string for debug purposes. Relax, it's a very secure set of credentials.")
		gapi.Config.ConnectionString = "user=kly password=lolkek dbname=klytest sslmode=disable"
	}
	driver := gapi.Config.Driver
	if driver == "" {
		driver = "postgres"
		gapi.Config.Driver = "postgres"
	}
	DB, err = sql.Open(driver, gapi.Config.ConnectionString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return gapi.Error{500, "Failed to connect to database"}
	}
	return Ping()
}
