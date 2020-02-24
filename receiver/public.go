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
	server.Addr = "[::1]:8080"
	log.WithField("address", server.Addr).Info("Starting http receiver")
	log.Fatal(server.ListenAndServe())
}
