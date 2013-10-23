// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

import (
	"fmt"
	"io"
	"math"
	"path"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/cznic/fileutil"
	"github.com/cznic/mathutil"
)

func use(...interface{}) {}

var dbg = func(s string, va ...interface{}) {
	_, fn, fl, _ := runtime.Caller(1)
	fmt.Printf("%s:%d: ", path.Base(fn), fl)
	fmt.Printf(s, va...)
	fmt.Println()
}

var caller = func(s string, va ...interface{}) {
	_, fn, fl, _ := runtime.Caller(2)
	fmt.Printf("%s:%d: ", path.Base(fn), fl)
	fmt.Printf(s, va...)
	fmt.Println()
}

func isNil(p interface{}) bool {
	switch x := p.(type) {
	case *x:
		if x == nil {
			return true
		}
	case *d:
		if x == nil {
			return true
		}
	}
	return false
}

func (t *Tree) dump() string {
	num := map[interface{}]int{}
	visited := map[interface{}]bool{}

	handle := func(p interface{}) int {
		if isNil(p) {
			return 0
		}

		if n, ok := num[p]; ok {
			return n
		}

		n := len(num) + 1
		num[p] = n
		return n
	}

	var pagedump func(interface{}, string)
	pagedump = func(p interface{}, pref string) {
		if isNil(p) || visited[p] {
			return
		}

		visited[p] = true
		switch x := p.(type) {
		case *x:
			h := handle(p)
			n := 0
			for i, v := range x.x {
				if v.ch != nil || v.sep != nil {
					n = i + 1
				}
			}
			fmt.Printf("%sX#%d n %d:%d {", pref, h, x.c, n)
			a := []interface{}{}
			for i, v := range x.x[:n] {
				a = append(a, v.ch, v.sep)
				if i != 0 {
					fmt.Printf(" ")
				}
				fmt.Printf("(C#%d D#%d)", handle(v.ch), handle(v.sep))
			}
			fmt.Printf("}\n")
			for _, p := range a {
				pagedump(p, pref+". ")
			}
		case *d:
			h := handle(p)
			n := 0
			for i, v := range x.d {
				if v.k != nil || v.v != nil {
					n = i + 1
				}
			}
			fmt.Printf("%sD#%d P#%d N#%d n %d:%d {", pref, h, handle(x.p), handle(x.n), x.c, n)
			for i, v := range x.d[:n] {
				if i != 0 {
					fmt.Printf(" ")
				}
				fmt.Printf("%v:%v", v.k, v.v)
			}
			fmt.Printf("}\n")
		}
	}

	pagedump(t.r, "")
	return ""
}

func rng() *mathutil.FC32 {
	x, err := mathutil.NewFC32(math.MinInt32/4, math.MaxInt32/4, false)
	if err != nil {
		panic(err)
	}

	return x
}

func cmp(a, b interface{}) int {
	return a.(int) - b.(int)
}

func TestGet0(t *testing.T) {
	r := TreeNew(cmp)
	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	v, ok := r.Get(42)
	if v != nil {
		t.Fatal(v)
	}

	if ok {
		t.Fatal(ok)
	}

}

func TestSetGet0(t *testing.T) {
	r := TreeNew(cmp)
	set := r.Set
	set(42, 314)
	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	v, ok := r.Get(42)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v.(int), 314; g != e {
		t.Fatal(g, e)
	}

	set(42, 278)
	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	v, ok = r.Get(42)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v.(int), 278; g != e {
		t.Fatal(g, e)
	}

	set(420, 0.5)
	if g, e := r.Len(), 2; g != e {
		t.Fatal(g, e)
	}

	v, ok = r.Get(42)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v.(int), 278; g != e {
		t.Fatal(g, e)
	}

	v, ok = r.Get(420)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v.(float64), 0.5; g != e {
		t.Fatal(g, e)
	}
}

func TestSetGet1(t *testing.T) {
	const N = 90000
	for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]int, N)
		for i := range a {
			a[i] = (i ^ x) << 1
		}
		for i, k := range a {
			set(k, k^x)
			if g, e := r.Len(), i+1; g != e {
				t.Fatal(i, g, e)
			}
		}

		for i, k := range a {
			v, ok := r.Get(k)
			if !ok {
				t.Fatal(i, k, v, ok)
			}

			if v == nil {
				t.Fatal(i, k)
			}

			if g, e := v.(int), k^x; g != e {
				t.Fatal(i, g, e)
			}

			k |= 1
			v, ok = r.Get(k)
			if ok {
				t.Fatal(i, k)
			}

			if v != nil {
				t.Fatal(i, k)
			}

		}
	}
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

func TestSetGet2(t *testing.T) {
	const N = 70000
	for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]int, N)
		rng := rng()
		for i := range a {
			a[i] = (rng.Next() ^ x) << 1
		}
		for i, k := range a {
			set(k, k^x)
			if g, e := r.Len(), i+1; g != e {
				t.Fatal(i, x, g, e)
			}
		}

		for i, k := range a {
			v, ok := r.Get(k)
			if !ok {
				t.Fatal(i, k, v, ok)
			}

			if v == nil {
				t.Fatal(i, k)
			}

			if g, e := v.(int), k^x; g != e {
				t.Fatal(i, g, e)
			}

			k |= 1
			v, ok = r.Get(k)
			if ok {
				t.Fatal(i, k)
			}

			if v != nil {
				t.Fatal(i, k)
			}

		}
	}
}

func TestSetGet3(t *testing.T) {
	r := TreeNew(cmp)
	set := r.Set
	var i int
	for i = 0; ; i++ {
		set(i, -i)
		if _, ok := r.r.(*x); ok {
			break
		}
	}
	for j := 0; j <= i; j++ {
		set(j, j)
	}

	for j := 0; j <= i; j++ {
		v, ok := r.Get(j)
		if !ok {
			t.Fatal(j)
		}

		if v == nil {
			t.Fatal(j)
		}

		if g, e := v.(int), j; g != e {
			t.Fatal(g, e)
		}
	}
}

func TestDelete0(t *testing.T) {
	r := TreeNew(cmp)
	if ok := r.Delete(0); ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	r.Set(0, 0)
	if ok := r.Delete(1); ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(0); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(0); ok {
		t.Fatal(ok)
	}

	r.Set(0, 0)
	r.Set(1, 1)
	if ok := r.Delete(1); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(1); ok {
		t.Fatal(ok)
	}

	if ok := r.Delete(0); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(0); ok {
		t.Fatal(ok)
	}

	r.Set(0, 0)
	r.Set(1, 1)
	if ok := r.Delete(0); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(0); ok {
		t.Fatal(ok)
	}

	if ok := r.Delete(1); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(1); ok {
		t.Fatal(ok)
	}
}

func TestDelete1(t *testing.T) {
	const N = 100000
	for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]int, N)
		for i := range a {
			a[i] = (i ^ x) << 1
		}
		for _, k := range a {
			set(k, 0)
		}

		for i, k := range a {
			ok := r.Delete(k)
			if !ok {
				t.Fatal(i, x, k)
			}

			if g, e := r.Len(), N-i-1; g != e {
				t.Fatal(i, g, e)
			}
		}
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

func TestDelete2(t *testing.T) {
	const N = 80000
	for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]int, N)
		rng := rng()
		for i := range a {
			a[i] = (rng.Next() ^ x) << 1
		}
		for _, k := range a {
			set(k, 0)
		}

		for i, k := range a {
			ok := r.Delete(k)
			if !ok {
				t.Fatal(i, x, k)
			}

			if g, e := r.Len(), N-i-1; g != e {
				t.Fatal(i, g, e)
			}
		}
	}
}

func TestEnumeratorNext(t *testing.T) {
	// seeking within 3 keys: 10, 20, 30
	table := []struct {
		k    int
		hit  bool
		keys []int
	}{
		{5, false, []int{10, 20, 30}},
		{10, true, []int{10, 20, 30}},
		{15, false, []int{20, 30}},
		{20, true, []int{20, 30}},
		{25, false, []int{30}},
		{30, true, []int{30}},
		{35, false, []int{}},
	}

	for i, test := range table {
		up := test.keys
		r := TreeNew(cmp)

		r.Set(10, 100)
		r.Set(20, 200)
		r.Set(30, 300)

		for verChange := 0; verChange < 16; verChange++ {
			//t.Logf("Seek %d", test.k)
			en, hit := r.Seek(test.k)

			if g, e := hit, test.hit; g != e {
				t.Fatal(i, g, e)
			}

			j := 0
			for {
				if verChange&(1<<uint(j)) != 0 {
					//t.Log("version change")
					r.Set(20, 200)
				}

				k, v, err := en.Next()
				if err != nil {
					if !fileutil.IsEOF(err) {
						t.Fatal(i, err)
					}

					break
				}

				//t.Logf("Next -> %v: %v", k, v)
				if j >= len(up) {
					t.Fatal(i, j, verChange)
				}

				if g, e := k.(int), up[j]; g != e {
					t.Fatal(i, j, verChange, g, e)
				}

				if g, e := v.(int), 10*up[j]; g != e {
					t.Fatal(i, g, e)
				}

				j++

			}

			if g, e := j, len(up); g != e {
				t.Fatal(i, j, g, e)
			}
		}

	}
}

func TestEnumeratorPrev(t *testing.T) {
	// seeking within 3 keys: 10, 20, 30
	table := []struct {
		k    int
		hit  bool
		keys []int
	}{
		{5, false, []int{10}},
		{10, true, []int{10}},
		{15, false, []int{20, 10}},
		{20, true, []int{20, 10}},
		{25, false, []int{30, 20, 10}},
		{30, true, []int{30, 20, 10}},
		{35, false, []int{}},
	}

	for i, test := range table {
		dn := test.keys
		r := TreeNew(cmp)

		r.Set(10, 100)
		r.Set(20, 200)
		r.Set(30, 300)

		for verChange := 0; verChange < 16; verChange++ {
			//t.Logf("Seek %d", test.k)
			en, hit := r.Seek(test.k)

			if g, e := hit, test.hit; g != e {
				t.Fatal(i, g, e)
			}

			j := 0
			for {
				if verChange&(1<<uint(j)) != 0 {
					//t.Log("version change")
					r.Set(20, 200)
				}

				k, v, err := en.Prev()
				if err != nil {
					if !fileutil.IsEOF(err) {
						t.Fatal(i, err)
					}

					break
				}

				//t.Logf("Prev -> %v: %v", k, v)
				if j >= len(dn) {
					t.Fatal(i, j, verChange)
				}

				if g, e := k.(int), dn[j]; g != e {
					t.Fatal(i, j, verChange, g, e)
				}

				if g, e := v.(int), 10*dn[j]; g != e {
					t.Fatal(i, g, e)
				}

				j++

			}

			if g, e := j, len(dn); g != e {
				t.Fatal(i, j, g, e)
			}
		}

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

func TestSeekFirst0(t *testing.T) {
	b := TreeNew(cmp)
	_, err := b.SeekFirst()
	if g, e := err, io.EOF; g != e {
		t.Fatal(g, e)
	}
}

func TestSeekFirst1(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(1, 10)
	en, err := b.SeekFirst()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Next()
	if k != 1 || v != 10 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekFirst2(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(1, 10)
	b.Set(2, 20)
	en, err := b.SeekFirst()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Next()
	if k != 1 || v != 10 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if k != 2 || v != 20 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekFirst3(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(2, 20)
	b.Set(3, 30)
	b.Set(1, 10)
	en, err := b.SeekFirst()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Next()
	if k != 1 || v != 10 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if k != 2 || v != 20 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if k != 3 || v != 30 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekLast0(t *testing.T) {
	b := TreeNew(cmp)
	_, err := b.SeekLast()
	if g, e := err, io.EOF; g != e {
		t.Fatal(g, e)
	}
}

func TestSeekLast1(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(1, 10)
	en, err := b.SeekLast()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Prev()
	if k != 1 || v != 10 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekLast2(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(1, 10)
	b.Set(2, 20)
	en, err := b.SeekLast()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Prev()
	if k != 2 || v != 20 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if k != 1 || v != 10 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekLast3(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(2, 20)
	b.Set(3, 30)
	b.Set(1, 10)
	en, err := b.SeekLast()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Prev()
	if k != 3 || v != 30 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if k != 2 || v != 20 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if k != 1 || v != 10 || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if err == nil {
		t.Fatal(k, v, err)
	}
}
