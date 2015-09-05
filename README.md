<!--
   (c) Copyright 2015 JONNALAGADDA Srinivas

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
-->

[![Build Status](https://travis-ci.org/js-ojus/sparsebitset.svg?branch=master)](https://travis-ci.org/js-ojus/sparsebitset)

### `sparsebitset`
A simple implementation of sparse bitsets for non-negative integers.

The representation is very simple, and uses a sequence of (offset, bits) pairs.  It is similar to that of Go's `x/tools/container/intsets` and Java's `java.util.BitSet`.  However, Go's package caters to negative integers as well, which I do not need.

The original motivation for `sparsebitset` comes from a need to store custom indexes of documents in a database.  Accordingly, `sparsebitset` trades CPU time for space.

### `sparsebitset` vs. `github.com/willf/bitset`
`sparsebitset` may be useful for sparse sets for which the space overhead of `bitset` ([BitSet](https://github.com/willf/bitset)) may be high.

For purely in-memory operations, when adequate memory is available, `bitset` usually performs much better than `sparsebitset`.  In particular, when data is dense or periodic with a short period, `bitset` is a better choice.  Please test with your real life data - in reasonable volumes - before choosing the package that is more suitable for you.

`sparsebitset` tries to provide an API that is mostly similar to that of `bitset`.  Users can switch between the two implementations with a little effort, based on the evolution (either dense --> sparse or sparse --> dense) of their data.

Here are a few differences to note.

* `sparsebitset` operates with `uint64` rather than `uint` almost everywhere.  This makes several parts of the code uniform.
* `sparsebitset` does not panic.  Upon encountering errors, it returns `nil` where `*BitSet` is expected.  Elsewhere, it returns an additional `error` value that must be checked.
* A few methods are not implemented.  Examples include JSON (de)serialisation methods.

The tests have been adopted from those for `bitset`, and modified appropriately to account for the small API differences.  Therefore, the tests are governed by the license of `bitset`.

### Installation
`sparsebitset` has no external dependencies.

`go get -v 'github.com/js-ojus/sparsebitset'`

### Status
`sparsebitset` passes all applicable tests of `bitset`, and is hence reasonably correct.  Please report an issue, should you encounter any incorrect behaviour.

`sparsebitset` has not been optimised in any way, yet.  All help is highly appreciated!

### Usage
Please see the tests for several examples of usage.  Here is a quick example.

```go
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
```

Here is another.

```go
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
```
