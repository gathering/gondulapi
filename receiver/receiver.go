package receiver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type Input struct {
	Method string
	Public bool
	Data   []byte
	URL    *url.URL
}

type Output struct {
	Data         interface{}
	Failed       bool
	CacheControl string
}

type allocator func() interface{}

type receiver struct {
	alloc allocator
	path  string
}

type Getter interface {
	Get(element string) error
}
type Putter interface {
	Put(element string) (error)
}
type Poster interface {
	Post() (error)
}
type Deleter interface {
	Delete(element string) (error)
}

var handles map[string]allocator

func (rcvr receiver) answer(w http.ResponseWriter, output Output, pretty bool) {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(output.Data, "", "  ")
	} else {
		b, err = json.Marshal(output.Data)
	}
	code := 200
	if output.Failed || err != nil {
		code = 400
	}
	if err != nil {
		log.Printf("Got error? %v", err)
	}
	w.Header().Set("ETag", "KEK")
	w.WriteHeader(code)
	if code == 204 {
		return
	}

	fmt.Fprintf(w, "%s\n", b)
}

func (rcvr receiver) get(w http.ResponseWriter, r *http.Request) (Input, error) {
	var d Input
	d.URL = r.URL
	d.Method = r.Method
	if r.ContentLength != 0 {
		d.Data = make([]byte, r.ContentLength)

		if n, err := io.ReadFull(r.Body, d.Data); err != nil {
			log.WithFields(log.Fields{
				"address":  r.RemoteAddr,
				"error":    err,
				"numbytes": n,
			}).Error("Read error from client")
			return d, fmt.Errorf("read failed: %v", err)
		}
	}

	return d, nil
}

func AddHandler(url string, a allocator) {
	if handles == nil {
		handles = make(map[string]allocator)
	}
	handles[url] = a
}

func (rcvr receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := rcvr.get(w, r)
	log.Debugf("got %s - err %v", input.Data, err)
	pretty := len(input.URL.Query()["pretty"]) > 0
	var output Output
	item := rcvr.alloc()
	if input.Method == "GET" {
		get := item.(Getter)
		err = get.Get(input.URL.Path[len(rcvr.path):])
		if err == nil {
			output.Data = get
		} else {
			output.Failed = true
			output.Data = struct{Message string;Error string}{Message: "Getting element failed",Error: fmt.Sprintf("%v",err)}
		}
	} else if input.Method == "PUT" {
		err := json.Unmarshal(input.Data, &item)
		if err != nil {
			output.Failed = true
			output.Data = struct{Message string;Error string}{Message: "Saving failed",Error: fmt.Sprintf("%v",err)}
		} else {
			put := item.(Putter)
			err = put.Put(input.URL.Path[len(rcvr.path):])
			if err != nil {
				output.Failed = true
				output.Data = struct{Message string;Error string}{Message: "Saving element failed",Error: fmt.Sprintf("%v",err)}
			} else {
				output.Data = struct{Message string}{Message: "Item stored"}
			}
		}
	} else if input.Method == "DELETE" {
		del := item.(Deleter)
		err = del.Delete(input.URL.Path[len(rcvr.path):])
		if err == nil {
			output.Data = struct{Message string}{Message: "Deleted"}
		} else {
			output.Failed = true
			output.Data = struct{Message string;Error string}{Message: "Deleting of element failed",Error: fmt.Sprintf("%v",err)}
		}
	}
	rcvr.answer(w, output, pretty)
}

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
