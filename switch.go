package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

type box struct {
	Sysname string
}

var boxes map[string]*box

func init() {
	boxes = make(map[string]*box)
	AddHandler("/switch/", func() interface{} { return &box{} })
}

func (b *box) Get(element string) error {
	ans, ok := boxes[element]
	if !ok{
		return fmt.Errorf("Switch %s doesn't exist", element)
	}
	*b = *ans
	return nil
}

func (b box) Put(element string) ( error) {
	_, ok := boxes[element]
	if ok{
		return fmt.Errorf("Switch %s already exist", element)
	}
	if b.Sysname == "" && element != "" {
		log.Printf("Blank sysname, using url-path")
		b.Sysname = element
	}
	if element == "" && b.Sysname != "" {
		element = b.Sysname
	}
	if b.Sysname != element {
		return fmt.Errorf("Switch url path %s doesn't match json-specified name %s", element,b.Sysname)
	}

	boxes[element] = &b
	log.Printf("Put element %s, data: %v\n", element, b)
	return nil
}

func (b box) Delete(element string) (error) {
	delete(boxes,element)
	return nil
}
