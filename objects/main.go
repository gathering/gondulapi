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

This file demonstrates two things:

1. How to implement basic GET/PUT/POST/DELETE using gondulapi/receiver.

2. How to use the simplified gondulapi/db methods to map your own interface
directly to a database element without writing SQL.

See types/ for how to implement custom-types not already supported by
database/sql.
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
API.

It mimics a common pattern where an object also contains its own name.

We use a tag to inform gondulapi/db that MgmtIP is mapped to the "ip"
column in the database.

Special note here: The only way for Go to distinguish between "The variable
was not provided" and "the variable was provided, but set to the
zero-value" is if the field is a pointer. That means that if Vlan is a
nil-pointer in Put, the client didn't specify it and we should not modify
it. If, however, it is a allocated, pointing to int<0>, it was provided,
but set to 0. This is important: If you change Vlan to "Vlan int" instead
of "Vlan *int", two things will happen:

1. If a client doesn't provide vlan on an update with PUT, we will still
set it to 0.

2. If a row in the database has vlan set to NULL, we will fail to scan that
since NULL can not be converted into an int.

In general, that means you want pointers.

It also implies that you should use constraints on your database to ensure
that the values are legal. Set NOT NULL if it shouldn't be null. Set default
values if it's OK to omit it on INSERT, and so on.
*/
type Thing struct {
	Sysname   string
	MgmtIP    *types.IP `column:"ip"`
	Vlan      *int
	Placement *types.Box
}

// Things demonstrates how to fetch multiple items at the same time.
type Things []Thing

func init() {
	// This is how we register for a url. The url is the same as used
	// for net/http. The func()... is something you can cargo-cult - it
	// is al allocation function for an empty instance of the data
	// model.
	receiver.AddHandler("/thing/", func() interface{} { return &Thing{} })
	receiver.AddHandler("/things/", func() interface{} { return &Things{} })
}

// Get is called on GET. b will be an empty thing. Fill it out, using the
// element to determine what we're looking for. If it fails: return an
// error. Simple.
func (b *Thing) Get(element string) error {
	return db.Get(element, "sysname", "things", b)
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
	}
	if b.Sysname == "" {
		log.Tracef("Blank sysname, using url-path")
		b.Sysname = element
	}
	if b.Sysname != element {
		return gondulapi.Errorf(400, "Thing url path %s doesn't match json-specified name %s", element, b.Sysname)
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

// Get uses SelectMany to fetch multiple rows. For this, "b" has to be a
// pointer to a slice, which it happens to be here (It's []Thing,
// remember).
func (b *Things) Get(element string) error {
	if element == "" {
		// Ok this is a bit of a hack, as it uses "WHERE 1 = 1" to
		// avoid having an other function without conditions.
		return db.SelectMany(1, "1", "things", b)
	}
	return db.SelectMany(element, "vlan", "things", b)
}
