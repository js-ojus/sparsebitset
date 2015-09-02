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

package sparsebitset

import (
	"log"
	"math"
)

const (
	// Size of a word -- `uint64` -- in bits.
	wordSize = uint64(64)

	// modWordSize is (`wordSize` - 1).
	modWordSize = wordSize - 1

	// Number of bits to right-shift by, to divide by wordSize.
	log2WordSize = uint64(6)

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

// block is a pair of (offset, mask).
type block struct {
	Offset uint64
	Mask   uint64
}

// setBit sets the bit at the given position.
func (b *block) setBit(n uint64) {
	b.Mask |= 1 << n
}

// clearBit clears the bit at the given position.
func (b *block) clearBit(n uint64) {
	b.Mask &^= 1 << n
}

// flipBit flips the bit at the given position.
func (b *block) flipBit(n uint64) {
	b.Mask ^= 1 << n
}

// testBit checks to see if the bit at the given position is set.
func (b *block) testBit(n uint64) bool {
	return (b.Mask & (1 << n)) > 0
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
func (a blockAry) insert(b block, idx uint64) (blockAry, error) {
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
func (a blockAry) delete(idx uint64) (blockAry, error) {
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
	if n == 0 {
		return a, ErrInvalidIndex
	}

	idx := n
	off := idx >> log2WordSize
	bit := idx & modWordSize

	i := -1
	for j, el := range a {
		if el.Offset > off {
			i = j
			break
		}
		if el.Offset == off {
			el.setBit(bit)
			return a, nil
		}
	}
	if i == -1 { // all blocks (if any) have smaller offsets
		i = len(a)
	}

	return a.insert(block{off, 1 << bit}, uint64(i))
}

// clearBit sets the bit at the given position to `0`.
func (a blockAry) clearBit(n uint64) (blockAry, error) {
	if n == 0 {
		return a, ErrInvalidIndex
	}

	idx := n
	off := idx >> log2WordSize
	bit := idx & modWordSize

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
	if popcount(a[i].Mask) == 0 {
		return a.delete(uint64(i))
	}
	return a, nil
}

// flipBit inverts the bit at the given position.
func (a blockAry) flipBit(n uint64) (blockAry, error) {
	if n == 0 {
		return a, ErrInvalidIndex
	}

	idx := n
	off := idx >> log2WordSize
	bit := idx & modWordSize

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

	idx := n
	off := idx >> log2WordSize
	bit := idx & modWordSize

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
// BitSet is **not** thread-safe!
func New(n uint64) *BitSet {
	if n == 0 {
		return nil
	}

	dens := bitDensity * float64(n)
	dens = math.Min(1.0, dens)
	return &BitSet{make(blockAry, 0, uint64(dens))}
}

// Len answers the number of bytes used by this bitset.
func (b *BitSet) Len() int {
	return len(b.set) * 16
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
		log.Println(err)
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
// Usage:
//   for idx, ok := set.NextSet(0); ok; idx, ok = set.NextSet(idx+1) {
//       ...
//   }
func (b *BitSet) NextSet(n uint64) (uint64, bool) {
	idx := n
	off := idx >> log2WordSize

	i := -1
	higher := false
	for j, el := range b.set {
		if el.Offset == off {
			i = j
			break
		}
		if el.Offset > off {
			i = j
			higher = true
			break
		}
	}
	if i == -1 { // given bit is larger than the largest in the set
		return 0, false
	}

	if !higher {
		w := b.set[i].Mask >> (n & modWordSize)
		if w > 0 {
			return n + trailingZeroes64(w), true
		}
	}
	return (off * wordSize) + trailingZeroes64(b.set[i].Mask), true
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
	copy(c.set, b.set)
	return &c
}

// Copy copies this bitset into the destination bitset.  It answers
// the size of the destination bitset.
func (b *BitSet) Copy(c *BitSet) int {
	if c == nil || len(c.set) == 0 {
		return 0
	}
	if len(c.set)%2 == 1 { // we need to store (offset, mask) pairs
		return -1
	}

	return copy(c.set, b.set)
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
		if el.Offset != cel.Offset || el.Mask != cel.Mask {
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
			if b.set[j].Mask == 0 {
				i = j
				break
			}
		}
		if i > -1 {
			b.set = append(b.set[:i], b.set[i+1:]...)
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

	res := b.Clone()
	l := len(b.set)
	lc := len(c.set)
	if lc < l {
		l = lc
	}
	for i := 0; i < l; i++ {
		res.set[i].Mask = b.set[i].Mask &^ c.set[i].Mask
	}
	res.prune()
	return res
}

// DifferenceCardinality answers the cardinality of the difference set
// between this bitset and the given bitset.  This does *not*
// construct an intermediate bitset.
func (b *BitSet) DifferenceCardinality(c *BitSet) (uint64, error) {
	if c == nil {
		return 0, ErrNilArgument
	}

	res := uint64(0)
	l := len(b.set)
	lc := len(c.set)
	if lc < l {
		l = lc
	}
	res += popcountSetMasked(b.set[:l], c.set[:l])
	return res, nil
}
