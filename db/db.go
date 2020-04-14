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

	_ "github.com/lib/pq" // for postgres support
	log "github.com/sirupsen/logrus"
)

// Connstr is the datbase/sql connection string used to open a connection
// to a postgres database. We only support postgres, because why would you
// need anything else?
var Connstr string

// DB is the main database handle used throughout the API
var DB *sql.DB

func Connect() {
	var err error
	if Connstr == "" {
		log.Printf("Using default connection string for debug purposes")
		Connstr = "user=kly password=lolkek dbname=klytest sslmode=disable"
	}
	DB, err = sql.Open("postgres", Connstr)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
	err = DB.Ping()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
	}
}

