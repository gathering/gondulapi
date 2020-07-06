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

/*
Package gondulapi provides the framework for building a HTTP REST-API
backed by a Postgresql Database. The package can be split into three:

1. The HTTP bit, found in receiver/. This deals with accepting HTTP
requests, parsing requests and mapping them to data types. It ensures proper division
of labour, and makes it less easy to make inconsistent APIs by enforcing that
if you GET something on an URL, PUT or POST to the same URL will accept the
exact same data type back. In other words, you can do:

	GET http://foo/blatti/x > file
	vim file // Change a field
	lwp-request -m PUT http://foo/blatti/x < file

And it will do what you expect, assuming the datastructure implements both
the Getter-interface and the Putter interface.

2. The SQL bit, found in db/. This is an attempt to use reflection to avoid
having to write database queries by hand. It is not meant to cover 100% of
all SQL access. Since it makes mapping a Go data type to an SQL table easy,
it is meant to inspire good database models, where the API can mostly just
get items back and forth.

3. "Misc" - or maybe I should say, the actual Gondul API. Which at this
moment isn't actually written. Some other bits fall under this category
though, such as config file management and logging. Not very exotic.

*/
package gondulapi

import "fmt"

// Report is an update report on write-requests. The precise meaning might
// vary, but the gist should be the same.
type Report struct {
	Affected int
	Ok       int
	Failed   int
	Error	 error `json:",omitempty"`
	Code	int `json:"-"`
}

// Getter implements Get method, which should fetch the object represented
// by the element path.
type Getter interface {
	Get(element string) error
}

// Putter is an idempotent method that requires an absolute path. It should
// (over-)write the object found at the element path.
type Putter interface {
	Put(element string) (Report, error)
}

// Poster is not necessarily idempotent, but can be. It should write the
// object provided, potentially generating a new ID for it if one isn't
// provided in the data structure itself.
type Poster interface {
	Post() (Report, error)
}

// Deleter should delete the object identified by the element. It should be
// idempotent, in that it should be safe to call it on already-deleted
// items.
type Deleter interface {
	Delete(element string) (Report, error)
}

// Errorf is a convenience-function to provide an Error data structure,
// which is essentially the same as fmt.Errorf(), but with an HTTP status
// code embedded into it which can be extracted.
func Errorf(code int, str string, v ...interface{}) Error {
	return Errori(code, fmt.Sprintf(str, v...))
}

// Errori creates an error with the given status-code, with i as the
// message. i should either be a text string or implement fmt.Stringer
func Errori(code int, i interface{}) Error {
	e := Error{
		Code:    code,
		Message: i,
	}
	return e
}

// Error is used to combine a text-based error with a HTTP error code.
type Error struct {
	Code	int `json:"-"`
	Message interface{}
}

// InternalError is provided for the common case of returning an opaque
// error that can be passed to a user.
var InternalError = Error{500, "Internal Server Error"}

// Error allows Error to implement the error interface. That's a whole lot
// of error in one sentence...
func (e Error) Error() string {
	if m, ok := e.Message.(string); ok {
		return m
	}
	if m, ok := e.Message.(fmt.Stringer); ok {
		return m.String()
	}

	return fmt.Sprintf("%v", e.Message)
}
