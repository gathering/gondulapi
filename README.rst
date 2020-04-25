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
