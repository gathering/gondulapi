package objects

import (
	"fmt"
	"github.com/gathering/gondulapi/receiver"
	log "github.com/sirupsen/logrus"
)

/* thing is a dummy structure to illustrate the core GET/PUT/POST/DELETE
 * API. It doesn't implement a persistent storage/database connection, only
 * stores data in memory.
 */
type thing struct {
	Sysname string
}

// thinges represent the internal storage of thinges. Indexed by name.
var thinges map[string]*thing

func init() {
	thinges = make(map[string]*thing)

	// This is how we register for a url. The url is the same as used
	// for net/http. The func()... is something you can cargo-cult - it
	// is al allocation function for an empty instance of the data
	// model.
	receiver.AddHandler("/thing/", func() interface{} { return &thing{} })
}

// Get is called on GET. b will be an empty thing. Fill it out, using the
// element to determine what we're looking for. If it fails: return an
// error. Simple.
func (b *thing) Get(element string) error {
	ans, ok := thinges[element]
	if !ok{
		return fmt.Errorf("Thing %s doesn't exist", element)
	}
	*b = *ans
	return nil
}

// Put is used to store an element with an absolute URL. In our case, the
// name of the element is also (potentially) present in the data it self -
// so we do a bit of magic. Note that this should NEVER generate a random
// name.
//
// b will contain the parsed data. element will be the name of the thing.
//
// PUT is idempotent. Calling it once with a set of parameters or a hundred
// times with the same parameters should yield the same result.
func (b thing) Put(element string) ( error) {
	_, ok := thinges[element]
	if ok{
		return fmt.Errorf("Thing %s already exist", element)
	}
	if b.Sysname == "" && element != "" {
		log.Printf("Blank sysname, using url-path")
		b.Sysname = element
	}
	if element == "" && b.Sysname != "" {
		element = b.Sysname
	}
	if b.Sysname != element {
		return fmt.Errorf("Thing url path %s doesn't match json-specified name %s", element,b.Sysname)
	}

	thinges[element] = &b
	log.Printf("Put element %s, data: %v\n", element, b)
	return nil
}

// Delete is called to delete an element.
func (b thing) Delete(element string) (error) {
	delete(thinges,element)
	return nil
}
