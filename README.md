# go-ttlcache

go-ttlcache is a library implementing an in-memory cache with TTLs.

Why yet another TTL map library? Compared with the others, this library:

* Has 0 dependencies outside of the standard library.
* Does not use any goroutines whatsoever.
* Expires items on write, and optimizes for fast reads.
