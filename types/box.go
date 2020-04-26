/*
Gondul GO API, box data type
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
	"fmt"
)

// Box contains the two opposing corners of a geometric box, matching
// Postgres' box type.
//
// Please note that Postgresql will always place the upper right corner
// first - so X1/Y1 might be X2/Y2 when you read it back out - but the
// semantics are correct.
type Box struct {
	X1 int
	X2 int
	Y1 int
	Y2 int
}

// Scan implements sql.Scan to enable gondulapi/db to read the SQL directly
// into the box.
func (b *Box) Scan(src interface{}) error {
	switch value := src.(type) {
	case string:
		_, err := fmt.Sscanf(value, "(%d,%d),(%d,%d)", &b.X1, &b.Y1, &b.X2, &b.Y2)
		return err
	case []byte:
		str := string(value)
		_, err := fmt.Sscanf(str, "(%d,%d),(%d,%d)", &b.X1, &b.Y1, &b.X2, &b.Y2)
		return err
	default:
		return fmt.Errorf("invalid box. Got type %T", src)
	}
}

// Value implements sql/driver's Value interface to provide implicit
// support for writing the value.
//
// In other words: Scan() enables SELECT and Value() allows INSERT/UPDATE.
func (b Box) Value() (driver.Value, error) {
	x := fmt.Sprintf("((%d,%d),(%d,%d))", b.X1, b.Y1, b.X2, b.Y2)
	return x, nil
}
