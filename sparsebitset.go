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

import "log"

const (
	// Size of a word -- `uint64` -- in bits.
	wordSize = uint64(64)

	// Number of bits to right-shift by, to divide by wordSize.
	log2WordSize = uint(6)

	// Density of bits, expressed as a fraction of the total space.
	bitDensity = 0.1
)

// Bit population count (Hamming Weight), taken from
// https://code.google.com/p/go/issues/detail?id=4988#c11.  Original
// by 'https://code.google.com/u/arnehormann/'.
func popcount(x uint64) (n uint64) {
	x -= (x >> 1) & 0x5555555555555555
	x = (x>>2)&0x3333333333333333 + x&0x3333333333333333
	x += x >> 4
	x &= 0x0f0f0f0f0f0f0f0f
	x *= 0x0101010101010101
	return x >> 56
}

// block is a pair of (offset, mask).
type block struct {
	Offset uint64
	Mask   uint64
}

// setBit sets the bit at the given position.
func (b *block) setBit(n uint) {
	b.Mask |= uint64(1 << n)
}

// clearBit clears the bit at the given position.
func (b *block) clearBit(n uint) {
	b.Mask &^= uint64(1 << n)
}

// flipBit flips the bit at the given position.
func (b *block) flipBit(n uint) {
	b.Mask ^= uint64(1 << n)
}

// testBit checks to see if the bit at the given position is set.
func (b *block) testBit(n uint) bool {
	return (b.Mask & uint64(1<<n)) > 0
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
func (a blockAry) insert(b block, idx uint) (blockAry, error) {
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
func (a blockAry) delete(idx uint) (blockAry, error) {
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
func (a blockAry) setBit(n uint) (blockAry, error) {
	if n == 0 {
		return a, ErrInvalidIndex
	}

	idx := uint64(n)
	off := idx >> log2WordSize
	bit := uint(idx % wordSize)

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
	if i == -1 { // All blocks (if any) have smaller offsets.
		i = len(a)
	}

	return a.insert(block{off, 1 << bit}, uint(i))
}

// clearBit sets the bit at the given position to `0`.
func (a blockAry) clearBit(n uint) (blockAry, error) {
	if n == 0 {
		return a, ErrInvalidIndex
	}

	idx := uint64(n)
	off := idx >> log2WordSize
	bit := uint(idx % wordSize)

	i := -1
	for j, el := range a {
		if el.Offset == off {
			i = j
			break
		}
	}
	if i == -1 { // Nothing to do.
		return a, nil
	}

	a[i].clearBit(bit)
	if popcount(a[i].Mask) == 0 {
		return a.delete(uint(i))
	}
	return a, nil
}

// flipBit inverts the bit at the given position.
func (a blockAry) flipBit(n uint) (blockAry, error) {
	if n == 0 {
		return a, ErrInvalidIndex
	}

	idx := uint64(n)
	off := idx >> log2WordSize
	bit := uint(idx % wordSize)

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
func (a blockAry) testBit(n uint) bool {
	if n == 0 {
		return false
	}

	idx := uint64(n)
	off := idx >> log2WordSize
	bit := uint(idx % wordSize)

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
func New(n uint) *BitSet {
	if n == 0 {
		return nil
	}

	return &BitSet{make([]block, 0, uint(bitDensity*float64(n)))}
}

// Len answers the number of bytes used by this bitset.
func (b *BitSet) Len() int {
	return len(b.set) * 16
}

// Test answers `true` if the bit at the given position is set;
// `false` otherwise.
func (b *BitSet) Test(n uint) bool {
	return b.set.testBit(n)
}

// Set sets the bit at the given position to `1`.
func (b *BitSet) Set(n uint) *BitSet {
	ary, err := b.set.setBit(n)
	if err != nil {
		log.Println(err)
		return nil
	}

	b.set = ary
	return b
}

// Clear sets the bit at the given position to `0`.
func (b *BitSet) Clear(n uint) *BitSet {
	ary, err := b.set.clearBit(n)
	if err != nil {
		log.Println(err)
		return nil
	}

	b.set = ary
	return b
}

// SetTo sets the bit at the given position to the given value.
func (b *BitSet) SetTo(n uint, val bool) *BitSet {
	if val {
		return b.Set(n)
	}
	return b.Clear(n)
}

// Flip inverts the bit at the given position.
func (b *BitSet) Flip(n uint) *BitSet {
	ary, err := b.set.flipBit(n)
	if err != nil {
		log.Println(err)
		return nil
	}

	b.set = ary
	return b
}
