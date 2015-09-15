// (c) Copyright 2015 JONNALAGADDA Srinivas
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package sparsebitset is a simple implementation of sparse bitsets
// for non-negative integers.
//
// The representation is very simple, and uses a sequence of (offset,
// bits) pairs.  It is similar to that of Go's
// `x/tools/container/intsets` and Java's `java.util.BitSet`.
// However, Go's package caters to negative integers as well, which I
// do not need.
//
// The original motivation for `sparsebitset` comes from a need to
// store custom indexes of documents in a database.  Accordingly,
// `sparsebitset` trades CPU time for space.
package sparsebitset

import (
	"encoding/binary"
	"io"
	"log"
)

const (
	// Size of a word -- `uint64` -- in bits.
	wordSize = uint64(64)

	// modWordSize is (`wordSize` - 1).
	modWordSize = wordSize - 1

	// Number of bits to right-shift by, to divide by wordSize.
	log2WordSize = uint64(6)

	// allOnes is a word with all bits set to `1`.
	allOnes uint64 = 0xffffffffffffffff

	// Density of bits, expressed as a fraction of the total space.
	bitDensity = 0.1
)

var deBruijn = [...]byte{
	0, 1, 56, 2, 57, 49, 28, 3, 61, 58, 42, 50, 38, 29, 17, 4,
	62, 47, 59, 36, 45, 43, 51, 22, 53, 39, 33, 30, 24, 18, 12, 5,
	63, 55, 48, 27, 60, 41, 37, 16, 46, 35, 44, 21, 52, 32, 23, 11,
	54, 26, 40, 15, 34, 20, 31, 10, 25, 14, 19, 9, 13, 8, 7, 6,
}

func trailingZeroes64(v uint64) uint64 {
	return uint64(deBruijn[((v&-v)*0x03f79d71b4ca8b09)>>58])
}

func offsetBits(n uint64) (uint64, uint64) {
	return n >> log2WordSize, n & modWordSize
}

// block is a pair of (offset, mask).
type block struct {
	Offset uint64
	Bits   uint64
}

// setBit sets the bit at the given position.
func (b *block) setBit(n uint64) {
	b.Bits |= 1 << n
}

// clearBit clears the bit at the given position.
func (b *block) clearBit(n uint64) {
	b.Bits &^= 1 << n
}

// flipBit flips the bit at the given position.
func (b *block) flipBit(n uint64) {
	b.Bits ^= 1 << n
}

// testBit checks to see if the bit at the given position is set.
func (b *block) testBit(n uint64) bool {
	return (b.Bits & (1 << n)) > 0
}

// blockAry makes manipulation of blocks easier.  It is also
// `sort.Sort`able.
type blockAry []block

// Len answers the number of blocks in this slice.
func (a blockAry) Len() int {
	return len(a)
}

// Less answers if the element at the first index is less than that at
// the second index given.
func (a blockAry) Less(i, j int) bool {
	return a[i].Offset < a[j].Offset
}

// Swap exchanges the elements at the given indices.
func (a blockAry) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// insert inserts the given block at the specified location.
func (a blockAry) insert(b block, idx uint32) (blockAry, error) {
	l := len(a)
	if int(idx) >= l {
		a = append(a, b)
		return a, nil
	}

	t := make([]block, 0, l+1)
	if idx > 0 {
		copy(t, a[:idx])
	}
	t = append(t, b)
	t = append(t, a[idx:]...)

	return t, nil
}

// delete removes the block at the specified location.
func (a blockAry) delete(idx uint32) (blockAry, error) {
	if int(idx) >= len(a) {
		return a, ErrInvalidIndex
	}
	if idx == 0 {
		return a[1:], nil
	}

	a = append(a[:idx], a[idx+1:]...)
	return a, nil
}

// setBit sets the bit at the given position to `1`.
func (a blockAry) setBit(n uint64) (blockAry, error) {
	off, bit := offsetBits(n)

	i := -1
	for j, el := range a {
		if el.Offset > off {
			i = j
			break
		}
		if el.Offset == off {
			el.setBit(bit)
			a[j] = el
			return a, nil
		}
	}
	if i == -1 { // all blocks (if any) have smaller offsets
		i = len(a)
	}

	return a.insert(block{off, 1 << bit}, uint32(i))
}

// clearBit sets the bit at the given position to `0`.
func (a blockAry) clearBit(n uint64) (blockAry, error) {
	off, bit := offsetBits(n)

	i := -1
	for j, el := range a {
		if el.Offset == off {
			i = j
			break
		}
	}
	if i == -1 { // nothing to do
		return a, nil
	}

	a[i].clearBit(bit)
	if a[i].Bits == 0 {
		return a.delete(uint32(i))
	}
	return a, nil
}

// flipBit inverts the bit at the given position.
func (a blockAry) flipBit(n uint64) (blockAry, error) {
	off, bit := offsetBits(n)

	i := -1
	for j, el := range a {
		if el.Offset == off {
			i = j
			break
		}
	}
	if i == -1 {
		return a, ErrItemNotFound
	}

	a[i].flipBit(bit)
	return a, nil
}

// testBit answers `true` if the bit at the given position is set;
// `false` otherwise.
func (a blockAry) testBit(n uint64) bool {
	if n == 0 {
		return false
	}

	off, bit := offsetBits(n)

	i := -1
	for j, el := range a {
		if el.Offset == off {
			i = j
			break
		}
	}
	if i == -1 {
		return false
	}

	return a[i].testBit(bit)
}

// BitSet is a compact representation of sparse positive integer sets.
type BitSet struct {
	set blockAry
}

// New creates a new BitSet using the given size hint.
//
// `BitSet` is **not** thread-safe!
func New(n uint64) *BitSet {
	// dens := bitDensity * float64(n)
	// if dens < 1.0 {
	// 	dens = 1.0
	// }
	return &BitSet{make(blockAry, 0, 1)}
}

// Len answers the number of bytes used by this bitset.
func (b *BitSet) Len() int {
	return len(b.set) * binary.Size(uint64(0))
}

// Test answers `true` if the bit at the given position is set;
// `false` otherwise.
func (b *BitSet) Test(n uint64) bool {
	return b.set.testBit(n)
}

// Set sets the bit at the given position to `1`.
func (b *BitSet) Set(n uint64) *BitSet {
	ary, err := b.set.setBit(n)
	if err != nil {
		log.Println(err, ":", n)
		return nil
	}

	b.set = ary
	return b
}

// Clear sets the bit at the given position to `0`.
func (b *BitSet) Clear(n uint64) *BitSet {
	ary, err := b.set.clearBit(n)
	if err != nil {
		log.Println(err)
		return nil
	}

	b.set = ary
	return b
}

// SetTo sets the bit at the given position to the given value.
func (b *BitSet) SetTo(n uint64, val bool) *BitSet {
	if val {
		return b.Set(n)
	}
	return b.Clear(n)
}

// Flip inverts the bit at the given position.
func (b *BitSet) Flip(n uint64) *BitSet {
	ary, err := b.set.flipBit(n)
	if err != nil {
		log.Println(err)
		return nil
	}

	b.set = ary
	return b
}

// NextSet answers the next bit that is set, starting with (and
// including) the given index.  The boolean part of the output tuple
// indicates the presence (`true`) or absence (`false`) of such a bit
// in this bitset.
//
// Example usage:
//   for idx, ok := set.NextSet(0); ok; idx, ok = set.NextSet(idx+1) {
//       ...
//   }
func (b *BitSet) NextSet(n uint64) (uint64, bool) {
	off, rsh := offsetBits(n)

	i := -1
	for j, el := range b.set {
		if el.Offset == off {
			w := el.Bits >> rsh
			if w > 0 {
				return n + trailingZeroes64(w), true
			}
		}
		if el.Offset > off {
			i = j
			break
		}
	}
	if i == -1 {
		return 0, false
	}

	return (b.set[i].Offset * wordSize) + trailingZeroes64(b.set[i].Bits), true
}

// ClearAll resets this bitset.
func (b *BitSet) ClearAll() *BitSet {
	b.set = b.set[:0]
	return b
}

// Clone answers a copy of this bitset.
func (b *BitSet) Clone() *BitSet {
	var c BitSet
	c.set = make(blockAry, 0, len(b.set))
	for _, el := range b.set {
		c.set = append(c.set, el)
	}
	return &c
}

// Copy copies this bitset into the destination bitset.  It answers
// the size of the destination bitset.
func (b *BitSet) Copy(c *BitSet) int {
	if c == nil {
		return 0
	}

	ctr := 0
	for _, el := range b.set {
		c.set = append(c.set, el)
		ctr++
	}
	return ctr * 2 * binary.Size(uint64(0))
}

// Count is an alias for `Cardinality`.
func (b *BitSet) Count() uint64 {
	return b.Cardinality()
}

// Cardinality answers the number of bits in this bitset that are set
// to `1`.
func (b *BitSet) Cardinality() uint64 {
	return popcountSet(b.set)
}

// Equal answers `true` iff the two sets have the same bits set to
// `1`.
func (b *BitSet) Equal(c *BitSet) bool {
	if c == nil {
		return false
	}
	lb := len(b.set)
	if lb != len(c.set) {
		return false
	}
	if lb == 0 { // both are empty
		return true
	}

	for i, el := range b.set {
		cel := c.set[i]
		if el.Offset != cel.Offset || el.Bits != cel.Bits {
			return false
		}
	}
	return true
}

// prune removes empty blocks from this bitset.
func (b *BitSet) prune() {
	chg := true
	resume := 0

	for chg {
		chg = false
		i := -1
		for j := resume; j < len(b.set); j++ {
			if b.set[j].Bits == 0 {
				i = j
				break
			}
		}
		if i > -1 {
			b.set, _ = b.set.delete(uint32(i))
			chg = true
			resume = i
		}
	}
}

// Difference performs a 'set minus' of the given bitset from this
// bitset.
func (b *BitSet) Difference(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	res := new(BitSet)
	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for i < lb && j < lc {
		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			res.set = append(res.set, bbl)
			i++

		case bbl.Offset == cbl.Offset:
			var t block
			t.Offset = bbl.Offset
			t.Bits = bbl.Bits &^ cbl.Bits
			res.set = append(res.set, t)
			i, j = i+1, j+1

		default:
			j++
		}
	}
	for ; i < lb; i++ {
		res.set = append(res.set, b.set[i])
	}

	res.prune()
	return res
}

// InPlaceDifference performs a 'set minus' of the given bitset from
// this bitset, updating this bitset itself.
func (b *BitSet) InPlaceDifference(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for i < lb && j < lc {
		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			i++

		case bbl.Offset == cbl.Offset:
			bbl.Bits &^= cbl.Bits
			b.set[i] = bbl
			i, j = i+1, j+1

		default:
			j++
		}
	}

	b.prune()
	return b
}

// DifferenceCardinality answers the cardinality of the difference set
// between this bitset and the given bitset.  This does *not*
// construct an intermediate bitset.
func (b *BitSet) DifferenceCardinality(c *BitSet) (uint64, error) {
	if c == nil {
		return 0, ErrNilArgument
	}

	return popcountSetAndNot(b.set, c.set), nil
}

// Intersection performs a 'set intersection' of the given bitset with
// this bitset.
func (b *BitSet) Intersection(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	res := new(BitSet)
	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for i < lb && j < lc {
		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			i++

		case bbl.Offset == cbl.Offset:
			var t block
			t.Offset = bbl.Offset
			t.Bits = bbl.Bits & cbl.Bits
			res.set = append(res.set, t)
			i, j = i+1, j+1

		default:
			j++
		}
	}

	res.prune()
	return res
}

// InPlaceIntersection performs a 'set intersection' of the given
// bitset with this bitset, updating this bitset itself.
func (b *BitSet) InPlaceIntersection(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for i < lb && j < lc {
		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			b.set[i].Bits = 0
			i++

		case bbl.Offset == cbl.Offset:
			bbl.Bits &= cbl.Bits
			b.set[i] = bbl
			i, j = i+1, j+1

		default:
			j++
		}
	}
	for ; i < lb; i++ {
		b.set[i].Bits = 0
	}

	b.prune()
	return b
}

// IntersectionCardinality answers the cardinality of the intersection
// set between this bitset and the given bitset.  This does *not*
// construct an intermediate bitset.
func (b *BitSet) IntersectionCardinality(c *BitSet) (uint64, error) {
	if c == nil {
		return 0, ErrNilArgument
	}

	return popcountSetAnd(b.set, c.set), nil
}

// Union performs a 'set union' of the given bitset with this bitset.
func (b *BitSet) Union(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	res := new(BitSet)
	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for i < lb && j < lc {
		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			res.set = append(res.set, bbl)
			i++

		case bbl.Offset == cbl.Offset:
			var t block
			t.Offset = bbl.Offset
			t.Bits = bbl.Bits | cbl.Bits
			res.set = append(res.set, t)
			i, j = i+1, j+1

		default:
			res.set = append(res.set, cbl)
			j++
		}
	}
	for ; i < lb; i++ {
		res.set = append(res.set, b.set[i])
	}
	for ; j < lc; j++ {
		res.set = append(res.set, c.set[j])
	}

	return res
}

// InPlaceUnion performs a 'set union' of the given bitset with this
// bitset, updating this bitset itself.
func (b *BitSet) InPlaceUnion(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for {
		if i >= lb || j >= lc {
			break
		}

		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			i++

		case bbl.Offset == cbl.Offset:
			bbl.Bits |= cbl.Bits
			b.set[i] = bbl
			i, j = i+1, j+1

		default:
			b.set, _ = b.set.insert(cbl, uint32(i))
			lb++
			i, j = i+1, j+1
		}
	}
	for ; j < lc; j++ {
		b.set = append(b.set, c.set[j])
	}

	return b
}

// UnionCardinality answers the cardinality of the union set between
// this bitset and the given bitset.  This does *not* construct an
// intermediate bitset.
func (b *BitSet) UnionCardinality(c *BitSet) (uint64, error) {
	if c == nil {
		return 0, ErrNilArgument
	}

	return popcountSetOr(b.set, c.set), nil
}

// SymmetricDifference performs a 'set symmetric difference' of the
// given bitset with this bitset.
func (b *BitSet) SymmetricDifference(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	res := new(BitSet)
	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for i < lb && j < lc {
		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			res.set = append(res.set, bbl)
			i++

		case bbl.Offset == cbl.Offset:
			var t block
			t.Offset = bbl.Offset
			t.Bits = bbl.Bits ^ cbl.Bits
			res.set = append(res.set, t)
			i, j = i+1, j+1

		default:
			res.set = append(res.set, cbl)
			j++
		}
	}
	for ; i < lb; i++ {
		res.set = append(res.set, b.set[i])
	}
	for ; j < lc; j++ {
		res.set = append(res.set, c.set[j])
	}

	res.prune()
	return res
}

// InPlaceSymmetricDifference performs a 'set symmetric difference' of
// the given bitset with this bitset, updating this bitset itself.
func (b *BitSet) InPlaceSymmetricDifference(c *BitSet) *BitSet {
	if c == nil {
		return nil
	}

	lb := len(b.set)
	lc := len(c.set)
	i, j := 0, 0
	for i < lb && j < lc {
		bbl, cbl := b.set[i], c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			i++

		case bbl.Offset == cbl.Offset:
			bbl.Bits ^= cbl.Bits
			b.set[i] = bbl
			i, j = i+1, j+1

		default:
			b.set, _ = b.set.insert(cbl, uint32(i))
			j++
		}
	}
	for ; j < lc; j++ {
		b.set = append(b.set, c.set[j])
	}

	b.prune()
	return b
}

// SymmetricDifferenceCardinality answers the cardinality of the
// symmetric difference set between this bitset and the given bitset.
// This does *not* construct an intermediate bitset.
func (b *BitSet) SymmetricDifferenceCardinality(c *BitSet) (uint64, error) {
	if c == nil {
		return 0, ErrNilArgument
	}

	return popcountSetXor(b.set, c.set), nil
}

// Complement answers a bit-wise complement of this bitset, up to the
// highest bit set in this bitset.
//
// N.B. Since bitset is not bounded, `a.complement().complement() !=
// a`.  This limits the usefulness of this operation.  Use with care!
func (b *BitSet) Complement() *BitSet {
	res := new(BitSet)

	lb := len(b.set)
	if lb == 0 {
		return res
	}

	off := uint64(0)
	for i, el := range b.set {
		for off < el.Offset {
			res.set = append(res.set, block{off, allOnes})
			off++
		}

		if i < lb-1 {
			res.set = append(res.set, block{el.Offset, ^el.Bits})
			off++
		}
	}
	res.set = append(res.set, b.set[lb-1])

	rel := res.set[len(res.set)-1]
	j := uint64(1)
	for (rel.Bits >> j) > 0 {
		j++
	}
	rel.Bits = rel.Bits << (64 - j)
	rel.Bits = ^rel.Bits >> (64 - j)
	res.set[len(res.set)-1] = rel

	// '0'th bit should be ignored.
	rel = res.set[0]
	rel.Bits = rel.Bits >> 1
	rel.Bits = rel.Bits << 1
	res.set[0] = rel

	res.prune()
	return res
}

// All answers `true` if all the bits in it, up to its highest set
// bit, are set to `1`; `false` otherwise.
func (b *BitSet) All() bool {
	lb := len(b.set)
	if lb == 0 {
		return true // is this correct?
	}

	off := uint64(0)
	for i, el := range b.set[:lb-1] {
		if el.Offset != off {
			return false
		}
		if el.Offset > 0 && i < lb-1 {
			if el.Bits != allOnes {
				return false
			}
		}

		off++
	}

	sel := b.set[lb-1]
	w := uint64(0)
	cp := popcount(sel.Bits)
	if sel.Offset == 0 { // handle '0'th bit
		cp++
		w = (sel.Bits | 1) & allOnes
	} else {
		w = sel.Bits ^ allOnes
	}
	tz := trailingZeroes64(w)
	if cp != tz {
		return false
	}

	return true
}

// IsEmpty answers `true` if this bitset is empty; `false` otherwise.
func (b *BitSet) IsEmpty() bool {
	return len(b.set) == 0
}

// None is an alias for `IsEmpty`.
func (b *BitSet) None() bool {
	return b.IsEmpty()
}

// Any answers `true` iff this bitset is not empty.
func (b *BitSet) Any() bool {
	return !b.IsEmpty()
}

// IsSuperSet answers `true` if this bitset includes all of the given
// bitset's elements.
func (b *BitSet) IsSuperSet(c *BitSet) bool {
	if c == nil || len(c.set) == 0 {
		return true
	}

	n, _ := c.DifferenceCardinality(b)
	if n > 0 {
		return false
	}
	return true
}

// IsStrictSuperSet answers `true` if this bitset is a superset of the
// given bitset, and includes at least one additional element.
func (b *BitSet) IsStrictSuperSet(c *BitSet) bool {
	lb := len(b.set)
	lc := len(c.set)
	if lb < lc {
		return false
	}

	i, j := 0, 0
	for i < lb && j < lc {
		bbl := b.set[i]
		cbl := c.set[j]

		switch {
		case bbl.Offset < cbl.Offset:
			i++

		case bbl.Offset == cbl.Offset:
			if cbl.Bits&^bbl.Bits > 0 {
				return false
			}
			i, j = i+1, j+1

		default:
			return false
		}
	}
	if j < lc { // `b` got exhausted before `c`
		return false
	}

	return true
}

// BinaryStorageSize answers the number of bytes that will be needed
// to serialise this bitset.
func (b *BitSet) BinaryStorageSize() int {
	return binary.Size(uint32(0)) + binary.Size(b.set)
}

// WriteTo serialises this bitset to the given `io.Writer`.
func (b *BitSet) WriteTo(w io.Writer) (int64, error) {
	var err error

	// Write length of the data to follow.
	lb := len(b.set)
	lb *= 2 * binary.Size(uint64(0))
	err = binary.Write(w, binary.BigEndian, uint32(lb))
	if err != nil {
		return 0, err
	}

	err = binary.Write(w, binary.BigEndian, b.set)
	if err != nil {
		return int64(binary.Size(uint32(0))), err
	}

	return int64(b.BinaryStorageSize()), nil
}

// ReadFrom de-serialises the data from the given `io.Reader` stream
// into this bitset.
//
// N.B. This method overwrites the data currently in this bitset.
func (b *BitSet) ReadFrom(r io.Reader) (int64, error) {
	var err error

	// Read length of the data that follows.
	var lb uint32
	err = binary.Read(r, binary.BigEndian, &lb)
	if err != nil {
		return 0, err
	}

	n := int(lb) / (2 * binary.Size(uint64(0)))
	set := make(blockAry, 0, n)
	err = binary.Read(r, binary.BigEndian, &set)
	if err != nil {
		return int64(binary.Size(uint32(0))), err
	}

	b.set = set
	return int64(b.BinaryStorageSize()), nil
}
