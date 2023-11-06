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
	"time"

	"github.com/gathering/gondulapi"
	"github.com/gathering/gondulapi/db"
	"github.com/gathering/gondulapi/receiver"
	"github.com/gathering/gondulapi/types"
	log "github.com/sirupsen/logrus"
)

// Switch represents a single switch, or box. It can be updated in bulk or
// singular.
type Switch struct {
	Sysname       *string
	MgmtIP4       *types.IP  `column:"mgmt_v4_addr"`
	MgmtIP6       *types.IP  `column:"mgmt_v6_addr"`
	LastUpdated   *time.Time `column:"last_updated"`
	PollFrequency *string    `column:"poll_frequency"`
	Locked        *bool
	Deleted       *bool
	DistroName    *string `column:"distro_name"`
	DistroPhyPort *string `column:"distro_phy_port"`
	Tags          *types.Jsonb
	Community     *string
	TrafficVlan   *int `column:"traffic_vlan"`
	MgmtVlan      *int `column:"mgmt_vlan"`
	Placement     *types.Box
}

// Switches is the slice-variant of Switch, allowing us to return and parse
// an array of switches.
type Switches []Switch

func init() {
	receiver.AddHandler("/switches/", func() interface{} { return &Switch{} })
	receiver.AddHandler("/switches", func() interface{} { return &Switches{} })
}

// Get a single switch from the database and return it. db.Get is a
// convenience that returns 404 if it doesn't exist and 400 if element is
// blank.
func (s *Switch) Get(element string) (gondulapi.Report, error) {
	return db.Get(s, "switches", "sysname", "=", element)
}

func strmatcher(element *string, s **string) error {
	if (element == nil || *element == "") && (*s == nil || **s == "") {
		return gondulapi.Errorf(400, "The id can't be blank for a oplog entry.")
	}
	if *element == "" {
		*element = **s
		return nil
	}
	if *s == nil {
		var news string
		*s = &news
		**s = *element
		return nil
	}
	return nil
}

// Put will update or add a provided switch. If the name on the url and the
// one contained in the data doesn't match, the switch will be renamed from
// what's on the url to what's in the data.
func (s Switch) Put(element string) (gondulapi.Report, error) {
	err := strmatcher(&element, &s.Sysname)
	if err != nil {
		return gondulapi.Report{Failed: 1}, err
	}
	if *s.Sysname != element {
		log.Printf("Renaming switch from %s to %s", element, *s.Sysname)
	}
	return db.Upsert(element, "sysname", "switches", s)
}

// Post will either update or insert a switch entirely contained in the
// provided object. For switches, it's the same as Put without an element.
func (s Switch) Post() (gondulapi.Report, error) {
	return s.Put("")
}

// Delete the switch
func (s Switch) Delete(element string) (gondulapi.Report, error) {
	return db.Delete(element, "sysname", "switches")
}

// Get multiple switches. Relies on s being a pointer to an array of
// structs (which it is).
func (s *Switches) Get(element string) (gondulapi.Report, error) {
	return db.SelectMany(s, "switches")
}

// Post all the provided switches in bulk.
func (s Switches) Post() (gondulapi.Report, error) {
	sn := []Switch(s)
	ret := gondulapi.Report{}
	for idx := range sn {
		report, err := sn[idx].Post()
		if err != nil {
			log.WithError(err).Printf("Single-item failed, but moving on with switch-update")
			ret.Failed++
		} else {
			ret.Ok++
			ret.Affected += report.Affected
		}
	}
	return ret, nil
}
