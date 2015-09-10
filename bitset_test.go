// Copyright 2014 Will Fitzgerald. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file tests bit sets

package sparsebitset

import (
	"math"
	"math/rand"
	"testing"
)

func TestEmptyBitSet(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("A zero-length bitset should be fine")
		}
	}()
	b := New(0)
	if b.Len() != 0 {
		t.Errorf("Empty set should have capacity 0, not %d", b.Len())
	}
}

func TestZeroValueBitSet(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("A zero-length bitset should be fine")
		}
	}()
	var b BitSet
	if b.Len() != 0 {
		t.Errorf("Empty set should have capacity 0, not %d", b.Len())
	}
}

func TestBitSetNew(t *testing.T) {
	v := New(16)
	if v.Test(0) != false {
		t.Errorf("Unable to make a bit set and read its 0th value.")
	}
}

func TestBitSetHuge(t *testing.T) {
	v := New(uint64(math.MaxUint32))
	if v.Test(0) != false {
		t.Errorf("Unable to make a huge bit set and read its 0th value.")
	}
}

// func TestLen(t *testing.T) {
// 	v := New(1000)
// 	if v.Len() != 1000 {
// 		t.Errorf("Len should be 1000, but is %d.", v.Len())
// 	}
// }

func TestBitSetIsClear(t *testing.T) {
	v := New(1000)
	for i := uint64(0); i < 1000; i++ {
		if v.Test(i) != false {
			t.Errorf("Bit %d is set, and it shouldn't be.", i)
		}
	}
}

func TestExendOnBoundary(t *testing.T) {
	v := New(32)
	defer func() {
		if r := recover(); r != nil {
			t.Error("Border out of index error should not have caused a panic")
		}
	}()
	v.Set(32)
}

func TestExpand(t *testing.T) {
	v := New(0)
	defer func() {
		if r := recover(); r != nil {
			t.Error("Expansion should not have caused a panic")
		}
	}()
	for i := uint64(0); i < 1000; i++ {
		v.Set(i)
	}
}

func TestBitSetAndGet(t *testing.T) {
	v := New(1000)
	v.Set(100)
	if v.Test(100) != true {
		t.Errorf("Bit %d is clear, and it shouldn't be.", 100)
	}
}

func TestIterate(t *testing.T) {
	v := New(10000)
	v.Set(0)
	v.Set(1)
	v.Set(2)
	data := make([]uint64, 3)
	c := 0
	for i, e := v.NextSet(0); e; i, e = v.NextSet(i + 1) {
		data[c] = i
		c++
	}
	if data[0] != 0 {
		t.Errorf("bug 0")
	}
	if data[1] != 1 {
		t.Errorf("bug 1")
	}
	if data[2] != 2 {
		t.Errorf("bug 2")
	}
	v.Set(10)
	v.Set(2000)
	data = make([]uint64, 5)
	c = 0
	for i, e := v.NextSet(0); e; i, e = v.NextSet(i + 1) {
		data[c] = i
		c++
	}
	if data[0] != 0 {
		t.Errorf("bug 0")
	}
	if data[1] != 1 {
		t.Errorf("bug 1")
	}
	if data[2] != 2 {
		t.Errorf("bug 2")
	}
	if data[3] != 10 {
		t.Errorf("bug 3")
	}
	if data[4] != 2000 {
		t.Errorf("bug 4")
	}

}

func TestSetTo(t *testing.T) {
	v := New(1000)
	v.SetTo(100, true)
	if v.Test(100) != true {
		t.Errorf("Bit %d is clear, and it shouldn't be.", 100)
	}
	v.SetTo(100, false)
	if v.Test(100) != false {
		t.Errorf("Bit %d is set, and it shouldn't be.", 100)
	}
}

func TestChain(t *testing.T) {
	if New(1000).Set(100).Set(99).Clear(99).Test(100) != true {
		t.Errorf("Bit %d is clear, and it shouldn't be.", 100)
	}
}

func TestOutOfBoundsLong(t *testing.T) {
	v := New(64)
	defer func() {
		if r := recover(); r != nil {
			t.Error("Long distance out of index error should not have caused a panic")
		}
	}()
	v.Set(1000)
}

func TestOutOfBoundsClose(t *testing.T) {
	v := New(65)
	defer func() {
		if r := recover(); r != nil {
			t.Error("Local out of index error should not have caused a panic")
		}
	}()
	v.Set(66)
}

func TestCount(t *testing.T) {
	tot := uint64(64*4 + 11) // just some multi unit64 number
	v := New(tot)
	checkLast := true
	for i := uint64(0); i < tot; i++ {
		sz := uint64(v.Count())
		if sz != i {
			t.Errorf("Count reported as %d, but it should be %d", sz, i)
			checkLast = false
			break
		}
		v.Set(i)
	}
	if checkLast {
		sz := uint64(v.Count())
		if sz != tot {
			t.Errorf("After all bits set, size reported as %d, but it should be %d", sz, tot)
		}
	}
}

// test setting every 3rd bit, just in case something odd is happening
func TestCount2(t *testing.T) {
	tot := uint64(64*4 + 11) // just some multi unit64 number
	v := New(tot)
	for i := uint64(0); i < tot; i += 3 {
		sz := uint64(v.Count())
		if sz != i/3 {
			t.Errorf("Count reported as %d, but it should be %d", sz, i)
			break
		}
		v.Set(i)
	}
}

// nil tests
func TestNullTest(t *testing.T) {
	var v *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Checking bit of null reference should have caused a panic")
		}
	}()
	v.Test(66)
}

func TestNullSet(t *testing.T) {
	var v *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Setting bit of null reference should have caused a panic")
		}
	}()
	v.Set(66)
}

func TestNullClear(t *testing.T) {
	var v *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Clearning bit of null reference should have caused a panic")
		}
	}()
	v.Clear(66)
}

func TestPanicDifferenceBNil(t *testing.T) {
	var b *BitSet
	var compare = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil First should should have caused a panic")
		}
	}()
	b.Difference(compare)
}

func TestPanicDifferenceCompareNil(t *testing.T) {
	var compare *BitSet
	var b = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil Second should should have caused a panic")
		}
	}()
	if b.Difference(compare) == nil {
		panic("empty bitset given")
	}
}

func TestPanicUnionBNil(t *testing.T) {
	var b *BitSet
	var compare = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil First should should have caused a panic")
		}
	}()
	b.Union(compare)
}

func TestPanicUnionCompareNil(t *testing.T) {
	var compare *BitSet
	var b = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil Second should should have caused a panic")
		}
	}()
	if b.Union(compare) == nil {
		panic("empty bitset given")
	}
}

func TestPanicIntersectionBNil(t *testing.T) {
	var b *BitSet
	var compare = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil First should should have caused a panic")
		}
	}()
	b.Intersection(compare)
}

func TestPanicIntersectionCompareNil(t *testing.T) {
	var compare *BitSet
	var b = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil Second should should have caused a panic")
		}
	}()
	if b.Intersection(compare) == nil {
		panic("empty bitset given")
	}
}

func TestPanicSymmetricDifferenceBNil(t *testing.T) {
	var b *BitSet
	var compare = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil First should should have caused a panic")
		}
	}()
	b.SymmetricDifference(compare)
}

func TestPanicSymmetricDifferenceCompareNil(t *testing.T) {
	var compare *BitSet
	var b = New(10)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil Second should should have caused a panic")
		}
	}()
	if b.SymmetricDifference(compare) == nil {
		panic("empty bitset given")
	}
}

func TestPanicComplementBNil(t *testing.T) {
	var b *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil should should have caused a panic")
		}
	}()
	b.Complement()
}

func TestPanicAnytBNil(t *testing.T) {
	var b *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil should should have caused a panic")
		}
	}()
	b.Any()
}

func TestPanicNonetBNil(t *testing.T) {
	var b *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil should should have caused a panic")
		}
	}()
	b.None()
}

func TestPanicAlltBNil(t *testing.T) {
	var b *BitSet
	defer func() {
		if r := recover(); r == nil {
			t.Error("Nil should should have caused a panic")
		}
	}()
	b.All()
}

func TestEqual(t *testing.T) {
	a := New(100)
	// b := New(99)
	c := New(100)
	// if a.Equal(b) {
	// 	t.Error("Sets of different sizes should be not be equal")
	// }
	// if !a.Equal(c) {
	// 	t.Error("Two empty sets of the same size should be equal")
	// }
	a.Set(99)
	c.Set(0)
	if a.Equal(c) {
		t.Error("Two sets with differences should not be equal")
	}
	c.Set(99)
	a.Set(0)
	if !a.Equal(c) {
		t.Error("Two sets with the same bits set should be equal")
	}
}

func TestUnion(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1)
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	na, _ := a.UnionCardinality(b)
	if na != 200 {
		t.Errorf("Union should have 200 bits set, but had %d", na)
	}
	nb, _ := b.UnionCardinality(a)
	if na != nb {
		t.Errorf("Union should be symmetric")
	}

	c := a.Union(b)
	d := b.Union(a)
	if c.Count() != 200 {
		t.Errorf("Union should have 200 bits set, but had %d", c.Count())
	}
	if !c.Equal(d) {
		t.Errorf("Union should be symmetric")
	}
}

func TestInPlaceUnion(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1)
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.Clone()
	c.InPlaceUnion(b)
	d := b.Clone()
	d.InPlaceUnion(a)
	if c.Count() != 200 {
		t.Errorf("Union should have 200 bits set, but had %d", c.Count())
	}
	if d.Count() != 200 {
		t.Errorf("Union should have 200 bits set, but had %d", d.Count())
	}
	if !c.Equal(d) {
		t.Errorf("Union should be symmetric")
	}
}

func TestIntersection(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1).Set(i)
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	na, _ := a.IntersectionCardinality(b)
	if na != 50 {
		t.Errorf("Intersection should have 50 bits set, but had %d", na)
	}
	nb, _ := b.IntersectionCardinality(a)
	if na != nb {
		t.Errorf("Intersection should be symmetric")
	}
	c := a.Intersection(b)
	d := b.Intersection(a)
	if c.Count() != 50 {
		t.Errorf("Intersection should have 50 bits set, but had %d", c.Count())
	}
	if !c.Equal(d) {
		t.Errorf("Intersection should be symmetric")
	}
}

func TestInplaceIntersection(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1).Set(i)
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.Clone()
	c.InPlaceIntersection(b)
	d := b.Clone()
	d.InPlaceIntersection(a)
	e := New(100)
	for i := uint64(76); i < 100; i++ {
		e.Set(i)
	}
	f := a.Clone()
	f.InPlaceIntersection(e)
	if c.Count() != 50 {
		t.Errorf("Intersection should have 50 bits set, but had %d", c.Count())
	}
	if d.Count() != 50 {
		t.Errorf("Intersection should have 50 bits set, but had %d", d.Count())
	}
	if !c.Equal(d) {
		t.Errorf("Intersection should be symmetric")
	}
	if f.Count() != 12 {
		t.Errorf("Intersection should have 12 bits set, but had %d", f.Count())
	}
}

func TestDifference(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1)
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	na, _ := a.DifferenceCardinality(b)
	if na != 50 {
		t.Errorf("a-b Difference should have 50 bits set, but had %d", na)
	}
	nb, _ := b.DifferenceCardinality(a)
	if nb != 150 {
		t.Errorf("b-a Difference should have 150 bits set, but had %d", nb)
	}

	c := a.Difference(b)
	d := b.Difference(a)
	if c.Count() != 50 {
		t.Errorf("a-b Difference should have 50 bits set, but had %d", c.Count())
	}
	if d.Count() != 150 {
		t.Errorf("b-a Difference should have 150 bits set, but had %d", d.Count())
	}
	if c.Equal(d) {
		t.Errorf("Difference, here, should not be symmetric")
	}
}

func TestInPlaceDifference(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)
		b.Set(i - 1)
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.Clone()
	c.InPlaceDifference(b)
	d := b.Clone()
	d.InPlaceDifference(a)
	e := a.Clone()
	for i := uint64(0); i < 100; i += 2 {
		e.Set(i)
	}
	e.InPlaceDifference(b)
	if c.Count() != 50 {
		t.Errorf("a-b Difference should have 50 bits set, but had %d", c.Count())
	}
	if d.Count() != 150 {
		t.Errorf("b-a Difference should have 150 bits set, but had %d", d.Count())
	}
	if c.Equal(d) {
		t.Errorf("Difference, here, should not be symmetric")
	}
	if e.Count() != 50 {
		t.Errorf("e-b Difference should have 50 bits set, but had %d", e.Count())
	}
}

func TestSymmetricDifference(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)            // 01010101010 ... 0000000
		b.Set(i - 1).Set(i) // 11111111111111111000000
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	na, _ := a.SymmetricDifferenceCardinality(b)
	if na != 150 {
		t.Errorf("a^b Difference should have 150 bits set, but had %d", na)
	}
	nb, _ := b.SymmetricDifferenceCardinality(a)
	if nb != 150 {
		t.Errorf("b^a Difference should have 150 bits set, but had %d", nb)
	}

	c := a.SymmetricDifference(b)
	d := b.SymmetricDifference(a)
	if c.Count() != 150 {
		t.Errorf("a^b Difference should have 150 bits set, but had %d", c.Count())
	}
	if d.Count() != 150 {
		t.Errorf("b^a Difference should have 150 bits set, but had %d", d.Count())
	}
	if !c.Equal(d) {
		t.Errorf("SymmetricDifference should be symmetric")
	}
}

func TestInPlaceSymmetricDifference(t *testing.T) {
	a := New(100)
	b := New(200)
	for i := uint64(1); i < 100; i += 2 {
		a.Set(i)            // 01010101010 ... 0000000
		b.Set(i - 1).Set(i) // 11111111111111111000000
	}
	for i := uint64(100); i < 200; i++ {
		b.Set(i)
	}
	c := a.Clone()
	c.InPlaceSymmetricDifference(b)
	d := b.Clone()
	d.InPlaceSymmetricDifference(a)
	if c.Count() != 150 {
		t.Errorf("a^b Difference should have 150 bits set, but had %d", c.Count())
	}
	if d.Count() != 150 {
		t.Errorf("b^a Difference should have 150 bits set, but had %d", d.Count())
	}
	if !c.Equal(d) {
		t.Errorf("SymmetricDifference should be symmetric")
	}
}

func TestComplement(t *testing.T) {
	a := New(50)
	a.Set(50)
	b := a.Complement()
	c := b.Complement()
	if !c.Equal(a) {
		t.Errorf("Complement failed: complement of complement should equal original")
	}
	a = New(50)
	a.Set(10).Set(20).Set(42)
	b = a.Complement()
	c = b.Complement()
	if !c.Equal(a) {
		t.Errorf("Complement failed: complement of complement should equal original")
	}
}

func TestIsSuperSet(t *testing.T) {
	a := New(500)
	b := New(300)
	c := New(200)

	// Setup bitsets
	// a and b overlap
	// only c is (strict) super set
	for i := uint64(0); i < 100; i++ {
		a.Set(i)
	}
	for i := uint64(50); i < 150; i++ {
		b.Set(i)
	}
	for i := uint64(0); i < 200; i++ {
		c.Set(i)
	}

	if a.IsSuperSet(b) == true {
		t.Errorf("IsSuperSet fails")
	}
	if a.IsSuperSet(c) == true {
		t.Errorf("IsSuperSet fails")
	}
	if b.IsSuperSet(a) == true {
		t.Errorf("IsSuperSet fails")
	}
	if b.IsSuperSet(c) == true {
		t.Errorf("IsSuperSet fails")
	}
	if c.IsSuperSet(a) != true {
		t.Errorf("IsSuperSet fails")
	}
	if c.IsSuperSet(b) != true {
		t.Errorf("IsSuperSet fails")
	}

	if a.IsStrictSuperSet(b) == true {
		t.Errorf("IsStrictSuperSet fails")
	}
	if a.IsStrictSuperSet(c) == true {
		t.Errorf("IsStrictSuperSet fails")
	}
	if b.IsStrictSuperSet(a) == true {
		t.Errorf("IsStrictSuperSet fails")
	}
	if b.IsStrictSuperSet(c) == true {
		t.Errorf("IsStrictSuperSet fails")
	}
	if c.IsStrictSuperSet(a) != true {
		t.Errorf("IsStrictSuperSet fails")
	}
	if c.IsStrictSuperSet(b) != true {
		t.Errorf("IsStrictSuperSet fails")
	}
}

// BENCHMARKS

func BenchmarkSet(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 100000
	s := New(uint64(sz))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Set(uint64(r.Int31n(int32(sz))))
	}
}

func BenchmarkGetTest(b *testing.B) {
	b.StopTimer()
	r := rand.New(rand.NewSource(0))
	sz := 100000
	s := New(uint64(sz))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Test(uint64(r.Int31n(int32(sz))))
	}
}

func BenchmarkSetExpand(b *testing.B) {
	b.StopTimer()
	sz := uint64(100000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		var s BitSet
		s.Set(sz)
	}
}

// go test -bench=Count
func BenchmarkCount(b *testing.B) {
	b.StopTimer()
	s := New(100000)
	for i := 0; i < 100000; i += 100 {
		s.Set(uint64(i))
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		s.Count()
	}
}

// go test -bench=Iterate
func BenchmarkIterate(b *testing.B) {
	b.StopTimer()
	s := New(10000)
	for i := 0; i < 10000; i += 3 {
		s.Set(uint64(i))
	}
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c := uint(0)
		for i, e := s.NextSet(0); e; i, e = s.NextSet(i + 1) {
			c++
		}
	}
}

// go test -bench=SparseIterate
func BenchmarkSparseIterate(b *testing.B) {
	b.StopTimer()
	s := New(100000)
	for i := 0; i < 100000; i += 30 {
		s.Set(uint64(i))
	}
	b.StartTimer()
	for j := 0; j < b.N; j++ {
		c := uint(0)
		for i, e := s.NextSet(0); e; i, e = s.NextSet(i + 1) {
			c++
		}
	}
}
