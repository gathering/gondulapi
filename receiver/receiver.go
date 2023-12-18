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
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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
	headers	     map[string]string
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
	w.Header().Set("Content-Type", "application/json")
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
	output.code = 200
	output.headers = make(map[string]string)
	var report gondulapi.Report
	var err error
	defer func() {
		log.WithFields(log.Fields{
			"output.code": output.code,
			"output.data": output.data,
			"error":       err,
		}).Trace("Request handled")
		gerr, havegerr := err.(gondulapi.Error)
		if err != nil && report.Error == nil {
			report.Error = err
		}
		if report.Code != 0 {
			output.code = report.Code
		} else if havegerr {
			log.Tracef("During REST defered reply, we got a gondulapi.Error: %v", gerr)
			output.code = gerr.Code
		} else if report.Error != nil {
			output.code = 500
		} else {
			output.code = 200
		}
		if output.data == nil && output.code != 204 {
			output.data = report
		}
	}()
	if input.method == "GET" {
		get, ok := item.(gondulapi.Getter)
		if !ok {
			output.data = message("%s on %s failed: No such method for this path", input.method, path)
			return
		}
		report, err = get.Get(input.url.Path[len(path):])
		log.Printf("GET err;  %v", err)
		if err != nil {
			return
		}
		output.data = get
		output.headers = report.Headers
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
		report, err = put.Put(input.url.Path[len(path):])
		output.data = report
	} else if input.method == "DELETE" {
		del, ok := item.(gondulapi.Deleter)
		if !ok {
			output.data = message("%s on %s failed: No such method for this path", input.method, path)
			return
		}
		report, err = del.Delete(input.url.Path[len(path):])
		output.data = report
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
		report, err = post.Post()
		output.data = report
	}
	return
}

// checkAuth verifies authentication
func checkAuth(item interface{}, r *http.Request, rcvr receiver) (output, error) {
	auth, ok := item.(gondulapi.Auther)
	if !ok {
		return output{}, nil
	}
	var user, pass string
	if r.Header["Authorization"] == nil || len(r.Header["Authorization"]) < 1 {
		user = ""
		pass = ""
	} else {
		hdr := r.Header["Authorization"][0]
		raw := strings.Split(hdr, "Basic ")
		clean, err := base64.StdEncoding.DecodeString(raw[1])
		log.Printf("Got auth: %v err: %v", clean, err)
		up := strings.Split(string(clean), ":")
		user = up[0]
		pass = up[1]
	}

	err := auth.Auth(rcvr.path, r.URL.Path[len(rcvr.path):], r.Method, user, pass)
	if err != nil {
		o := output{}
		o.code = 401
		o.data = "damn"
		return o, err
	}
	return output{}, nil
}

// ServeHTTP implements the net/http ServeHTTP handler. It does this by
// first reading input data, then allocating a data structure specified on
// the receiver originally through AddHandler, then parses input data onto
// that data and replies. All input/output is valid JSON.
func (rcvr receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := rcvr.get(w, r)
	log.WithFields(log.Fields{
		"data": string(input.data),
		"err":  err,
	}).Trace("Got")
	pretty := len(input.url.Query()["pretty"]) > 0
	item := rcvr.alloc()
	if output, err := checkAuth(item, r, rcvr); err != nil {
		rcvr.answer(w, output, pretty)
		return
	}
	output := handle(item, input, rcvr.path)
	rcvr.answer(w, output, pretty)
}
