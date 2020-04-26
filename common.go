/*
Gondul GO API
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

package gondulapi

import "fmt"

// Getter implements Get method, which should fetch the object represented
// by the element path.
type Getter interface {
	Get(element string) error
}

// Putter is an idempotent method that requires an absolute path. It should
// (over-)write the object found at the element path.
type Putter interface {
	Put(element string) error
}

// Poster is not necessarily idempotent, but can be. It should write the
// object provided, potentially generating a new ID for it if one isn't
// provided in the data structure itself.
type Poster interface {
	Post() error
}

// Deleter should delete the object identified by the element. It should be
// idempotent, in that it should be safe to call it on already-deleted
// items.
type Deleter interface {
	Delete(element string) error
}

// Errorf is a convenience-function to provide an Error data structure,
// which is essentially the same as fmt.Errorf(), but with an HTTP status
// code embedded into it which can be extracted.
func Errorf(code int, str string, v ...interface{}) Error {
	e := Error{
		Code:    code,
		Message: fmt.Sprintf(str, v...),
	}
	return e
}

// Error is used to combine a text-based error with a HTTP error code.
type Error struct {
	Code    int
	Message string
}

// InternalError is provided for the common case of returning an opaque
// error that can be passed to a user.
var InternalError = Error{500, "Internal Server Error"}

// Error allows Error to implement the error interface. That's a whole lot
// of error in one sentence...
func (e Error) Error() string {
	return e.Message
}
