package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)


type receiver struct {}

type Input struct {
	Method	string
	Public	bool
	Data	map[string]interface{}
	URL	*url.URL
}

type Output struct {
	Data interface{}
	Failed bool
	CacheControl string
	ETag string
}

func (rcvr receiver) answer(w http.ResponseWriter, output Output, pretty bool) {
	var b []byte
	var err error
	if pretty {
		b, err = json.MarshalIndent(output.Data,"","  ")
	} else {
		b, err = json.Marshal(output.Data)
	}
	code := 200
	if output.Failed  || err != nil{
		code = 400
	}
	if err != nil {
		log.Printf("Got error? %v", err)
	}
	w.Header().Set("ETag","KEK")
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
	if r.ContentLength == 0 {
		return d, fmt.Errorf("No input-data at all.")
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
		return d, fmt.Errorf("Failed to parse json: %w", err)
	}

	return d, nil
}

type box struct {
	Sysname string
}

func (b *box) Get(element string) (Output, error) {
	var o Output
	o.Data = "hi"
	o.ETag = "kjeks"
	return o,nil	
}
func handle(i Input) (Output, error) {
	var o Output
	type server struct {
		Sysname string
	}
	_, ok := o.Data.(server)
	if ok {
		fmt.Printf("yai")
	}
	o.Data = i.Data
	o.ETag = "kjeks"
	return o,nil	
}

func (i Input) handleError(err error) (output Output) {
	output.Failed = true
	output.Data = struct {
		Message string
		Error string
	}{
		Message: "Input error",
		Error: err.Error(),
	}
	log.Printf("%s %v: Failed to parse: %v", i.Method, i.URL, err)
	return output
}

func (rcvr receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	input, err := rcvr.get(w, r)
	log.Debugf("got %s - err %v",input.Data,err)
	pretty := len(input.URL.Query()["pretty"])>0
	var output Output
	if err == nil {
		output, _ = handle(input)	
	} else {
		output = input.handleError(err)
	}
	if input.URL.Path[0:len("/switch/")] == "/switch/" {
		fmt.Printf("hei")
		b := box{}
		output,_ = b.Get(input.URL.Path[len("/switch"):])
	} else {
		fmt.Printf("kek: %v", input.URL.Path)
	}
	rcvr.answer(w, output,pretty)
}

func main() {
	server := http.Server{}
	serveMux := http.NewServeMux()
	server.Handler = serveMux
	serveMux.Handle("/", receiver{})
	server.Addr =  "[::1]:8080"
	log.WithField("address", server.Addr).Info("Starting http receiver")
	log.Fatal(server.ListenAndServe())
}
