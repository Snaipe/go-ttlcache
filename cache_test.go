// Copyright © Franklin "Snaipe" Mathieu <me@snai.pe>, et al.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

package ttlcache

import (
	"testing"
	"time"
	"math/rand"
)

func TestCache(t *testing.T) {
	stringCache := New[string, string]()
	stringCache.Set("foo", "1", 1*time.Hour)
	stringCache.Set("bar", "2", 1*time.Nanosecond)
	stringCache.Set("baz", "3", 1*time.Hour)

	foo, ok := stringCache.Get("foo")
	if !ok {
		t.Fatalf("expected key foo to be in cache, but it was not (cache: %v)", stringCache)
	}
	if foo != "1" {
		t.Fatalf("expected key foo to have value 1, but got %v", foo)
	}

	_, ok = stringCache.Get("bar")
	if ok {
		t.Fatal("expected key bar to have expired, but it was still present")
	}

	stringCache.Expire("foo")
	_, ok = stringCache.Get("foo")
	if ok {
		t.Fatal("expected key foo to have expired, but it was still present")
	}
}

func BenchmarkCache(b *testing.B) {
	b.Run("set", func (b *testing.B) {
		c := New[int, int]()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			c.Set(i, i, time.Nanosecond)
		}
	})

	b.Run("touch", func (b *testing.B) {
		c := New[int, int]()
		for i := 0; i < b.N; i++ {
			c.Set(i, i, time.Hour)
		}

		indices := make([]int, b.N)
		for i := 0; i < b.N; i++ {
			indices[i] = i
		}
		rand.Shuffle(len(indices), func(i, j int) {
			indices[i], indices[j] = indices[j], indices[i]
		})
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			c.Set(indices[i], i, time.Hour)
		}
	})

	b.Run("expire", func (b *testing.B) {
		c := New[int, int]()
		for i := 0; i < b.N; i++ {
			c.Set(i, i, time.Hour)
		}

		indices := make([]int, b.N)
		for i := 0; i < b.N; i++ {
			indices[i] = i
		}
		rand.Shuffle(len(indices), func(i, j int) {
			indices[i], indices[j] = indices[j], indices[i]
		})
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			c.Expire(indices[i])
		}
	})
}
