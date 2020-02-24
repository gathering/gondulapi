/*
Gondul GO API, core objects
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

/*
Package objects implements a number of objects. Currently just one, though,
and a dummy-one at that.

It is intended to be imported for side effects, but there's no reason these
objects can't be exported like any other.

An object is a data structure that implements one of Putter, Getter, Poster
and Deleter from gondulapi to do RESTful stuff.

The current Thing object is intended as a demonstration.
*/
package objects

import (
	"fmt"
	"github.com/gathering/gondulapi"
	"github.com/gathering/gondulapi/db"
	"github.com/gathering/gondulapi/receiver"
	"github.com/gathering/gondulapi/types"
	log "github.com/sirupsen/logrus"
)

/*
Thing is a dummy structure to illustrate the core GET/PUT/POST/DELETE
API. It doesn't implement a persistent storage/database connection, only
stores data in memory.

It mimics a common pattern where an object also contains its own name.
*/
type Thing struct {
	Sysname   string
	MgmtIP    types.IP
	Placement ThingPlacement
}

// ThingPlacement illustrates nested data structures which are implicitly
// handled, and, also mimicks the placement logic of switches in Gondul,
// which use an X1/X2 Y1/Y2 coordinate system to determine left-most and
// right-most X and top-most and bottom-most Y.
//
// I can't really remember which is which, but that is besides the point!
type ThingPlacement struct {
	X1 int
	X2 int
	Y1 int
	Y2 int
}

func init() {

	// This is how we register for a url. The url is the same as used
	// for net/http. The func()... is something you can cargo-cult - it
	// is al allocation function for an empty instance of the data
	// model.
	receiver.AddHandler("/thing/", func() interface{} { return &Thing{} })
}

// Get is called on GET. b will be an empty thing. Fill it out, using the
// element to determine what we're looking for. If it fails: return an
// error. Simple.
func (b *Thing) Get(element string) error {
	rows, err := db.DB.Query("SELECT sysname,ip FROM things WHERE sysname = $1", element)
	if err != nil {
		return err
	}
	defer func() {
		rows.Close()
	}()

	ok := rows.Next()
	if !ok {
		return gondulapi.Errorf(404, "Thing %s doesn't exist", element)
	}
	err = rows.Scan(&b.Sysname, &b.MgmtIP)
	if err != nil {
		return err
	}
	if rows.Next() {
		return gondulapi.Errorf(500, "Thing %s has multiple copies....", element)
	}

	return nil
}

// Put is used to store an element with an absolute URL. In our case, the
// name of the element is also (potentially) present in the data it self -
// so we do a bit of magic. Note that this should NEVER generate a random
// name.
//
// b will contain the parsed data. element will be the name of the thing.
//
// PUT is idempotent. Calling it once with a set of parameters or a hundred
// times with the same parameters should yield the same result.
func (b Thing) Put(element string) error {
	if element == "" {
		return gondulapi.Errorf(400, "PUT requires an element path to put")
		element = b.Sysname
	}
	if b.Sysname == "" {
		log.Printf("Blank sysname, using url-path")
		b.Sysname = element
	}
	if b.Sysname != element {
		return fmt.Errorf("Thing url path %s doesn't match json-specified name %s", element, b.Sysname)
	}
	if b.exists(element) {
		return b.update()
	}
	return b.save()
}

func (b Thing) exists(element string) bool {
	var existing Thing
	err := existing.Get(element)
	exists := true
	if err != nil {
		gerr, ok := err.(gondulapi.Error)
		if ok && gerr.Code == 404 {
			exists = false
		}
	}
	return exists
}

func (b Thing) save() error {
	_, err := db.DB.Exec("INSERT INTO things (sysname,ip) VALUES($1,$2)", b.Sysname, b.MgmtIP.String())
	return err
}

func (b Thing) update() error {
	_, err := db.DB.Exec("UPDATE things SET ip = $1 WHERE sysname = $2", b.MgmtIP.String(), b.Sysname)
	return err
}

// Post stores the provided object. It's bugged. I know.
func (b Thing) Post() error {
	return b.save()
}

// Delete is called to delete an element.
func (b Thing) Delete(element string) error {
	_, err := db.DB.Exec("DELETE FROM things WHERE sysname = $1", b.Sysname)
	return err
}
