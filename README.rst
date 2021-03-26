Gondul API
==========

This is the API engine that will be used for the Gondul backend in the
future. At present, this is very much a work in progress and should NOT be
used unless you talk to me directly about it first - unless you like
breaking things.

The design goals are:

1. Make it very hard to do the wrong thing.
2. Enforce/ensure HTTP RESTful best behavior.
3. Minimize boilerplate-code
4. Make prototyping fun and easy.

To achieve this, we aim that users of the Gondul API will work mainly with
organizing their own data types and how they are interconnected, and not
worry about how that is parsed to/from JSON or checked for type-safety.

The HTTP-bit is pretty small, but important. It's worth noting that users
of the api-code will *not* have access to any information about the caller.
This is a design decision - it's not your job, it's the job of the
HTTP-server (which is represented by the Gondul API engine here). If your
data types rely on data from the client, you're doing it wrong.

The database-bit can be split into a few categories as well. But mainly, it
is an attempt to make it unnecessary to write a lot of boiler-plate to get
sensible behavior. It is currently written with several known flaws, or
trade-offs, so it's not suited for large deployments, but can be considered
a POC or research.

In general, the DB engine uses introspection to figure out how to figure
out how to retrieve and save an object. The Update mechanism will only
update fields that are actually provided (if that is possible to detect!).

How to use it
-------------

Don't do it if you don't talk to me first :D

Assuming you have...

To use gondulapi, know that everything in the `objects` package can and
should be separate repos - this is what you want to provide/write yourself.
For now, what's in there is kept because it's a hassle to develop a library
in a separate repo from the code using it at the moment.

So what you want to do is make your own objects/ package and cmd/test to
kick it off.

cmd/test is easy enough: It just imports objects for side effects and
starts the receiver.

Your job is to make objects. An object is something that can be represented
by a URL. Objects are made by defining a data type and providing at least
ONE of the gondulapi.Getter / Putter / Poster / Deleter interfaces, then
letting the receiver module know about the object with
```receiver.AddHandler```::

	receiver.AddHandler("/test/", func() interface{} { return &Test{} })

This will register the Test object to the /test/ url. When the receiver
gets a request for /test/, it will call the allocation function which
allocates an empty Test struct. For write-methods, the receiver will
json-parse the body of the request for you before the appropriate
Put/Post/Delete method is called (if you implement it).

For GET, the inverse is true: The struct will remain empty, but you need to
implement the code that fills in the blanks.

This means that your data types must implement MarshalJSON and
UnmarshalJSON.

Database stuff
--------------

Since 99% of Gondul is simple database access where a URL matches a table
(or view, if you're fancy :D), `gondulapi/db` implements some fancy as
fudge introspection magic. It might be a bit overkill...

But: The idea behind `gondulapi/db` is to provide you with both the regular
database abstractions you're used to, and kick it up a notch by using
marshaling and introspection to generate the correct queries, for both
insert and get.

E.g.: If your object only has elements that sql/driver can deal with (e.g.:
uses sql.Scan and sql.Value interface), you can use db.Select() or
db.Update() without writing any SQL. It's not 100% pretty, but OK. Here's
an exmaple for getting a documentation stub::

	return db.Get(ds, "docs", "family", "=", family, "shortname", "=", shortname)

Here "ds" is a reference to the Docstub struct that will be populated,
"docs" is the table to look for, and the rest are quadruplets of
key/operator/values used in the WHERE statement.

The following is a complete implementation of POST for a documentation
stub, excluding the init() AddHandler call.

::

   type Docstub struct {
           Family    *string
           Shortname *string
           Name      *string
           Sequence  *int
           Content   *string
           *auth.ReadPublic
   }

   func (ds Docstub) Post() (gondulapi.Report, error) {
           if ds.Family == nil || *ds.Family == "" || ds.Shortname == nil || *ds.Shortname == "" {
                   return gondulapi.Report{Failed: 1}, gondulapi.Errorf(400, "Need to provide Family and Shortname for doc stubs")
           }
           return db.Upsert(ds, "docs", "family", "=", ds.Family, "shortname", "=", ds.Shortname)
   }

The write functions all return a report combined with an error. This is to
provide feedback to the user on how many items were modified/added.

Where this gets more interesting is for more complex objects or less
trivial data types. E.g.: If your data type deals with time, time.Time can
be used without worrying about converting it back and forth.

You can also implement more complex data types yourself. Some are provided
in `gondulapi/types`, where the raw value didn't implement sql.Scan or
similar. Examples are IP addresses, generic JSON, Postgres' box datatype,
and so forth.

But even if this is provided, you can still just use your own SQL as well.
But hopefully you wont feel like doing that.

Some things that are worth considering:

1. This is not meant to be highly performant. Introspection isn't
   super-expensive, and most likely you have other problems if you think this
   is your bottleneck, but yeah.
2. It's likely to change as more of Gondul starts using it.
3. You may want to consider using a view for complicated SELECT's. It all
   depends on where you want your complexity.
4. This should work for MySQL, but has only been tested for Postgres.

