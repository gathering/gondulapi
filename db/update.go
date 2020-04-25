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

// Package db provides convenience-functions for mapping Go-types to a
// database. It does not address a few well-known issues with database
// code and is not particularly fast, but is intended to make 90% of the
// database-work trivial, while not attempting to solve the last 10% at
// all.
//
// The convenience-functions work well if you are not regularly handling
// updates to the same content in parallel, and you do not depend on
// extreme performance for SELECT. If you are unsure if this is the case,
// I'm willing to bet five kilograms of bananas that it's not. You'd know
// if it was.
//
// You can always just use the db/DB handle directly, which is provided
// intentionally.
//
// db tries to automatically map a type to a row, both on insert/update and
// select. It does this using introspection and the official database/sql
// package's interfaces for scanning results and packing data types. So if
// your data types implement sql.Scanner and sql/driver.Value, you can use
// them directly with 0 extra boiler-plate.
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
func Update(needle interface{}, haystack string, table string, d interface{}) error {
	kvs, err := enumerate(haystack, false, d)
	if err != nil {
		log.WithError(err).Error("Update(): enumerate() failed")
		return err
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
		log.WithError(err).WithField("lead", lead).Error("Update(): EXEC failed")
		return err
	}
	return nil
}

// Insert adds the object to the table specified. It only provides the
// non-nil-pointer objects as fields, so it is up to the caller and the
// database schema to enforce default values.
func Insert(table string, d interface{}) error {
	kvs, err := enumerate("-", false, d)
	if err != nil {
		log.WithError(err).Error("Insert(): Enumerate failed")
		return err
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
		log.WithError(err).WithField("lead", lead).Error("Insert(): EXEC failed")
		return err
	}
	return nil
}

// Upsert makes database-people cringe by first checking if an element
// exists, if it does, it is updated. If it doesn't, it is inserted. This
// is NOT a transaction-safe implementation, which means: use at your own
// peril. The biggest risks are:
//
// 1. If the element is created by a third party during Upsert, the update
// will fail because we will issue an INSERT instead of UPDATE. This will
// generate an error, so can be handled by the frontend.
//
// 2. If an element is deleted by a third party during Upsert, Upsert will
// still attempt an UPDATE, which will fail silently (for now). This can be
// handled by a front-end doing a double-check, or by just assuming it
// doesn't happen often enough to be worth fixing.
func Upsert(needle interface{}, haystack string, table string, d interface{}) error {
	found, err := Exists(needle, haystack, table)
	if err != nil {
		return err
	}
	if found {
		return Update(needle, haystack, table, d)
	}
	return Insert(table, d)
}

// Delete will delete the element, and will also delete duplicates.
func Delete(needle interface{}, haystack string, table string) (err error) {
	q := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", table, haystack)
	_, err = DB.Exec(q, needle)
	if err != nil {
		log.WithError(err).WithField("query", q).Error("Delete(): Query failed")
		return
	}
	return nil
}
