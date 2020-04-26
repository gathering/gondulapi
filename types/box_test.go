/*
Gondul GO API, box data type tests
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

package types_test

import (
	"testing"

	h "github.com/gathering/gondulapi/helper"
	"github.com/gathering/gondulapi/types"
)

func TestBox(t *testing.T) {
	box := &types.Box{}

	err := box.Scan("(0,0),(0,0)")
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, box.X1, 0)
	h.CheckEqual(t, box.X2, 0)
	h.CheckEqual(t, box.Y1, 0)
	h.CheckEqual(t, box.Y2, 0)
	ret, err := box.Value()
	h.CheckEqual(t, err, nil)
	str := ret.(string)
	h.CheckEqual(t, str, "((0,0),(0,0))")

	err = box.Scan("(10,20),(30,40)")
	h.CheckEqual(t, err, nil)
	h.CheckEqual(t, box.X1, 10)
	h.CheckEqual(t, box.X2, 30)
	h.CheckEqual(t, box.Y1, 20)
	h.CheckEqual(t, box.Y2, 40)
	ret, err = box.Value()
	h.CheckEqual(t, err, nil)
	str = ret.(string)
	h.CheckEqual(t, str, "((10,20),(30,40))")

	err = box.Scan("kjeks")
	h.CheckNotEqual(t, err, nil)
}
