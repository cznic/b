// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

import (
	"math"
	"runtime/debug"
	"testing"

	"github.com/cznic/mathutil"
)

func cmp(a, b int) int {
	return a - b
}

func rng() *mathutil.FC32 {
	x, err := mathutil.NewFC32(math.MinInt32, math.MaxInt32, false)
	if err != nil {
		panic(err)
	}

	return x
}

func BenchmarkSetSeq(b *testing.B) {
	r := TreeNew(cmp)
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Set(i, i)
	}
}

func BenchmarkSetRnd(b *testing.B) {
	r := TreeNew(cmp)
	rng := rng()
	a := make([]int, b.N)
	for i := range a {
		a[i] = rng.Next()
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for _, v := range a {
		r.Set(v, 0)
	}
}

func BenchmarkGetSeq(b *testing.B) {
	r := TreeNew(cmp)
	for i := 0; i < b.N; i++ {
		r.Set(i, i)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Get(i)
	}
}

func BenchmarkGetRnd(b *testing.B) {
	r := TreeNew(cmp)
	rng := rng()
	a := make([]int, b.N)
	for i := range a {
		a[i] = rng.Next()
	}
	for _, v := range a {
		r.Set(v, 0)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for _, v := range a {
		r.Get(v)
	}
}

func BenchmarkDelSeq(b *testing.B) {
	r := TreeNew(cmp)
	for i := 0; i < b.N; i++ {
		r.Set(i, i)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Delete(i)
	}
}

func BenchmarkDelRnd(b *testing.B) {
	r := TreeNew(cmp)
	rng := rng()
	a := make([]int, b.N)
	for i := range a {
		a[i] = rng.Next()
	}
	for _, v := range a {
		r.Set(v, 0)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for _, v := range a {
		r.Delete(v)
	}
}

func BenchmarkSeekSeq(b *testing.B) {
	t := TreeNew(cmp)
	for i := 0; i < b.N; i++ {
		t.Set(i, 0)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Seek(i)
	}
}

func BenchmarkSeekRnd(b *testing.B) {
	r := TreeNew(cmp)
	rng := rng()
	a := make([]int, b.N)
	for i := range a {
		a[i] = rng.Next()
	}
	for _, v := range a {
		r.Set(v, 0)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for _, v := range a {
		r.Seek(v)
	}
}

func BenchmarkNext1e3(b *testing.B) {
	const N = 1e3
	t := TreeNew(cmp)
	for i := 0; i < N; i++ {
		t.Set(i, 0)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en, err := t.SeekFirst()
		if err != nil {
			b.Fatal(err)
		}
		n := 0
		for {
			if _, _, err = en.Next(); err != nil {
				break
			}
			n++
		}
		if n != N {
			b.Fatal(n)
		}
	}
}

func BenchmarkPrev1e3(b *testing.B) {
	const N = 1e3
	t := TreeNew(cmp)
	for i := 0; i < N; i++ {
		t.Set(i, 0)
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en, err := t.SeekLast()
		if err != nil {
			b.Fatal(err)
		}
		n := 0
		for {
			if _, _, err = en.Prev(); err != nil {
				break
			}
			n++
		}
		if n != N {
			b.Fatal(n)
		}
	}
}
