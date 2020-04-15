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

	log "github.com/sirupsen/logrus"
)

type keyvals struct {
	keys   []string
	values []interface{}
}

func enumerate(haystack string, all bool, d interface{}) (keyvals, error) {
	st := reflect.TypeOf(d)
	v := reflect.ValueOf(d)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}

	kvs := keyvals{}
	kvs.keys = make([]string, 0)
	kvs.values = make([]interface{}, 0)

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

		if field.Type.Kind() == reflect.Ptr && value.IsNil() {
			if !all {
				continue
			}
		} else {
			value = reflect.Indirect(value)
		}
		if col == haystack || col == "-" {
			continue
		}
		kvs.keys = append(kvs.keys, col)
		kvs.values = append(kvs.values, value.Interface())
	}
	return kvs, nil
}

// Update attempts to update the object in the database, using the provided
// string and matching the haystack with the needle. It skips fields that
// are nil-pointers.
func Update(table string, haystack string, needle interface{}, d interface{}) error {
	kvs, err := enumerate(haystack, d)
	if err != nil {
		panic(err)
	}
	lead := fmt.Sprintf("UPDATE %s SET ", table)
	comma := ""
	last := 0
	for idx := range kvs.keys {
		lead = fmt.Sprintf("%s%s%s = $%d", lead, comma, kvs.keys[idx], idx+1)
		comma = ", "
		last = idx
	}
	lead = fmt.Sprintf("%s WHERE %s = $%d", lead, haystack, last+2)
	kvs.values = append(kvs.values, needle)
	_, err = DB.Exec(lead, kvs.values...)
	if err != nil {
		log.Printf("DB.Exec(\"%s\",kvs.values...) failed: %v", lead, err)
		return err
	}
	return nil
}

// Insert adds the object to the table specified. It only provides the
// non-nil-pointer objects as fields, so it is up to the caller and the
// database schema to enforce default values.
func Insert(table string, d interface{}) error {
	kvs, err := enumerate("-", d)
	if err != nil {
		panic(err)
	}
	lead := fmt.Sprintf("INSERT INTO %s (", table)
	middle := ""
	comma := ""
	for idx := range kvs.keys {
		lead = fmt.Sprintf("%s%s%s ", lead, comma, kvs.keys[idx])
		middle = fmt.Sprintf("%s%s$%d ", middle, comma, idx+1)
		comma = ", "
	}
	lead = fmt.Sprintf("%s) VALUES(%s)", lead, middle)
	_, err = DB.Exec(lead, kvs.values...)
	if err != nil {
		log.Printf("DB.Exec(\"%s\",kvs.values...) failed: %v", lead, err)
		return err
	}
	return nil
}
