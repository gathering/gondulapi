/*
Gondul GO API, db tests
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

package db_test

import (
	"testing"

	"github.com/gathering/gondulapi/db"
	h "github.com/gathering/gondulapi/helper"
	"github.com/gathering/gondulapi/types"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.TraceLevel)
}

// system is not completely random. Having Ignored in the middle is
// important to properly test ignored fields and the accounting, which has
// been buggy in the past.
type system struct {
	Sysname   string
	Ip        *types.IP
	Ignored   *string `column:"-"`
	Vlan      *int
	Placement *types.Box
}

func TestSelectMany(t *testing.T) {
	systems := make([]system, 0)
	err := db.SelectMany(1, "1", "things", &systems)
	h.CheckNotEqual(t, err, nil)
	err = db.Connect()
	h.CheckEqual(t, err, nil)
	err = db.SelectMany(1, "1", "things", &systems)
	h.CheckEqual(t, err, nil)
	h.CheckNotEqual(t, len(systems), 0)
	t.Logf("Passed base test, got %d items back", len(systems))

	indirect := make([]*system, 0)
	err = db.SelectMany(1, "1", "things", &indirect)
	h.CheckEqual(t, err, nil)
	h.CheckNotEqual(t, len(indirect), 0)

	err = db.SelectMany(1, "2", "things", &systems)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, len(systems), 0)

	err = db.SelectMany(1, "1", "asfasf", &systems)
	h.CheckNotEqual(t, err, nil)

	err = db.SelectMany(1, "1", "things", nil)
	h.CheckNotEqual(t, err, nil)

	err = db.SelectMany(1, "1", "things", systems)
	h.CheckNotEqual(t, err, nil)

	aSystem := system{}
	err = db.SelectMany(1, "1", "things", &aSystem)
	h.CheckNotEqual(t, err, nil)
	db.DB.Close()
	db.DB = nil
}

func TestSelect(t *testing.T) {
	item := system{}
	found, err := db.Select(1, "1", "things", &item)
	h.CheckNotEqual(t, err, nil)

	err = db.Connect()
	h.CheckEqual(t, err, nil)

	found, err = db.Select(1, "1", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckNotEqual(t, found, false)

	found, err = db.Select(1, "1", "things", item)
	h.CheckNotEqual(t, err, nil)
	h.CheckNotEqual(t, found, true)

	found, err = db.Select(1, "sysnax", "things", &item)
	h.CheckNotEqual(t, err, nil)
	h.CheckNotEqual(t, found, true)

	found, err = db.Select("e1-3", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, found, true)
	h.CheckEqual(t, item.Sysname, "e1-3")
	h.CheckEqual(t, *item.Vlan, 1)
	db.DB.Close()
	db.DB = nil
}

func TestUpdate(t *testing.T) {
	item := system{}
	err := db.Connect()
	h.CheckEqual(t, err, nil)

	found, err := db.Select("e1-3", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, found, true)
	h.CheckEqual(t, *item.Vlan, 1)

	*item.Vlan = 42
	err = db.Update("e1-3", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)

	*item.Vlan = 0
	found, err = db.Select("e1-3", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, found, true)
	h.CheckEqual(t, *item.Vlan, 42)

	*item.Vlan = 1
	err = db.Update("e1-3", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)

	*item.Vlan = 0
	found, err = db.Select("e1-3", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, found, true)
	h.CheckEqual(t, *item.Vlan, 1)
	db.DB.Close()
	db.DB = nil
}

func TestInsert(t *testing.T) {
	item := system{}
	err := db.Connect()
	h.CheckEqual(t, err, nil)

	found, err := db.Select("kjeks", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, found, false)

	item.Sysname = "kjeks"
	vlan := 42
	item.Vlan = &vlan
	newip, err := types.NewIP("192.168.2.1")
	h.CheckEqual(t, err, nil)
	item.Ip = &newip
	err = db.Insert("things", &item)
	h.CheckEqual(t, err, nil)

	found, err = db.Select("kjeks", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, found, true)

	err = db.Delete("kjeks", "sysname", "things")
	h.CheckEqual(t, err, nil)

	err = db.Upsert("kjeks", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, *item.Vlan, 42)

	*item.Vlan = 8128
	err = db.Upsert("kjeks", "sysname", "things", &item)
	h.CheckEqual(t, err, nil)

	systems := make([]system, 0)
	err = db.SelectMany("kjeks", "sysname", "things", &systems)
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, len(systems), 1)
	h.CheckEqual(t, *systems[0].Vlan, 8128)

	err = db.Delete("kjeks", "sysname", "things")
	h.CheckEqual(t, err, nil)
	db.DB.Close()
	db.DB = nil
}
