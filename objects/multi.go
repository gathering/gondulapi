/*
Gondul GO API, tech-online test-objects
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

/*
 * Note that this file (multi.go) is a mix between PoC for richer
 * selector-support, and future Tech:Online backend. It will eventually be
 * moved out of the gondulapi-repo and into the tech-online repo, but is
 * developed here for convenience since it drives gondulapi updates at the
 * moment
 */

import (
	"fmt"
	"strings"
	"time"

	"github.com/gathering/gondulapi"
	"github.com/gathering/gondulapi/auth"
	"github.com/gathering/gondulapi/db"
	"github.com/gathering/gondulapi/receiver"
)

// Test is a single test(result), with all relevant descriptions. It is
// used mainly for writes, but can be Get() from as well - because why not.
type Test struct {
	Track            *string
	Station          *int
	Hash             *string       // Hash is a unique identifier for a single test, defined by the poster. Typically a sha-sum of the title. The key for a single test is, thus, track/station/hash
	Title            *string       // Short title
	Time             *time.Time    // Update time. Can be left empty on PUT/POST (updated by triggers)
	Description      *string       // Longer description for the test
	Status           *string       // Actual status-result. Should probably be OK / WARN/ FAIL or something (to be defined)
	Participant      *string       // Participant ID... somewhat legacy. Might be removed.
	Seq		 *int	       // Sorting ID
	*auth.ReadPublic `column:"-"`              // Authentication enforced for writes, not reads.
	id               []interface{} // see mkid
}

// StationTests is an array of tests associated with a single station (and
// track). It is used for reading all status entries for a single station.
// In the future, it might support Delete() to clear all status for a
// station.
type StationTests []Test

type Docstub struct {
	Family    *string
	Shortname *string
	Name      *string
	Sequence  *int
	Content   *string
	*auth.ReadPublic `column:"-"`
}

type Docs []Docstub

func init() {
	receiver.AddHandler("/test/", func() interface{} { return &Test{} })
	receiver.AddHandler("/tests/track/", func() interface{} { return &StationTests{} })
	receiver.AddHandler("/doc/family/", func() interface{} { return &Docstub{} })
	receiver.AddHandler("/doc/", func() interface{} { return &Docs{} })
}

func (ds *Docstub) Get(element string) error {
	family, shortname := "", ""
	element = strings.Replace(element, "/", " ", -1)
	_, err := fmt.Sscanf(element, "%s shortname %s", &family, &shortname)
	if err != nil {
		return gondulapi.Errorf(400, "Invalid search string, need family/%%s/shortname/%%s, got family/%s, err: %v", element, err)
	}
	return db.Get(ds, "docs", "family", "=", family, "shortname", "=", shortname)
}

func (ds Docstub) Put(element string) (gondulapi.Report, error) {
	family, shortname := "", ""
	element = strings.Replace(element, "/", " ", -1)
	_, err := fmt.Sscanf(element, "%s shortname %s", &family, &shortname)
	if err != nil {
		return gondulapi.Report{Failed: 1}, gondulapi.Errorf(400, "Invalid search string, need family/%%s/shortname/%%s, got family/%s, err: %v", element, err)
	}
	return db.Upsert(ds, "docs", "family", "=", family, "shortname", "=", shortname)
}

func (ds Docstub) Post() (gondulapi.Report, error) {
	if ds.Family == nil || *ds.Family == "" || ds.Shortname == nil || *ds.Shortname == "" {
		return gondulapi.Report{Failed: 1}, gondulapi.Errorf(400, "Need to provide Family and Shortname for doc stubs")
	}
	return db.Upsert(ds, "docs", "family", "=", ds.Family, "shortname", "=", ds.Shortname)
}

func (d *Docs) Get(element string) error {
	return db.SelectMany(d, "docs", "family", "=", element)
}

// Get an array of tests associated with a station, uses the
// url path /test/track/$TRACKID/station/$STATIONID for readability.
func (st *StationTests) Get(element string) error {
	track, station := "", 0
	element = strings.Replace(element, "/", " ", -1)
	_, err := fmt.Sscanf(element, "%s station %d", &track, &station)
	if err != nil {
		return gondulapi.Errorf(400, "Invalid search string, need track/%%s/station/%%d, got station/%s, err: %v", element, err)
	}
	return db.SelectMany(st, "results", "track", "=", track, "station", "=", station)
}

// mkid is a convenience-function to parse the URL for a test and backfill
// it into t, building t.id while we're at it which can be parsed to any
// gondulapi.db function that accepts variadic search arguments.
func (t *Test) mkid(element string) error {
	t.Track = new(string)
	t.Station = new(int)
	t.Hash = new(string)
	element = strings.Replace(element, "/", " ", -1)
	_, err := fmt.Sscanf(element, "track %s station %d hash %s", t.Track, t.Station, t.Hash)
	if err != nil {
		return gondulapi.Errorf(400, "Invalid search string, need track/%%s/station/%%d/hash/%%s, got %s, err: %v", element, err)
	}
	t.id = []interface{}{"track", "=", t.Track, "station", "=", t.Station, "hash", "=", t.Hash}
	return nil
}

// Get a single test
func (t *Test) Get(element string) error {
	if err := t.mkid(element); err != nil {
		return err
	}
	return db.Get(t, "results", t.id...)
}

// Put a single test - uses upsert: if it exists, it is updated, if it
// doesn't it is added.
func (t Test) Put(element string) (gondulapi.Report, error) {
	if err := t.mkid(element); err != nil {
		return gondulapi.Report{Failed: 1}, err
	}
	return db.Upsert(t, "results", t.id...)
}

// Post a single test - Also uses upsert, but ignores the URL and requires
// all fields to be present in the data instead.
func (t Test) Post() (gondulapi.Report, error) {
	if t.Track == nil || *t.Track == "" || t.Station == nil || *t.Station == 0 || t.Hash == nil || *t.Hash == "" {
		return gondulapi.Report{Failed: 1}, gondulapi.Errorf(400, "POST must define both track and station as non-0 values")
	}
	return db.Upsert(t, "results", "track", "=", t.Track, "station", "=", t.Station, "hash", "=", t.Hash)
}

// Delete all tests that match the url (which SHOULD be just one)
func (t Test) Delete(element string) (gondulapi.Report, error) {
	if err := t.mkid(element); err != nil {
		return gondulapi.Report{Failed: 1}, err
	}
	return db.Delete("results", t.id...)
}
