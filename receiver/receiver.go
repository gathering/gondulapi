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

package receiver

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gathering/gondulapi"

	log "github.com/sirupsen/logrus"
)

var handles map[string]Allocator

type input struct {
	method string
	public bool
	data   []byte
	url    *url.URL
}

type output struct {
	code         int
	data         interface{}
	failed       bool
	cachecontrol string
}

type receiver struct {
	alloc Allocator
	path  string
}

// answer replies to a HTTP request with the provided output, optionally
// formatting the output prettily. It also calculates an ETag.
func (rcvr receiver) answer(w http.ResponseWriter, output output, pretty bool) {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(output.data, "", "  ")
	} else {
		b, err = json.Marshal(output.data)
	}
	code := output.code
	if err != nil {
		log.Printf("Json marshal error: %v", err)
		b = []byte(`{"Message": "JSON marshal error. Very weird."}`)
		code = 500
	}
	etagraw := sha256.Sum256(b)
	etagstr := hex.EncodeToString(etagraw[:])
	w.Header().Set("ETag", etagstr)
	w.WriteHeader(code)
	if code == 204 {
		return
	}

	fmt.Fprintf(w, "%s\n", b)
}

// get is a badly named function in the context of HTTP since what it
// really does is just read the body of a HTTP request. In my defence, it
// used to do more. But what have it done for me lately?!
func (rcvr receiver) get(w http.ResponseWriter, r *http.Request) (input, error) {
	var input input
	input.url = r.URL
	input.method = r.Method
	log.WithFields(log.Fields{
		"url":     r.URL,
		"method":  r.Method,
		"address": r.RemoteAddr,
	}).Infof("Request")

	if r.ContentLength != 0 {
		input.data = make([]byte, r.ContentLength)

		if n, err := io.ReadFull(r.Body, input.data); err != nil {
			log.WithFields(log.Fields{
				"address":  r.RemoteAddr,
				"error":    err,
				"numbytes": n,
			}).Error("Read error from client")
			return input, fmt.Errorf("read failed: %v", err)
		}
	}

	return input, nil
}

// message is a convenience function
func message(str string, v ...interface{}) (m struct {
	Message string
	Error   string `json:",omitempty"`
}) {
	m.Message = fmt.Sprintf(str, v...)
	return
}

// handle figures out what Method the input has, casts item to the correct
// interface and calls the relevant function, if any, for that data. For
// PUT and POST it also parses the input data.
func handle(item interface{}, input input, path string) (output output) {
	output.failed = true
	output.code = 200
	var err error
	defer func() {
		if output.failed != true {
			if output.data == nil {
				output.data = message("%s on %s successful", input.method, path)
			}
			return
		}
		code := 400
		if gerr, ok := err.(gondulapi.Error); ok {
			code = gerr.Code
		}
		output.code = code
		if output.data == nil {
			m := message("%s on %s failed", input.method, path)
			if err != nil {
				m.Error = fmt.Sprintf("%v", err)
			}
			output.data = m
		}
	}()
	if input.method == "GET" {
		get, ok := item.(gondulapi.Getter)
		if !ok {
			output.data = message("%s on %s failed: No such method for this path", input.method, path)
			return
		}
		err = get.Get(input.url.Path[len(path):])
		if err == nil {
			output.data = get
			output.failed = false
		}
	} else if input.method == "PUT" {
		err = json.Unmarshal(input.data, &item)
		if err != nil {
			return
		}
		put, ok := item.(gondulapi.Putter)
		if !ok {
			output.data = message("%s on %s failed: No such method for this path", input.method, path)
			return
		}
		err = put.Put(input.url.Path[len(path):])
		if err != nil {
			return
		}
		output.failed = false
	} else if input.method == "DELETE" {
		del, ok := item.(gondulapi.Deleter)
		if !ok {
			output.data = message("%s on %s failed: No such method for this path", input.method, path)
			return
		}
		err = del.Delete(input.url.Path[len(path):])
		if err != nil {
			return
		}
		output.failed = false
	} else if input.method == "POST" {
		err = json.Unmarshal(input.data, &item)
		if err != nil {
			return
		}
		post, ok := item.(gondulapi.Poster)
		if !ok {
			output.data = message("%s on %s failed: No such method for this path", input.method, path)
			return
		}
		err = post.Post()
		if err != nil {
			return
		}
		output.failed = false
	}
	return
}

// ServeHTTP implements the net/http ServeHTTP handler. It does this by
// first reading input data, then allocating a data structure specified on
// the receiver originally through AddHandler, then parses input data onto
// that data and replies. All input/output is valid JSON.
func (rcvr receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := rcvr.get(w, r)
	log.Debugf("got %s - err %v", input.data, err)
	pretty := len(input.url.Query()["pretty"]) > 0
	item := rcvr.alloc()
	output := handle(item, input, rcvr.path)
	rcvr.answer(w, output, pretty)
}
