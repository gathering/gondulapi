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

package objects

import (
	"fmt"
	"time"

	"github.com/gathering/gondulapi"
	"github.com/gathering/gondulapi/db"
	"github.com/gathering/gondulapi/receiver"
)

// Oplog is a single oplog entry. It can be created with POST, or updated
// with PUT referencing the id.
type Oplog struct {
	Id       *int
	Time     *time.Time
	Systems  *string
	Username *string
	Log      *string
}

// Oplogs is an array of oplog entries, and can only be fetched (with Get).
type Oplogs []Oplog

func init() {
	receiver.AddHandler("/oplog/", func() interface{} { return &Oplog{} })
	receiver.AddHandler("/oplog", func() interface{} { return &Oplogs{} })
}

func (o *Oplog) Get(element string) error {
	return db.Get(o, "oplog", "id", "=", element)
}

func intmatcher(element *string, i **int) error {
	if (element == nil || *element == "") && *i == nil {
		return gondulapi.Errorf(400, "The id can't be blank for a oplog entry.")
	}
	if *element == "" {
		*element = fmt.Sprintf("%d", **i)
		return nil
	}
	if *i == nil {
		var newi int
		*i = &newi
		_, err := fmt.Sscanf(*element, "%d", *i)
		return err
	}
	if *element != fmt.Sprintf("%d", **i) {
		return gondulapi.Errorf(400, "The identifier in the data and the URL don't match.")
	}
	return nil
}

func (o Oplog) Put(element string) (gondulapi.Report, error) {
	err := intmatcher(&element, &o.Id)
	if err != nil {
		return gondulapi.Report{Failed: 1}, err
	}
	return db.Upsert(o, "oplog", "id", "=", element)
}

func (o Oplog) Post() (gondulapi.Report, error) {
	return db.Insert(o, "oplog")
}

// Delete the switch
func (o Oplog) Delete(element string) (gondulapi.Report, error) {
	return db.Delete(element, "id", "oplog")
}

// Get multiple switches. Relies on s being a pointer to an array of
// structs (which it is).
func (os *Oplogs) Get(element string) error {
	return db.SelectMany(os, "oplog")
}
