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

This file demonstrates three things:

1. How to implement basic GET/PUT/POST/DELETE using gondulapi/receiver.

2. How to use the simplified gondulapi/db methods to map your own interface
directly to a database element without writing SQL.

3. Implementing custom-types that are supported implicitly by gondulapi/db.

*/
package objects

import (
	"database/sql/driver"
	"fmt"
	"github.com/gathering/gondulapi"
	"github.com/gathering/gondulapi/db"
	"github.com/gathering/gondulapi/receiver"
	"github.com/gathering/gondulapi/types"
	log "github.com/sirupsen/logrus"
)

/*
Thing is a dummy structure to illustrate the core GET/PUT/POST/DELETE
API.

It mimics a common pattern where an object also contains its own name.

We use a tag to inform gondulapi/db that MgmtIP is mapped to the "ip"
column in the database.
*/
type Thing struct {
	Sysname   string
	MgmtIP    *types.IP `column:"ip"`
	Placement *ThingPlacement
}

// ThingPlacement illustrates nested data structures which are implicitly
// handled, and, also mimicks the placement logic of switches in Gondul,
// which use an X1/X2 Y1/Y2 coordinate system to determine left-most and
// right-most X and top-most and bottom-most Y.
//
// Please note that Postgresql will always place the upper right corner
// first - so X1/Y1 might be X2/Y2 when you read it back out - but the
// semantics are correct.
type ThingPlacement struct {
	X1 int
	X2 int
	Y1 int
	Y2 int
}

// Scan implements sql.Scan to enable gondulapi/db to read the SQL directly
// into our custom type.
func (tp *ThingPlacement) Scan(src interface{}) error {
	switch value := src.(type) {
	case []byte:
		str := string(value)
		_, err := fmt.Sscanf(str, "(%d,%d),(%d,%d)", &tp.X1, &tp.Y1, &tp.X2, &tp.Y2)
		return err
	default:
		return fmt.Errorf("invalid IP")
	}
}

// Value implements sql/driver's Value interface to provide implicit
// support for writing the value.
//
// In other words: Scan() enables SELECT and Value() allows INSERT/UPDATE.
func (tp ThingPlacement) Value() (driver.Value, error) {
	x := fmt.Sprintf("((%d,%d),(%d,%d))", tp.X1, tp.Y1, tp.X2, tp.Y2)
	return x, nil
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
	found, err := db.Select(element, "sysname", "things", b)
	if err != nil {
		log.Printf("Error: %v", err)
		return gondulapi.Errorf(500, "Unable to fetch %s from the database....", element)
	}
	if !found {
		return gondulapi.Errorf(404, "Couldn't find %s", element)
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
	return db.Upsert(b.Sysname, "sysname", "things", b)
}

// Post stores the provided object. Unlike PUT, it will always use INSERT.
func (b Thing) Post() error {
	return db.Insert("things", b)
}

// Delete is called to delete an element.
func (b Thing) Delete(element string) error {
	err := db.Delete(element, "sysname", "things")
	if err != nil {
		fmt.Printf("delete lolz: %v\n", err)
	}
	return err
}
