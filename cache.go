// Copyright © Franklin "Snaipe" Mathieu <me@snai.pe>, et al.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package ttlcache

import (
	"container/heap"
	"sync"
	"time"
)

// Cache is an implementation of an in-memory cache using TTLs. It only expires
// items on write, which means that it is possible for a value to survive
// past the expiration time that it was inserted with.
type Cache[K comparable, V any] struct {
	// OnExpire gets called whenever a key expires from the cache.
	OnExpire func(key K, value V)

	cache      map[K]*cacheBucket[K, V]
	expireList expireList[K, V]
	mux        sync.RWMutex
}

func New[K comparable, V any]() *Cache[K, V] {
	return &Cache[K, V]{
		cache: make(map[K]*cacheBucket[K, V]),
	}
}

// Set assigns the specified value to the specified key in the cache, with
// an expiration of ttl.
func (cache *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	bucket, ok := cache.cache[key]
	if !ok {
		cache.flush()

		bucket = &cacheBucket[K, V]{
			key:    key,
			idx:    cache.expireList.Len(),
		}
		cache.expireList.Push(bucket)
		cache.cache[key] = bucket
	}

	bucket.val = value
	bucket.expiry = time.Now().Add(ttl)
	heap.Fix(&cache.expireList, bucket.idx)
}

// Get retrieves the value in the cache for the specified key if it exists,
// as well as whether the value was found.
func (cache *Cache[K, V]) Get(key K) (value V, found bool) {
	cache.mux.RLock()
	defer cache.mux.RUnlock()

	bucket, found := cache.cache[key]
	if found {
		value = bucket.val
	}
	return value, found
}

// Expire expires the value associated with the specified key, if any.
func (cache *Cache[K, V]) Expire(key K) {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	bucket, found := cache.cache[key]
	if found {
		cache.delete(bucket)
	}
}

// Flush removes all expired keys from the cache.
func (cache *Cache[K, V]) Flush() {
	cache.mux.Lock()
	defer cache.mux.Unlock()

	cache.flush()
}

func (cache *Cache[K, V]) flush() {
	now := time.Now()
	for {
		bucket, ok := cache.expireList.Peek()
		if !ok || bucket.expiry.After(now) {
			break
		}
		cache.delete(bucket)
	}
}

func (cache *Cache[K, V]) delete(bucket *cacheBucket[K, V]) {
	delete(cache.cache, bucket.key)
	heap.Remove(&cache.expireList, bucket.idx)
	if onExpire := cache.OnExpire; onExpire != nil {
		onExpire(bucket.key, bucket.val)
	}
}

type cacheBucket[K, V any] struct {
	expiry time.Time
	idx    int // cache buckets know their position in the expire list
	key    K
	val    V
}

type expireList[K, V any] struct {
	elts []*cacheBucket[K, V]
}

func (c *expireList[K, V]) Peek() (*cacheBucket[K, V], bool) {
	if len(c.elts) > 0 {
		return c.elts[0], true
	}
	return nil, false
}

// expireList must implement sort.Interface and container/heap.Interface

func (l *expireList[K, V]) Len() int {
	return len(l.elts)
}

func (l *expireList[K, V]) Less(i, j int) bool {
	return l.elts[i].expiry.Before(l.elts[j].expiry)
}

func (l *expireList[K, V]) Swap(i, j int) {
	l.elts[i], l.elts[j] = l.elts[j], l.elts[i]
	// Fix the expire list indices
	l.elts[i].idx, l.elts[j].idx = i, j
}

func (l *expireList[K, V]) Push(x any) {
	l.elts = append(l.elts, x.(*cacheBucket[K, V]))
}

func (l *expireList[K, V]) Pop() (val any) {
	val = l.elts[len(l.elts)-1]
	l.elts[len(l.elts)-1] = nil // don't keep referencing the item
	l.elts = l.elts[:len(l.elts)-1]
	return val
}
