// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package b implements a B+tree.
//
// Changelog
//
// 2014-04-18: Added new method Put.
//
// Generic types
//
// Keys and their associated values are interface{} typed, similar to all of
// the containers in the standard library.
//
// Semiautomatic production of a type specific variant of this package is
// supported via
//
//	$ make generic
//
// This command will write to stdout a version of the btree.go file where
// every key type occurrence is replaced by the word 'key' (written in all
// CAPS) and every value type occurrence is replaced by the word 'value'
// (written in all CAPS). Then you have to replace these tokens with your
// desired type(s), using any technique you're comfortable with.
//
// This is how, for example, 'example/int.go' was created:
//
//	$ mkdir example
//	$
//	$ # Note: the command bellow must be actually written using the words
//	$ # 'key' and 'value' in all CAPS. The proper form is avoided in this
//	$ # documentation to not confuse any text replacement mechanism.
//	$
//	$ make generic | sed -e 's/key/int/g' -e 's/value/int/g' > example/int.go
//
// No other changes to int.go are (strictly) necessary, it compiles just fine.
//
// Running the benchmarks for 1000 keys on a machine with Intel X5450 CPU @ 3
// GHz, Go release 1.3.
//
//	$ go test -bench 1e3 example/all_test.go example/int.go
//	PASS
//	BenchmarkSetSeq1e3	   10000	    263951 ns/op
//	BenchmarkGetSeq1e3	   10000	    154410 ns/op
//	BenchmarkSetRnd1e3	    5000	    392690 ns/op
//	BenchmarkGetRnd1e3	   10000	    181776 ns/op
//	BenchmarkDelRnd1e3	    5000	    323795 ns/op
//	BenchmarkSeekSeq1e3	   10000	    235939 ns/op
//	BenchmarkSeekRnd1e3	    5000	    299997 ns/op
//	BenchmarkNext1e3	  200000	     14202 ns/op
//	BenchmarkPrev1e3	  200000	     13842 ns/op
//	ok  	command-line-arguments	30.620s
package b

import (
	"io"
)

//TODO check vs orig initialize/finalize

const (
	kx = 128 // min 2 //TODO benchmark tune this number if using custom key/value type(s).
	kd = 64  // min 1 //TODO benchmark tune this number if using custom key/value type(s).
)

type (
	// Cmp compares a and b. Return value is:
	//
	//	< 0 if a <  b
	//	  0 if a == b
	//	> 0 if a >  b
	//
	Cmp func(a, b interface{} /*K*/) int

	d struct { // data page
		c int
		d [2*kd + 1]de
		n *d
		p *d
	}

	de struct { // d element
		k interface{} /*K*/
		v interface{} /*V*/
	}

	// Enumerator captures the state of enumerating a tree. It is returned
	// from the Seek* methods. The enumerator is aware of any mutations
	// made to the tree in the process of enumerating it and automatically
	// resumes the enumeration at the proper key, if possible.
	//
	// However, once an Enumerator returns io.EOF to signal "no more
	// items", it does no more attempt to "resync" on tree mutation(s).  In
	// other words, io.EOF from an Enumaretor is "sticky" (idempotent).
	Enumerator struct {
		err error
		hit bool
		i   int
		k   interface{} /*K*/
		q   *d
		t   *Tree
		ver int64
	}

	// Tree is a B+tree.
	Tree struct {
		c     int
		cmp   Cmp
		first *d
		last  *d
		r     interface{}
		ver   int64
	}

	xe struct { // x element
		ch interface{}
		k  interface{} /*K*/
	}

	x struct { // index page
		c int
		x [2*kx + 2]xe
	}
)

var ( // R/O zero values
	zd  d
	zde de
	zx  x
	zxe xe
	zk  interface{} /*K*/
)

func clr(q interface{}) {
	switch x := q.(type) {
	case *x:
		for i := 0; i <= x.c; i++ { // Ch0 Sep0 ... Chn-1 Sepn-1 Chn
			clr(x.x[i].ch)
		}
		*x = zx // GC
	case *d:
		*x = zd // GC
	}
}

// -------------------------------------------------------------------------- x

func newX(ch0 interface{}) *x {
	r := &x{}
	r.x[0].ch = ch0
	return r
}

func (q *x) extract(i int) {
	q.c--
	if i < q.c {
		copy(q.x[i:], q.x[i+1:q.c+1])
		q.x[q.c].ch = q.x[q.c+1].ch
		q.x[q.c].k = zk  // GC
		q.x[q.c+1] = zxe // GC
	}
}

func (q *x) insert(i int, k interface{} /*K*/, ch interface{}) *x {
	c := q.c
	if i < c {
		q.x[c+1].ch = q.x[c].ch
		copy(q.x[i+2:], q.x[i+1:c])
		q.x[i+1].k = q.x[i].k
	}
	c++
	q.c = c
	q.x[i].k = k
	q.x[i+1].ch = ch
	return q
}

func (q *x) siblings(i int) (l, r *d) {
	if i >= 0 {
		if i > 0 {
			l = q.x[i-1].ch.(*d)
		}
		if i < q.c {
			r = q.x[i+1].ch.(*d)
		}
	}
	return
}

// -------------------------------------------------------------------------- d

func (l *d) mvL(r *d, c int) {
	copy(l.d[l.c:], r.d[:c])
	copy(r.d[:], r.d[c:r.c])
	l.c += c
	r.c -= c
}

func (l *d) mvR(r *d, c int) {
	copy(r.d[c:], r.d[:r.c])
	copy(r.d[:c], l.d[l.c-c:])
	r.c += c
	l.c -= c
}

// ----------------------------------------------------------------------- Tree

// TreeNew returns a newly created, empty Tree. The compare function is used
// for key collation.
func TreeNew(cmp Cmp) *Tree {
	return &Tree{cmp: cmp}
}

// Clear removes all K/V pairs from the tree.
func (t *Tree) Clear() {
	if t.r == nil {
		return
	}

	clr(t.r)
	t.c, t.first, t.last, t.r = 0, nil, nil, nil
	t.ver++
}

func (t *Tree) cat(p *x, q, r *d, pi int) {
	t.ver++
	q.mvL(r, r.c)
	if r.n != nil {
		r.n.p = q
	} else {
		t.last = q
	}
	q.n = r.n //TODO recycle r
	if p.c > 1 {
		p.extract(pi)
		p.x[pi].ch = q
	} else { //TODO recycle r
		t.r = q
	}
}

func (t *Tree) catX(p, q, r *x, pi int) {
	t.ver++
	q.x[q.c].k = p.x[pi].k
	copy(q.x[q.c+1:], r.x[:r.c])
	q.c += r.c + 1
	q.x[q.c].ch = r.x[r.c].ch //TODO recycle r
	if p.c > 1 {
		p.c--
		pc := p.c
		if pi < pc {
			p.x[pi].k = p.x[pi+1].k
			copy(p.x[pi+1:], p.x[pi+2:pc+1])
			p.x[pc].ch = p.x[pc+1].ch
			p.x[pc].k = zk     // GC
			p.x[pc+1].ch = nil // GC
		}
		return
	}

	t.r = q //TODO recycle r
}

// Delete removes the k's KV pair, if it exists, in which case Delete returns
// true.
func (t *Tree) Delete(k interface{} /*K*/) (ok bool) {
	pi := -1
	var p *x
	q := t.r
	if q == nil {
		return false
	}

	for {
		var i int
		i, ok = t.find(q, k)
		if ok {
			switch x := q.(type) {
			case *x:
				if x.c < kx && q != t.r {
					t.underflowX(p, &x, pi, &i)
				}
				pi = i + 1
				p = x
				q = x.x[pi].ch
				ok = false
				continue
			case *d:
				t.extract(x, i)
				if x.c >= kd {
					return true
				}

				if q != t.r {
					t.underflow(p, x, pi)
				} else if t.c == 0 {
					t.Clear()
				}
				return true
			}
		}

		switch x := q.(type) {
		case *x:
			if x.c < kx && q != t.r {
				t.underflowX(p, &x, pi, &i)
			}
			pi = i
			p = x
			q = x.x[i].ch
		case *d:
			return false
		}
	}
}

func (t *Tree) extract(q *d, i int) { // (r interface{} /*V*/) {
	t.ver++
	//r = q.d[i].v // prepared for Extract
	q.c--
	if i < q.c {
		copy(q.d[i:], q.d[i+1:q.c+1])
	}
	q.d[q.c] = zde // GC
	t.c--
	return
}

func (t *Tree) find(q interface{}, k interface{} /*K*/) (i int, ok bool) {
	var mk interface{} /*K*/
	l := 0
	switch x := q.(type) {
	case *x:
		h := x.c - 1
		for l <= h {
			m := (l + h) >> 1
			mk = x.x[m].k
			switch cmp := t.cmp(k, mk); {
			case cmp > 0:
				l = m + 1
			case cmp == 0:
				return m, true
			default:
				h = m - 1
			}
		}
	case *d:
		h := x.c - 1
		for l <= h {
			m := (l + h) >> 1
			mk = x.d[m].k
			switch cmp := t.cmp(k, mk); {
			case cmp > 0:
				l = m + 1
			case cmp == 0:
				return m, true
			default:
				h = m - 1
			}
		}
	}
	return l, false
}

//A // First returns the first item of the tree in the key collating order, or
//A // (zero-value, zero-value) if the tree is empty.
//A func (t *Tree) First() (k interface{} /*K*/, v interface{} /*V*/) {
//A 	if q := t.first; q != nil {
//A 		q := &q.d[0]
//A 		k, v = q.k, q.v
//A 	}
//A 	return
//A }

// Get returns the value associated with k and true if it exists. Otherwise Get
// returns (zero-value, false).
func (t *Tree) Get(k interface{} /*K*/) (v interface{} /*V*/, ok bool) {
	q := t.r
	if q == nil {
		return
	}

	for {
		var i int
		if i, ok = t.find(q, k); ok {
			switch x := q.(type) {
			case *x:
				q = x.x[i+1].ch
				continue
			case *d:
				return x.d[i].v, true
			}
		}
		switch x := q.(type) {
		case *x:
			q = x.x[i].ch
		default:
			return
		}
	}
}

func (t *Tree) insert(q *d, i int, k interface{} /*K*/, v interface{} /*V*/) *d {
	t.ver++
	c := q.c
	if i < c {
		copy(q.d[i+1:], q.d[i:c])
	}
	c++
	q.c = c
	q.d[i].k, q.d[i].v = k, v
	t.c++
	return q
}

//A // Last returns the last item of the tree in the key collating order, or
//A // (zero-value, zero-value) if the tree is empty.
//A func (t *Tree) Last() (k interface{} /*K*/, v interface{} /*V*/) {
//A 	if q := t.last; q != nil {
//A 		q := &q.d[q.c-1]
//A 		k, v = q.k, q.v
//A 	}
//A 	return
//A }

// Len returns the number of items in the tree.
func (t *Tree) Len() int {
	return t.c
}

func (t *Tree) overflow(p *x, q *d, pi, i int, k interface{} /*K*/, v interface{} /*V*/) {
	t.ver++
	l, r := p.siblings(pi)

	if l != nil && l.c < 2*kd {
		l.mvL(q, 1)
		t.insert(q, i-1, k, v)
		p.x[pi-1].k = q.d[0].k
		return
	}

	if r != nil && r.c < 2*kd {
		if i < 2*kd {
			q.mvR(r, 1)
			t.insert(q, i, k, v)
			p.x[pi].k = r.d[0].k
		} else {
			t.insert(r, 0, k, v)
			p.x[pi].k = k
		}
		return
	}

	t.split(p, q, pi, i, k, v)
}

// Seek returns an Enumerator positioned on a an item such that k >= item's
// key. ok reports if k == item.key The Enumerator's position is possibly
// after the last item in the tree.
func (t *Tree) Seek(k interface{} /*K*/) (e *Enumerator, ok bool) {
	q := t.r
	if q == nil {
		e = &Enumerator{nil, false, 0, k, nil, t, t.ver}
		return
	}

	for {
		var i int
		if i, ok = t.find(q, k); ok {
			switch x := q.(type) {
			case *x:
				q = x.x[i+1].ch
				continue
			case *d:
				e = &Enumerator{nil, ok, i, k, x, t, t.ver}
				return
			}
		}
		switch x := q.(type) {
		case *x:
			q = x.x[i].ch
		case *d:
			e = &Enumerator{nil, ok, i, k, x, t, t.ver}
			return
		}
	}
}

// SeekFirst returns an enumerator positioned on the first KV pair in the tree,
// if any. For an empty tree, err == io.EOF is returned and e will be nil.
func (t *Tree) SeekFirst() (e *Enumerator, err error) {
	q := t.first
	if q == nil {
		return nil, io.EOF
	}

	return &Enumerator{nil, true, 0, q.d[0].k, q, t, t.ver}, nil
}

// SeekLast returns an enumerator positioned on the last KV pair in the tree,
// if any. For an empty tree, err == io.EOF is returned and e will be nil.
func (t *Tree) SeekLast() (e *Enumerator, err error) {
	q := t.last
	if q == nil {
		return nil, io.EOF
	}

	return &Enumerator{nil, true, q.c - 1, q.d[q.c-1].k, q, t, t.ver}, nil
}

// Set sets the value associated with k.
func (t *Tree) Set(k interface{} /*K*/, v interface{} /*V*/) {
	//dbg("--- PRE Set(%v, %v)\n%s", k, v, t.dump())
	//defer func() {
	//	dbg("--- POST\n%s\n====\n", t.dump())
	//}()
	pi := -1
	var p *x
	q := t.r
	if q == nil {
		z := t.insert(&d{}, 0, k, v)
		t.r, t.first, t.last = z, z, z
		return
	}

	for {
		i, ok := t.find(q, k)
		if ok {
			switch x := q.(type) {
			case *x:
				q = x.x[i+1].ch
				continue
			case *d:
				x.d[i].v = v
			}
			return
		}

		switch x := q.(type) {
		case *x:
			if x.c > 2*kx {
				t.splitX(p, &x, pi, &i)
			}
			pi = i
			p = x
			q = x.x[i].ch
		case *d:
			switch {
			case x.c < 2*kd:
				t.insert(x, i, k, v)
			default:
				t.overflow(p, x, pi, i, k, v)
			}
			return
		}
	}
}

// Put combines Get and Set in a more efficient way where the tree is walked
// only once. The upd(ater) receives (old-value, true) if a KV pair for k
// exists or (zero-value, false) otherwise. It can then return a (new-value,
// true) to create or overwrite the existing value in the KV pair, or
// (whatever, false) if it decides not to create or not to update the value of
// the KV pair.
//
// 	tree.Set(k, v) conceptually equals
//
// 	tree.Put(k, func(k, v []byte){ return v, true }([]byte, bool))
//
// modulo the differing return values.
func (t *Tree) Put(k interface{} /*K*/, upd func(oldV interface{} /*V*/, exists bool) (newV interface{} /*V*/, write bool)) (oldV interface{} /*V*/, written bool) {
	pi := -1
	var p *x
	q := t.r
	var newV interface{} /*V*/
	if q != nil {
		for {
			i, ok := t.find(q, k)
			if ok {
				switch x := q.(type) {
				case *x:
					q = x.x[i+1].ch
					continue
				case *d:
					oldV = x.d[i].v
					newV, written = upd(oldV, true)
					if !written {
						return
					}

					x.d[i].v = newV
				}
				return
			}

			switch x := q.(type) {
			case *x:
				if x.c > 2*kx {
					t.splitX(p, &x, pi, &i)
				}
				pi = i
				p = x
				q = x.x[i].ch
			case *d: // new KV pair
				newV, written = upd(newV, false)
				if !written {
					return
				}

				switch {
				case x.c < 2*kd:
					t.insert(x, i, k, newV)
				default:
					t.overflow(p, x, pi, i, k, newV)
				}
				return
			}
		}
	}

	// new KV pair in empty tree
	newV, written = upd(newV, false)
	if !written {
		return
	}

	z := t.insert(&d{}, 0, k, newV)
	t.r, t.first, t.last = z, z, z
	return
}

func (t *Tree) split(p *x, q *d, pi, i int, k interface{} /*K*/, v interface{} /*V*/) {
	t.ver++
	r := &d{}
	if q.n != nil {
		r.n = q.n
		r.n.p = r
	} else {
		t.last = r
	}
	q.n = r
	r.p = q

	copy(r.d[:], q.d[kd:2*kd])
	for i := range q.d[kd:] {
		q.d[kd+i] = zde
	}
	q.c = kd
	r.c = kd
	var done bool
	if i > kd {
		done = true
		t.insert(r, i-kd, k, v)
	}
	if pi >= 0 {
		p.insert(pi, r.d[0].k, r)
	} else {
		t.r = newX(q).insert(0, r.d[0].k, r)
	}
	if done {
		return
	}

	t.insert(q, i, k, v)
}

func (t *Tree) splitX(p *x, pp **x, pi int, i *int) {
	t.ver++
	q := *pp
	r := &x{}
	copy(r.x[:], q.x[kx+1:])
	q.c = kx
	r.c = kx
	if pi >= 0 {
		p.insert(pi, q.x[kx].k, r)
	} else {
		t.r = newX(q).insert(0, q.x[kx].k, r)
	}
	q.x[kx].k = zk
	for i := range q.x[kx+1:] {
		q.x[kx+i+1] = zxe
	}
	if *i > kx {
		*pp = r
		*i -= kx + 1
	}
}

func (t *Tree) underflow(p *x, q *d, pi int) {
	t.ver++
	l, r := p.siblings(pi)

	if l != nil && l.c+q.c >= 2*kd {
		l.mvR(q, 1)
		p.x[pi-1].k = q.d[0].k
	} else if r != nil && q.c+r.c >= 2*kd {
		q.mvL(r, 1)
		p.x[pi].k = r.d[0].k
		r.d[r.c] = zde // GC
	} else if l != nil {
		t.cat(p, l, q, pi-1)
	} else {
		t.cat(p, q, r, pi)
	}
}

func (t *Tree) underflowX(p *x, pp **x, pi int, i *int) {
	t.ver++
	var l, r *x
	q := *pp

	if pi >= 0 {
		if pi > 0 {
			l = p.x[pi-1].ch.(*x)
		}
		if pi < p.c {
			r = p.x[pi+1].ch.(*x)
		}
	}

	if l != nil && l.c > kx {
		q.x[q.c+1].ch = q.x[q.c].ch
		copy(q.x[1:], q.x[:q.c])
		q.x[0].ch = l.x[l.c].ch
		q.x[0].k = p.x[pi-1].k
		q.c++
		*i++
		l.c--
		p.x[pi-1].k = l.x[l.c].k
		return
	}

	if r != nil && r.c > kx {
		q.x[q.c].k = p.x[pi].k
		q.c++
		q.x[q.c].ch = r.x[0].ch
		p.x[pi].k = r.x[0].k
		copy(r.x[:], r.x[1:r.c])
		r.c--
		rc := r.c
		r.x[rc].ch = r.x[rc+1].ch
		r.x[rc].k = zk
		r.x[rc+1].ch = nil
		return
	}

	if l != nil {
		*i += l.c + 1
		t.catX(p, l, q, pi-1)
		*pp = l
		return
	}

	t.catX(p, q, r, pi)
}

// ----------------------------------------------------------------- Enumerator

// Next returns the currently enumerated item, if it exists and moves to the
// next item in the key collation order. If there is no item to return, err ==
// io.EOF is returned.
func (e *Enumerator) Next() (k interface{} /*K*/, v interface{} /*V*/, err error) {
	if err = e.err; err != nil {
		return
	}

	if e.ver != e.t.ver {
		f, hit := e.t.Seek(e.k)
		if !e.hit && hit {
			if err = f.next(); err != nil {
				return
			}
		}

		*e = *f
	}
	if e.q == nil {
		e.err, err = io.EOF, io.EOF
		return
	}

	if e.i >= e.q.c {
		if err = e.next(); err != nil {
			return
		}
	}

	i := e.q.d[e.i]
	k, v = i.k, i.v
	e.k, e.hit = k, false
	e.next()
	return
}

func (e *Enumerator) next() error {
	if e.q == nil {
		e.err = io.EOF
		return io.EOF
	}

	switch {
	case e.i < e.q.c-1:
		e.i++
	default:
		if e.q, e.i = e.q.n, 0; e.q == nil {
			e.err = io.EOF
		}
	}
	return e.err
}

// Prev returns the currently enumerated item, if it exists and moves to the
// previous item in the key collation order. If there is no item to return, err
// == io.EOF is returned.
func (e *Enumerator) Prev() (k interface{} /*K*/, v interface{} /*V*/, err error) {
	if err = e.err; err != nil {
		return
	}

	if e.ver != e.t.ver {
		f, hit := e.t.Seek(e.k)
		if !e.hit && hit {
			if err = f.prev(); err != nil {
				return
			}
		}

		*e = *f
	}
	if e.q == nil {
		e.err, err = io.EOF, io.EOF
		return
	}

	if e.i >= e.q.c {
		if err = e.next(); err != nil {
			return
		}
	}

	i := e.q.d[e.i]
	k, v = i.k, i.v
	e.k, e.hit = k, false
	e.prev()
	return
}

func (e *Enumerator) prev() error {
	if e.q == nil {
		e.err = io.EOF
		return io.EOF
	}

	switch {
	case e.i > 0:
		e.i--
	default:
		if e.q = e.q.p; e.q == nil {
			e.err = io.EOF
			break
		}

		e.i = e.q.c - 1
	}
	return e.err
}
