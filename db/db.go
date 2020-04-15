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
	log "github.com/sirupsen/logrus"
)

// DB is the main database handle used throughout the API
var DB *sql.DB

// Connect sets up the database connection, using the configured
// ConnectionString, and ensures it is working.
func Connect() error {
	var err error
	if gapi.Config.ConnectionString == "" {
		log.Printf("Using default connection string for debug purposes")
		gapi.Config.ConnectionString = "user=kly password=lolkek dbname=klytest sslmode=disable"
	}
	DB, err = sql.Open("postgres", gapi.Config.ConnectionString)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return gapi.Error{500, "Failed to connect to database"}
	}
	err = DB.Ping()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return gapi.Error{500, "Failed to connect to database"}
	}
	return nil
}
