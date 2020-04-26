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

package db

import (
	"fmt"
	"github.com/gathering/gondulapi"
	log "github.com/sirupsen/logrus"
	"reflect"
)

// Select populates the provided interface(should be a pointer) by
// performing a simple select on the table, matching haystack with needle.
// E.g: select (elements of d) from table where haystack = needle;
//
// It is not particularly fast, and uses reflection. It is well suited for
// simple objects, but if performance is important, roll your own.
//
// It only populates exported values, and will use the "column" tag as an
// alternate name if needed. The data fields are fetched with sql.Scan(),
// so the data types need to implement sql.Scan() somehow.
//
// If the element is found, "found" is returned as true. If the element is
// found, but fails to scan, found is returned as true, but err is is
// non-nil. If an error occurred before the query is executed, or with the
// query itself (e.g.: bad table, or database issues), found is returned as
// false, and error is set. As such: if found is true, you can trust it, if
// it is false, you can only be absolutely sure the element doesn't exist
// if err is false.
//
// It needs to do two passes (might be a better way if someone better at
// the inner workings of Go reflection than me steps up). The first pass
// iterates over the fields of d, preparing both the query and allocating
// zero-values of the relevant objects. After this, the query is executed
// and the values are stored on the temporary values. The last pass stores
func Select(needle interface{}, haystack string, table string, d interface{}) (found bool, err error) {
	st := reflect.ValueOf(d)
	if st.Kind() != reflect.Ptr {
		log.Error("Select() called with non-pointer interface. This wouldn't really work.")
		return false,gondulapi.InternalError
	}
	st = reflect.Indirect(st)

	// Set up a slice for the response
	retv := reflect.MakeSlice(reflect.SliceOf(st.Type()), 0, 0)
	retvi := retv.Interface()

	// Do the actual work :D
	err = SelectMany(needle, haystack, table, &retvi)

	if err != nil {
		log.WithError(err).Error("Call to SelectMany() from Select() failed.")
		return false,gondulapi.InternalError
	}
	// retvi will be overwritten with the response (because that's how
	// append works), so retv now points to the empty original - update
	// it.
	retv = reflect.ValueOf(retvi)
	if retv.Len() == 0 {
		return false, nil
	}
	reply := retv.Index(0)
	setthis := reflect.Indirect(reflect.ValueOf(d))
	setthis.Set(reply)
	return true, nil
}

// SelectMany selects multiple rows from the table, populating the slice
// pointed to by d. It must, as such, be called with a pointer to a slice as
// the d-value (it checks).
//
// It returns the data in d, with an error if something failed, obviously.
// It's not particularly fast, or bullet proof, but:
//
// 1. It handles the needle safely, e.g., it lets the sql driver do the
// escaping.
//
// 2. The haystack and table is NOT safe.
//
// 3. It uses database/sql.Scan, so as long as your elements implement
// that, it will Just Work.
//
// It works by first determining the base object/type to fetch by digging
// into d with reflection. Once that is established, it iterates over the
// discovered base-structure and does two things: creates the list of
// columns to SELECT, and creates a reflect.Value for each column to store
// the result. Once this loop is done, it executes the query, then iterates
// over the replies, storing them in new base elements. At the very end,
// the *d is overwritten with the new slice.
func SelectMany(needle interface{}, haystack string, table string, d interface{}) error {
	if DB == nil {
		log.Errorf("Tried to issue SelectMany() without a DB object")
		return gondulapi.InternalError
	}
	dval := reflect.ValueOf(d)
	// This is needed because we need to be able to update with a
	// potentially new slice.
	if dval.Kind() != reflect.Ptr {
		log.Errorf("SelectMany() called with non-pointer interface. This wouldn't really work. Got %T", d)
		return gondulapi.InternalError
	}
	dval = reflect.Indirect(dval)

	// This enables Select() to work, and generally masks over issues
	// where the type is obscured by layers of casting and conversion.
	if dval.Kind() == reflect.Interface {
		dval = dval.Elem()
	}
	// And obviously it needs to actually be a slice.
	if dval.Kind() != reflect.Slice {
		log.Errorf("SelectMany() must be called with pointer-to-slice, e.g: &[]foo, got: %T inner is: %v / %#v / %s / kind: %s", d, dval, dval, dval, dval.Kind())
		return gondulapi.InternalError
	}

	// st stores the type we need to return an array, while fieldList
	// stores the actual base element. Usually, they are the same,
	// unless you pass []*foo, in which case st will represent *foo and
	// fieldList will represent foo.
	st := dval.Type()
	st = st.Elem()
	fieldList := st
	if fieldList.Kind() == reflect.Ptr {
		fieldList = fieldList.Elem()
	}

	// We make a new slice - this is what we will actually return/set
	retv := reflect.MakeSlice(reflect.SliceOf(st), 0, 0)

	keys, comma := "", ""
	sample := reflect.New(fieldList)
	sampleUnderscoreRaw := sample.Interface()
	kvs, err := enumerate("-", true, &sampleUnderscoreRaw)
	if err != nil {
		log.WithError(err).Errorf("enumerate() failed during query. This is bad.")
		return gondulapi.InternalError
	}
	for idx := range kvs.keys {
		keys = fmt.Sprintf("%s%s%s", keys, comma, kvs.keys[idx])
		comma = ","
	}
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", keys, table, haystack)
	log.WithField("query", q).Tracef("Select()")
	rows, err := DB.Query(q, needle)
	if err != nil {
		log.WithError(err).WithField("query", q).Info("Select(): SELECT failed on DB.Query")
		return gondulapi.InternalError
	}
	defer func() {
		rows.Close()
	}()

	// Read the rows...
	for {
		ok := rows.Next()
		if !ok {
			break
		}
		err = rows.Scan(kvs.newvals...)
		if err != nil {
			log.WithError(err).WithField("query", q).Info("Select(): SELECT failed to scan")
			return gondulapi.InternalError
		}
		log.Tracef("SELECT(): Record found and scanned.")

		// Create the new slice element
		newidx := reflect.New(st)
		newidx = reflect.Indirect(newidx)

		// If it's an array of pointers, we need to fiddle a bit.
		// This is probably not prefect.
		newval := newidx
		if newidx.Kind() == reflect.Ptr {
			newval = reflect.New(st.Elem()) // returns a _pointer_ to the new value, which is why this works.
			newidx.Set(newval)
			newval = reflect.Indirect(newval)
		}

		for idx := range kvs.newvals {
			newv := reflect.Indirect(reflect.ValueOf(kvs.newvals[idx]))
			value := newval.Field(kvs.keyidx[idx])
			value.Set(newv)
		}
		retv = reflect.Append(retv, newidx)
	}

	// Finally - store the new slice to the pointer provided as input
	setthis := reflect.Indirect(reflect.ValueOf(d))
	setthis.Set(retv)
	return nil
}

// Exists checks if a row where haystack matches the needle exists on the
// given table. It returns found=true if it does. It returns found=false if
// it doesn't find it - including if an error occurs (which will also be
// returned).
func Exists(needle interface{}, haystack string, table string) (found bool, err error) {
	q := fmt.Sprintf("SELECT * FROM %s WHERE %s = $1 LIMIT 1", table, haystack)
	rows, err := DB.Query(q, needle)
	if err != nil {
		log.WithError(err).WithField("query", q).Info("Exists(): SELECT failed")
		return false,gondulapi.InternalError
	}
	defer func() {
		// XXX: Unsure if this is actually needed here, to be
		// honest.
		rows.Close()
	}()
	ok := rows.Next()
	if !ok {
		return false,nil
	}
	found = true
	return
}

// Get is a convenience-wrapper for Select that return suitable
// gondulapi-errors if the needle is the Zero-value, if the database-query
// fails or if the item isn't found.
//
// It is provided so callers can implement receiver.Getter by simply
// calling this to get reasonable default-behavior.
func Get(needle interface{}, haystack string, table string, item interface{}) error {
	value := reflect.ValueOf(needle)
	if value.IsZero() {
		return gondulapi.Errorf(400, "No item to look for provided")
	}
	found, err := Select(needle, haystack, table, item)
	if err != nil {
		return gondulapi.InternalError
	}
	if !found {
		return gondulapi.Errorf(404, "Couldn't find item %v", needle)
	}
	return nil
}
