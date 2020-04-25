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
	"fmt"
	"reflect"
	"unicode"

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
// them _back_ onto the original d interface.
func Select(needle string, haystack string, table string, d interface{}) (found bool, err error) {
	st := reflect.TypeOf(d)
	v := reflect.ValueOf(d)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
		v = v.Elem()
	}


	vals := make([]interface{},0)
	keys := ""
	comma := ""
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		value := v.Field(i)
		if !unicode.IsUpper(rune(field.Name[0])) {
			continue
		}
		col := field.Name
		if ncol, ok := field.Tag.Lookup("column"); ok {
			col = ncol
		}
		keys = fmt.Sprintf("%s%s%s",keys,comma,col)
		comma=", "
		newv := reflect.New(value.Type())
		intfval := newv.Interface()
		vals = append(vals, intfval)
	}
	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1",keys,table,haystack)
	rows, err := DB.Query(q,needle)
	if err != nil {
		return
	}
	defer func() {
		rows.Close()
	}()

	ok := rows.Next()
	if !ok {
		return
	}
	found = true
	err = rows.Scan(vals...)
	if err != nil {
		return
	}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		value := v.Field(i)
		if !unicode.IsUpper(rune(field.Name[0])) {
			continue
		}
		newv := reflect.Indirect(reflect.ValueOf(vals[i]))
		value.Set(newv)
	}
	return
}
