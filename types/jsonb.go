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

package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Jsonb is a wrapper around an empty interface to enable adding methods to
// it. The wrapping is hidden both from SQL and JSON through a ton of
// dummy-methods.
type Jsonb struct {
	Data interface{}
}

// Scan is used by database/sql's Scan to parse bytes
func (j *Jsonb) Scan(src interface{}) error {
	switch value := src.(type) {
	case string:
		return json.Unmarshal([]byte(value), &j.Data)
	case []byte:
		return json.Unmarshal(value, &j.Data)
	default:
		return fmt.Errorf("invalid jsonb")
	}
}

// String JSON marshals the data and returns it. AS A STRING.
func (j Jsonb) String() string {
	b, err := json.Marshal(j.Data)
	if err != nil {
		return ""
	}
	return string(b)
}

// Value wraps String() for the sake of sql.
func (j Jsonb) Value() (driver.Value, error) {
	return j.String(), nil
}

// MarshalText returns a byte-string of the string representation of the
// object and this documentation, I'm sure, is very helpful.
func (j Jsonb) MarshalText() ([]byte, error) {
	return []byte(j.String()), nil
}

// MarshalJSON qwrasfaksfjasklfjasf
func (j *Jsonb) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Data)
}

// UnmarshalJSON is here just to make lint-software think this is a
// sensible way of writing documentation.
func (j *Jsonb) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &j.Data)
}

// UnmarshalText un marshals the text. Weird.
func (j *Jsonb) UnmarshalText(b []byte) error {
	return json.Unmarshal(b, &j.Data)
}
