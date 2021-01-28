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
	//"time"
	"fmt"

	"github.com/gathering/gondulapi"
	"github.com/gathering/gondulapi/db"
	"github.com/gathering/gondulapi/receiver"
)


type Test struct {
	Track	*int
	Station	*int
	Title	*string
	Description *string
	Status	*string
	Participant *string
}

func init() {
	receiver.AddHandler("/test/", func() interface{} { return &Test{} })
}

func (t *Test) Get(element string) error {
	t.Track = new(int)
	t.Station = new(int)
	_, err := fmt.Sscanf(element, "%d/%d", t.Track, t.Station)
	if err != nil {
		return gondulapi.Errorf(400, "Invalid search string, need %%d/%%d, got %s, err: %v", element, err)
	}
	return db.Get(t, "results", "track","=",t.Track,"station","=",t.Station)
}

func (t Test) Put(element string) (gondulapi.Report, error) {
	t.Track = new(int)
	t.Station = new(int)
	_, err := fmt.Sscanf(element, "%d/%d", t.Track, t.Station)
	if err != nil {
		return gondulapi.Report{Failed:1},gondulapi.Errorf(400, "Invalid search string, need %%d/%%d, got %s, err: %v", element, err)
	}
	return db.Upsert(t, "results", "track","=",t.Track,"station","=",t.Station)
}

func (t Test) Post() (gondulapi.Report, error) {
	if t.Track == nil || *t.Track == 0 || t.Station == nil || *t.Station == 0 {
		return gondulapi.Report{Failed:1},gondulapi.Errorf(400, "POST must define both track and station as non-0 values")
	}
	return db.Upsert(t, "results", "track","=", t.Track, "station","=",t.Station)
}
func (t Test) Delete(element string) (gondulapi.Report, error) {
	t.Track = new(int)
	t.Station = new(int)
	_, err := fmt.Sscanf(element, "%d/%d", t.Track, t.Station)
	if err != nil {
		return gondulapi.Report{Failed:1},gondulapi.Errorf(400, "Invalid search string, need %%d/%%d, got %s, err: %v", element, err)
	}
	return db.Delete("results", "track","=",t.Track,"station","=",t.Station)
}
