/*
Gondul GO API, http receiver code
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
Package receiver is scaffolding around net/http that facilitates a
RESTful HTTP API with certain patterns implicitly enforced:

- When working on the same urls, all Methods should use the exact same
data structures. E.g.: What you PUT is the same as what you GET out
again. No cheating.

- ETag is computed for all responses.

- All responses are JSON-encoded, including error messages.

See objects/thing.go for how to use this, but the essence is:

1. Make whatever data structure you need.
2. Implement one or more of gondulapi.Getter/Putter/Poster/Deleter.
3. Use AddHandler() to register that data structure on a URL path
4. Grab lunch.

Receiver tries to do all HTTP and caching-related tasks for you, so you
don't have to.
*/
package receiver

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// AddHandler registeres an allocator/data structure with a url. The
// allocator should be a function returning an empty datastrcuture which
// implements one or more of gondulapi.Getter, Putter, Poster and Deleter
func AddHandler(url string, a Allocator) {
	if handles == nil {
		handles = make(map[string]Allocator)
	}
	handles[url] = a
}

// Allocator is used to allocate a data structure that implements at least
// one of Getter, Putter, Poster or Deleter from gondulapi.
type Allocator func() interface{}

// Start a net/http server and handle all requests registered. Never
// returns.
func Start() {
	server := http.Server{}
	serveMux := http.NewServeMux()
	server.Handler = serveMux
	for idx, h := range handles {
		log.Printf("idx: %v h: %v\n", idx, h)
		serveMux.Handle(idx, receiver{alloc: h, path: idx})
	}
	server.Addr = ":8080"
	log.WithField("address", server.Addr).Info("Starting http receiver")
	log.Fatal(server.ListenAndServe())
}
