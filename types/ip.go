/*
Gondul GO API, data types
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

// Package types provides a set of common data types that implement the
// relevant interfaces for use within gondulapi.
//
// In general, that means implementing JSON marshaling/unmarshaling and
// database/sql's Scanner and driver.Value interface for SQL operations.
//
// Any data type that implement JSON marshaling and can be handled by SQL
// can be used without any further scaffolding. So these data types are
// just a bit outside of the norm.
package types

import (
	"database/sql/driver"
	"fmt"
	"net"
)

// IP represent an IP and an optional netmask/size. It is provided because
// while net/IP provides much of what is needed for marshalling to/from
// text and json, it doesn't provide a convenient netmask. And it doesn't
// implement database/sql's Scanner either. This type implements both
// MarshalText/UnmarshalText and Scan(), and to top it off: String(). this
// should ensure maximum usability.
type IP struct {
	IP   net.IP
	Mask int
}

// Scan is used by database/sql's Scan to parse bytes to an IP structure.
func (i *IP) Scan(src interface{}) error {
	switch value := src.(type) {
	case string:
		return i.UnmarshalText([]byte(value))
	case []byte:
		return i.UnmarshalText(value)
	default:
		return fmt.Errorf("invalid IP")
	}
}

// String returns a ip address in string format, with an optional /mask if
// the IP has a non-zero mask.
func (i *IP) String() string {
	if i.Mask != 0 {
		return fmt.Sprintf("%s/%d", i.IP.String(), i.Mask)
	}
	return i.IP.String()
}

// Value implements sql/driver's Value interface to provide implicit
// support for writing the value.
func (i IP) Value() (driver.Value, error) {
	return i.String(), nil
}

// MarshalText returns a text version of an ip using String, it is used by
// various encoders, including the JSON encoder.
func (i IP) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText parses a byte string onto an IP structure. It is used by
// various encoders, including the JSON encoder.
func (i *IP) UnmarshalText(b []byte) error {
	s := string(b)
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		ip = net.ParseIP(s)
		i.Mask = 0
		if ip == nil {
			return err
		}
	} else {
		i.Mask, _ = ipnet.Mask.Size()
	}
	i.IP = ip
	return nil
}

// NewIP parses the text string and returns it as an IP data structure
func NewIP(src string) (IP, error) {
	i := IP{}
	err := i.UnmarshalText([]byte(src))
	return i, err
}
