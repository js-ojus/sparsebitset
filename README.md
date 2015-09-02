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

## `sparsebitset`
A simple implementation of sparse bitsets for positive integers.

It is being extracted from an implementation of custom database indexes for attributes of semi-structured documents.

The representation is very simple, and uses a sequence of (offset, mask) pairs.  It is similar to that of Go's `x/tools/container/intsets` and Java's `java.util.BitSet`.  However, Go's package caters to negative integers as well, which I do not need.  Also, I needed a simple way to serialise and deserialise the sets to/from `[]byte`, because I am using them to store custom indexes in a database.

## `sparsebitset` vs. `willf/bitset`
`sparsebitset` should be useful for sparse sets for which the space overhead of [BitSet](https://github.com/willf/bitset) may be high.  Please test with your real life data - in reasonable volumes - before choosing one.

However, `sparsebitset` tries to provide an API that is mostly similar to that of [BitSet](https://github.com/willf/bitset).  Users can switch between the two implementations with a little effort, based on the evolution (either dense --> sparse or sparse --> dense) of their data.

Here are a few of the differences to note.

* `sparsebitset` operates with `uint64` rather than `uint` almost everywhere.  This simplifies several parts of the code.
* `sparsebitset` does not panic.  Upon encountering errors, it returns `nil` where `*BitSet` is expected.  Elsewhere, it returns an additional `error` value that must be checked.
