package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)


type receiver struct {
	url string
}

type Input struct {
	Method	string
	Public	bool
	Data	map[string]interface{}
	URL	*url.URL
}

type Output struct {
	Data interface{}
	CacheControl string
	ETag string
}

type httpReturn struct {
	Message string
	Error	error
}

func (rcvr receiver) answer(w http.ResponseWriter, output Output, pretty bool) {
	var b []byte
	if pretty {
		b, _ = json.MarshalIndent(output.Data,"","  ")
	} else {
		b, _ = json.Marshal(output.Data)
	}
	code := 200
	w.Header().Set("ETag","KEK")
	w.WriteHeader(code)
	if code == 204 {
		return
	}

	fmt.Fprintf(w, "%s\n", b)
}

func (rcvr receiver) get(w http.ResponseWriter, r *http.Request) (Input, error) {
	var d Input
	if r.ContentLength == 0 {
		return d, fmt.Errorf("bah missing data")
	}

	b := make([]byte, r.ContentLength)

	if n, err := io.ReadFull(r.Body, b); err != nil {
		log.WithFields(log.Fields{
			"address":  r.RemoteAddr,
			"error":    err,
			"numbytes": n,
		}).Error("Read error from client")
		return d, fmt.Errorf("read failed: %v", err)
	}

	err := json.Unmarshal(b,&d.Data)
	
	if err != nil {
		d.Data = make(map[string]interface{})
		d.Data["Message"] = "bad"
	}
	d.Method = r.Method
	d.URL = r.URL

	return d, nil
}

func handle(i Input) (Output, error) {
	var o Output
	o.Data = i.Data
	o.ETag = "kjeks"
	return o,nil	
}

func (rcvr receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := rcvr.get(w, r)
	log.Printf("got %s",input.Data)
	var output Output
	if err == nil {
		output, _ = handle(input)	
	} else {
		output.Data = httpReturn{
			Message: "Input error",
			Error: err,
		}
	}
	v := input.URL.Query()
	pretty := len(v["pretty"])>0
	rcvr.answer(w, output,pretty)
}

func main() {
	server := http.Server{}
	serveMux := http.NewServeMux()
	server.Handler = serveMux
	serveMux.Handle("/", receiver{"/"})
	server.Addr =  "[::1]:8080"
	log.WithField("address", server.Addr).Info("Starting http receiver")
	log.Fatal(server.ListenAndServe())
}
