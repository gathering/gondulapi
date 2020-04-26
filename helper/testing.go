/*
Gondul GO API, helper-functions
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

// testing.go provides convienence-functions for tests.

// Package helper provides various minor utility-functions used throughout
// the gondul api.
package helper

import (
	"testing"
)

// CheckEqual compares a to b, ensuring they are equal.
func CheckEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a != b {
		t.Errorf("Test failed, a != b: %v != %v", a, b)
	}
}

// CheckNotEqual oh my god, really golint, this is not documenting stuff,
// it is producing noise. A function named "CheckNotEqual" should not
// require an explanation.
func CheckNotEqual(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if a == b {
		t.Errorf("Test failed, a == b: %v == %v", a, b)
	}
}
