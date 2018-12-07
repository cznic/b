// Copyright 2014 The b Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package b

import (
	"io"
	"math"
	"strconv"
	"runtime/debug"
	"testing"

	"github.com/cznic/fileutil"
	"github.com/cznic/mathutil"
)

func rng() *mathutil.FC32 {
	x, err := mathutil.NewFC32(math.MinInt32/4, math.MaxInt32/4, false)
	if err != nil {
		panic(err)
	}

	return x
}

func cmp(a, b uint64) int64 {
	return int64(a) - int64(b)
}

func TestGet0(t *testing.T) {
	r := TreeNew(cmp)
	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	_, ok := r.Get(uint64(42))
	if ok {
		t.Fatal(ok)
	}

}


const key1 uint64 = 42
const key2 uint64 = 420
const value1 string = "314"
const value2 string = "278"
const value3 string = "50"

func TestSetGet0(t *testing.T) {
	r := TreeNew(cmp)
	set := r.Set
	set(key1, value1)
	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	v, ok := r.Get(key1)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v, value1; g != e {
		t.Fatal(g, e)
	}

	set(key1, value2)
	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	v, ok = r.Get(key1)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v, value2; g != e {
		t.Fatal(g, e)
	}

	set(key2, value3)
	if g, e := r.Len(), 2; g != e {
		t.Fatal(g, e)
	}

	v, ok = r.Get(key1)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v, value2; g != e {
		t.Fatal(g, e)
	}

	v, ok = r.Get(key2)
	if !ok {
		t.Fatal(ok)
	}

	if g, e := v, value3; g != e {
		t.Fatal(g, e)
	}
}

var keylist = []uint64{10, 20, 30, 12, 23, 31, 03, 02, 01, 33}
var valuelist = []string{"10", "20", "30", "12", "23", "31", "03", "02", "01", "33"}

func TestSetGet1(t *testing.T) {
	const N = 90000
	//for _, x := range []uint64{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
	for _, x := range []uint64{0, 1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]uint64, N)
		for i := range a {
			a[i] = (uint64(i) ^ x) << 1
		}
		for i, k := range a {
			set(k, valuelist[k%uint64(len(keylist))])
			if g, e := r.Len(), i+1; g != e {
				t.Fatal(i, g, e)
			}
		}

		for i, k := range a {
			v, ok := r.Get(k)
			if !ok {
				t.Fatal(i, k, v, ok)
			}

			if g, e := v, valuelist[k%uint64(len(keylist))]; g != e {
				t.Fatal(i, g, e)
			}

			k |= 1
			_, ok = r.Get(k)
			if ok {
				t.Fatal(i, k)
			}

		}
	}
}

func BenchmarkSetSeq1e3(b *testing.B) {
	benchmarkSetSeq(b, 1e3)
}

func BenchmarkSetSeq1e4(b *testing.B) {
	benchmarkSetSeq(b, 1e4)
}

func BenchmarkSetSeq1e5(b *testing.B) {
	benchmarkSetSeq(b, 1e5)
}

func BenchmarkSetSeq1e6(b *testing.B) {
	benchmarkSetSeq(b, 1e6)
}

func benchmarkSetSeq(b *testing.B, n int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := TreeNew(cmp)
		debug.FreeOSMemory()
		b.StartTimer()
		for j := 0; j < n; j++ {
			r.Set(uint64(j), strconv.FormatInt(int64(j), 10))
		}
		b.StopTimer()
		r.Close()
	}
	b.StopTimer()
}

func BenchmarkGetSeq1e3(b *testing.B) {
	benchmarkGetSeq(b, 1e3)
}

func BenchmarkGetSeq1e4(b *testing.B) {
	benchmarkGetSeq(b, 1e4)
}

func BenchmarkGetSeq1e5(b *testing.B) {
	benchmarkGetSeq(b, 1e5)
}

func BenchmarkGetSeq1e6(b *testing.B) {
	benchmarkGetSeq(b, 1e6)
}

func benchmarkGetSeq(b *testing.B, n int) {
	r := TreeNew(cmp)
	for i := 0; i < n; i++ {
		r.Set(uint64(i), strconv.FormatInt(int64(i), 10))
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			r.Get(uint64(j))
		}
	}
	b.StopTimer()
	r.Close()
}

func BenchmarkSetRnd1e3(b *testing.B) {
	benchmarkSetRnd(b, 1e3)
}

func BenchmarkSetRnd1e4(b *testing.B) {
	benchmarkSetRnd(b, 1e4)
}

func BenchmarkSetRnd1e5(b *testing.B) {
	benchmarkSetRnd(b, 1e5)
}

func BenchmarkSetRnd1e6(b *testing.B) {
	benchmarkSetRnd(b, 1e6)
}

func benchmarkSetRnd(b *testing.B, n int) {
	rng := rng()
	a := make([]int, n)
	for i := range a {
		a[i] = rng.Next()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := TreeNew(cmp)
		debug.FreeOSMemory()
		b.StartTimer()
		for _, v := range a {
			r.Set(uint64(v), "0")
		}
		b.StopTimer()
		r.Close()
	}
	b.StopTimer()
}

func BenchmarkGetRnd1e3(b *testing.B) {
	benchmarkGetRnd(b, 1e3)
}

func BenchmarkGetRnd1e4(b *testing.B) {
	benchmarkGetRnd(b, 1e4)
}

func BenchmarkGetRnd1e5(b *testing.B) {
	benchmarkGetRnd(b, 1e5)
}

func BenchmarkGetRnd1e6(b *testing.B) {
	benchmarkGetRnd(b, 1e6)
}

func benchmarkGetRnd(b *testing.B, n int) {
	r := TreeNew(cmp)
	rng := rng()
	a := make([]uint64, n)
	for i := range a {
		a[i] = uint64(rng.Next())
	}
	for _, v := range a {
		r.Set(uint64(v), "0")
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range a {
			r.Get(uint64(v))
		}
	}
	b.StopTimer()
	r.Close()
}

func TestSetGet2(t *testing.T) {
	const N = 70000
	//for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
	for _, x := range []uint64{0, 1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]uint64, N)
		rng := rng()
		for i := range a {
			a[i] = (uint64(rng.Next()) ^ x) << 1
		}
		for i, k := range a {
			set(k, strconv.FormatInt(int64(uint64(k)^x), 10))
			if g, e := r.Len(), i+1; g != e {
				t.Fatal(i, x, g, e)
			}
		}

		for i, k := range a {
			v, ok := r.Get(uint64(k))
			if !ok {
				t.Fatal(i, k, v, ok)
			}

			if g, e := v, strconv.FormatInt(int64(uint64(k)^x), 10); g != e {
				t.Fatal(i, g, e)
			}

			k |= 1
			_, ok = r.Get(uint64(k))
			if ok {
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
		set(uint64(i), strconv.FormatInt(int64(-i), 10))
		if _, ok := r.r.(*x); ok {
			break
		}
	}
	for j := 0; j <= i; j++ {
		set(uint64(j), strconv.FormatInt(int64(j), 10))
	}

	for j := 0; j <= i; j++ {
		v, ok := r.Get(uint64(j))
		if !ok {
			t.Fatal(j)
		}

		if g, e := v, strconv.FormatInt(int64(j), 10); g != e {
			t.Fatal(g, e)
		}
	}
}

const key3 uint64 = 0
const key4 uint64 = 1
const value4 string = ""
const value5 string = "1"

func TestDelete0(t *testing.T) {
	r := TreeNew(cmp)
	if ok := r.Delete(key3); ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	r.Set(key3, value4)
	if ok := r.Delete(key4); ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(key3); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(key3); ok {
		t.Fatal(ok)
	}

	r.Set(key3, value4)
	r.Set(key4, value5)
	if ok := r.Delete(key4); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(key4); ok {
		t.Fatal(ok)
	}

	if ok := r.Delete(key3); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(key3); ok {
		t.Fatal(ok)
	}

	r.Set(key3, value4)
	r.Set(key4, value5)
	if ok := r.Delete(key3); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 1; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(key3); ok {
		t.Fatal(ok)
	}

	if ok := r.Delete(key4); !ok {
		t.Fatal(ok)
	}

	if g, e := r.Len(), 0; g != e {
		t.Fatal(g, e)
	}

	if ok := r.Delete(key4); ok {
		t.Fatal(ok)
	}
}

func TestDelete1(t *testing.T) {
	const N = 100000
	//for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
	for _, x := range []uint64{0, 1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]uint64, N)
		for i := range a {
			a[i] = (uint64(i) ^ x) << 1
		}
		for _, k := range a {
			set(k, value4)
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

func BenchmarkDelSeq1e3(b *testing.B) {
	benchmarkDelSeq(b, 1e3)
}

func BenchmarkDelSeq1e4(b *testing.B) {
	benchmarkDelSeq(b, 1e4)
}

func BenchmarkDelSeq1e5(b *testing.B) {
	benchmarkDelSeq(b, 1e5)
}

func BenchmarkDelSeq1e6(b *testing.B) {
	benchmarkDelSeq(b, 1e6)
}

func benchmarkDelSeq(b *testing.B, n int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := TreeNew(cmp)
		for i := 0; i < n; i++ {
			r.Set(uint64(i), strconv.FormatInt(int64(i), 10))
		}
		debug.FreeOSMemory()
		b.StartTimer()
		for j := 0; j < n; j++ {
			r.Delete(uint64(j))
		}
		b.StopTimer()
		r.Close()
	}
	b.StopTimer()
}

func BenchmarkDelRnd1e3(b *testing.B) {
	benchmarkDelRnd(b, 1e3)
}

func BenchmarkDelRnd1e4(b *testing.B) {
	benchmarkDelRnd(b, 1e4)
}

func BenchmarkDelRnd1e5(b *testing.B) {
	benchmarkDelRnd(b, 1e5)
}

func BenchmarkDelRnd1e6(b *testing.B) {
	benchmarkDelRnd(b, 1e6)
}

func benchmarkDelRnd(b *testing.B, n int) {
	rng := rng()
	a := make([]uint64, n)
	for i := range a {
		a[i] = uint64(rng.Next())
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := TreeNew(cmp)
		for _, v := range a {
			r.Set(v, value4)
		}
		debug.FreeOSMemory()
		b.StartTimer()
		for _, v := range a {
			r.Delete(v)
		}
		b.StopTimer()
		r.Close()
	}
	b.StopTimer()
}

func TestDelete2(t *testing.T) {
	const N = 80000
	//for _, x := range []int{0, -1, 0x555555, 0xaaaaaa, 0x314259} {
	for _, x := range []uint64{0, 1, 0x555555, 0xaaaaaa, 0x314259} {
		r := TreeNew(cmp)
		set := r.Set
		a := make([]uint64, N)
		rng := rng()
		for i := range a {
			a[i] = (uint64(rng.Next()) ^ x) << 1
		}
		for _, k := range a {
			set(k, value4)
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

const key10 uint64 = 5
const key11 uint64 = 10
const key12 uint64 = 15
const key13 uint64 = 20
const key14 uint64 = 25
const key15 uint64 = 30
const key16 uint64 = 35

const hit10 bool = false
const hit11 bool = true
const hit12 bool = false
const hit13 bool = true
const hit14 bool = false
const hit15 bool = true
const hit16 bool = false

var vlist10 = []uint64{10, 20, 30}
var vlist11 = []uint64{10, 20, 30}
var vlist12 = []uint64{20, 30}
var vlist13 = []uint64{20, 30}
var vlist14 = []uint64{30}
var vlist15 = []uint64{30}
var vlist16 = []uint64{}

const value10 string = "100"
const value11 string = "200"
const value12 string = "300"

func TestEnumeratorNext(t *testing.T) {
	// seeking within 3 keys: 10, 20, 30
	table := []struct {
		k    uint64
		hit  bool
		keys []uint64
	}{
		{key10, hit10, vlist10},
		{key11, hit11, vlist11},
		{key12, hit12, vlist12},
		{key13, hit13, vlist13},
		{key14, hit14, vlist14},
		{key15, hit15, vlist15},
		{key16, hit16, vlist16},
	}

	for i, test := range table {
		up := test.keys
		r := TreeNew(cmp)

		r.Set(key11, value10)
		r.Set(key13, value11)
		r.Set(key15, value12)

		for verChange := 0; verChange < 16; verChange++ {
			en, hit := r.Seek(test.k)

			if g, e := hit, test.hit; g != e {
				t.Fatal(i, g, e)
			}

			j := 0
			for {
				if verChange&(1<<uint(j)) != 0 {
					r.Set(key13, value11)
				}

				k, v, err := en.Next()
				if err != nil {
					if !fileutil.IsEOF(err) {
						t.Fatal(i, err)
					}

					break
				}

				if j >= len(up) {
					t.Fatal(i, j, verChange)
				}

				if g, e := k, up[j]; g != e {
					t.Fatal(i, j, verChange, g, e)
				}

				if g, e := v, strconv.FormatInt(int64(10*up[j]), 10); g != e {
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

var vlist20 = []uint64{}
var vlist21 = []uint64{10}
var vlist22 = []uint64{10}
var vlist23 = []uint64{20, 10}
var vlist24 = []uint64{20, 10}
var vlist25 = []uint64{30, 20, 10}
var vlist26 = []uint64{30, 20, 10}

func TestEnumeratorPrev(t *testing.T) {
	// seeking within 3 keys: 10, 20, 30
	table := []struct {
		k    uint64
		hit  bool
		keys []uint64
	}{
		{key10, hit10, vlist20},
		{key11, hit11, vlist21},
		{key12, hit12, vlist22},
		{key13, hit13, vlist23},
		{key14, hit14, vlist24},
		{key15, hit15, vlist25},
		{key16, hit16, vlist26},
	}

	for i, test := range table {
		dn := test.keys
		r := TreeNew(cmp)

		r.Set(key11, value10)
		r.Set(key13, value11)
		r.Set(key15, value12)

		for verChange := 0; verChange < 16; verChange++ {
			en, hit := r.Seek(test.k)

			if g, e := hit, test.hit; g != e {
				t.Fatal(i, g, e)
			}

			j := 0
			for {
				if verChange&(1<<uint(j)) != 0 {
					r.Set(key13, value11)
				}

				k, v, err := en.Prev()
				if err != nil {
					if !fileutil.IsEOF(err) {
						t.Fatal(i, err)
					}

					break
				}

				if j >= len(dn) {
					t.Fatal(i, j, verChange)
				}

				if g, e := k, dn[j]; g != e {
					t.Fatal(i, j, verChange, g, e)
				}

				if g, e := v, strconv.FormatInt(int64(10*dn[j]), 10); g != e {
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

func TestEnumeratorPrevSanity(t *testing.T) {
	// seeking within 3 keys: 10, 20, 30
	table := []struct {
		k      uint64
		hit    bool
		kOut   uint64
		vOut   string
		errOut error
	}{
		{key11, hit11, key11, value10, nil},
		{key13, hit13, key13, value11, nil},
		{key15, hit15, key15, value12, nil},
		{key16, hit16, key15, value12, nil},
		{key14, hit14, key13, value11, nil},
		{key12, hit12, key11, value10, nil},
		{key10, hit10, 0, "", io.EOF},
	}

	for i, test := range table {
		r := TreeNew(cmp)

		r.Set(key11, value10)
		r.Set(key13, value11)
		r.Set(key15, value12)

		en, hit := r.Seek(test.k)

		if g, e := hit, test.hit; g != e {
			t.Fatal(i, g, e)
		}

		k, v, err := en.Prev()

		if g, e := err, test.errOut; g != e {
			t.Fatal(i, g, e)
		}
		if g, e := k, test.kOut; g != e {
			t.Fatal(i, g, e)
		}
		if g, e := v, test.vOut; g != e {
			t.Fatal(i, g, e)
		}
	}
}

func BenchmarkSeekSeq1e3(b *testing.B) {
	benchmarkSeekSeq(b, 1e3)
}

func BenchmarkSeekSeq1e4(b *testing.B) {
	benchmarkSeekSeq(b, 1e4)
}

func BenchmarkSeekSeq1e5(b *testing.B) {
	benchmarkSeekSeq(b, 1e5)
}

func BenchmarkSeekSeq1e6(b *testing.B) {
	benchmarkSeekSeq(b, 1e6)
}

func benchmarkSeekSeq(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		t := TreeNew(cmp)
		for j := 0; j < n; j++ {
			t.Set(uint64(j), "")
		}
		debug.FreeOSMemory()
		b.StartTimer()
		for j := 0; j < n; j++ {
			e, _ := t.Seek(uint64(j))
			e.Close()
		}
		b.StopTimer()
		t.Close()
	}
	b.StopTimer()
}

func BenchmarkSeekRnd1e3(b *testing.B) {
	benchmarkSeekRnd(b, 1e3)
}

func BenchmarkSeekRnd1e4(b *testing.B) {
	benchmarkSeekRnd(b, 1e4)
}

func BenchmarkSeekRnd1e5(b *testing.B) {
	benchmarkSeekRnd(b, 1e5)
}

func BenchmarkSeekRnd1e6(b *testing.B) {
	benchmarkSeekRnd(b, 1e6)
}

func benchmarkSeekRnd(b *testing.B, n int) {
	r := TreeNew(cmp)
	rng := rng()
	a := make([]uint64, n)
	for i := range a {
		a[i] = uint64(rng.Next())
	}
	for _, v := range a {
		r.Set(v, "")
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range a {
			e, _ := r.Seek(v)
			e.Close()
		}
	}
	b.StopTimer()
	r.Close()
}

func BenchmarkNext1e3(b *testing.B) {
	benchmarkNext(b, 1e3)
}

func BenchmarkNext1e4(b *testing.B) {
	benchmarkNext(b, 1e4)
}

func BenchmarkNext1e5(b *testing.B) {
	benchmarkNext(b, 1e5)
}

func BenchmarkNext1e6(b *testing.B) {
	benchmarkNext(b, 1e6)
}

func benchmarkNext(b *testing.B, n int) {
	t := TreeNew(cmp)
	for i := 0; i < n; i++ {
		t.Set(uint64(i), "")
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en, err := t.SeekFirst()
		if err != nil {
			b.Fatal(err)
		}

		m := 0
		for {
			if _, _, err = en.Next(); err != nil {
				break
			}
			m++
		}
		if m != n {
			b.Fatal(m)
		}
	}
	b.StopTimer()
	t.Close()
}

func BenchmarkPrev1e3(b *testing.B) {
	benchmarkPrev(b, 1e3)
}

func BenchmarkPrev1e4(b *testing.B) {
	benchmarkPrev(b, 1e4)
}

func BenchmarkPrev1e5(b *testing.B) {
	benchmarkPrev(b, 1e5)
}

func BenchmarkPrev1e6(b *testing.B) {
	benchmarkPrev(b, 1e6)
}

func benchmarkPrev(b *testing.B, n int) {
	t := TreeNew(cmp)
	for i := 0; i < n; i++ {
		t.Set(uint64(i), "")
	}
	debug.FreeOSMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		en, err := t.SeekLast()
		if err != nil {
			b.Fatal(err)
		}

		m := 0
		for {
			if _, _, err = en.Prev(); err != nil {
				break
			}
			m++
		}
		if m != n {
			b.Fatal(m)
		}
	}
	b.StopTimer()
	t.Close()
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
	b.Set(uint64(1), "10")
	en, err := b.SeekFirst()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Next()
	if k != uint64(1) || v != "10" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekFirst2(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(uint64(1), "10")
	b.Set(uint64(2), "20")
	en, err := b.SeekFirst()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Next()
	if k != uint64(1) || v != "10" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if k != uint64(2) || v != "20" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekFirst3(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(uint64(2), "20")
	b.Set(uint64(3), "30")
	b.Set(uint64(1), "10")
	en, err := b.SeekFirst()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Next()
	if k != uint64(1) || v != "10" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if k != uint64(2) || v != "20" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Next()
	if k != uint64(3) || v != "30" || err != nil {
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
	b.Set(uint64(1), "10")
	en, err := b.SeekLast()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Prev()
	if k != uint64(1) || v != "10" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekLast2(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(uint64(1), "10")
	b.Set(uint64(2), "20")
	en, err := b.SeekLast()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Prev()
	if k != uint64(2) || v != "20" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if k != uint64(1) || v != "10" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestSeekLast3(t *testing.T) {
	b := TreeNew(cmp)
	b.Set(uint64(2), "20")
	b.Set(uint64(3), "30")
	b.Set(uint64(1), "10")
	en, err := b.SeekLast()
	if err != nil {
		t.Fatal(err)
	}

	k, v, err := en.Prev()
	if k != uint64(3) || v != "30" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if k != uint64(2) || v != "20" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if k != uint64(1) || v != "10" || err != nil {
		t.Fatal(k, v, err)
	}

	k, v, err = en.Prev()
	if err == nil {
		t.Fatal(k, v, err)
	}
}

func TestPut(t *testing.T) {
	tab := []struct {
		pre    []uint64 // even index: K, odd index: V
		newK   uint64   // Put(newK, ...
		oldV   string   // Put()->oldV
		exists bool  // upd(exists)
		write  bool  // upd()->write
		post   []uint64 // even index: K, odd index: V
	}{
		// 0
		{
			[]uint64{},
			uint64(1), "0", false, false,
			[]uint64{},
		},
		{
			[]uint64{},
			uint64(1), "0", false, true,
			[]uint64{1, 9},
		},
		{
			[]uint64{1, 10},
			uint64(0), "0", false, false,
			[]uint64{1, 10},
		},
		{
			[]uint64{1, 10},
			uint64(0), "0", false, true,
			[]uint64{0, 9, 1, 10},
		},
		{
			[]uint64{1, 10},
			uint64(1), "10", true, false,
			[]uint64{1, 10},
		},

		// 5
		{
			[]uint64{1, 10},
			uint64(1), "10", true, true,
			[]uint64{1, 9},
		},
		{
			[]uint64{1, 10},
			uint64(2), "0", false, false,
			[]uint64{1, 10},
		},
		{
			[]uint64{1, 10},
			uint64(2), "0", false, true,
			[]uint64{1, 10, 2, 9},
		},
	}

	for iTest, test := range tab {
		tr := TreeNew(cmp)
		for i := 0; i < len(test.pre); i += 2 {
			k, v := test.pre[i], strconv.FormatInt(int64(test.pre[i+1]), 10)
			tr.Set(k, v)
		}

		oldV, written := tr.Put(test.newK, func(old string, exists bool) (newV string, write bool) {
			if g, e := exists, test.exists; g != e {
				t.Fatal(iTest, g, e)
			}

			if exists {
				if g, e := old, test.oldV; g != e {
					t.Fatal(iTest, g, e)
				}
			}
			return "9", test.write
		})
		if test.exists {
			if g, e := oldV, test.oldV; g != e {
				t.Fatal(iTest, g, e)
			}
		}

		if g, e := written, test.write; g != e {
			t.Fatal(iTest, g, e)
		}

		n := len(test.post)
		en, err := tr.SeekFirst()
		if err != nil {
			if n == 0 && err == io.EOF {
				continue
			}

			t.Fatal(iTest, err)
		}

		for i := 0; i < len(test.post); i += 2 {
			k, v, err := en.Next()
			if err != nil {
				t.Fatal(iTest, err)
			}

			if g, e := k, test.post[i]; g != e {
				t.Fatal(iTest, g, e)
			}

			if g, e := v, strconv.FormatInt(int64(test.post[i+1]), 10); g != e {
				t.Fatal(iTest, g, e)
			}
		}

		_, _, err = en.Next()
		if g, e := err, io.EOF; g != e {
			t.Fatal(iTest, g, e)
		}
	}
}
