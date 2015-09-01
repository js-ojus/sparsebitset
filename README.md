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

## sparsebitset
A simple implementation of sparse bitsets for positive integers.

It is being extracted from an implementation of custom database indexes for attributes of semi-structured documents.

The representation is very simple, and uses a sequence of (offset, mask) pairs.  It is similar to that of Go's `x/tools/container/intsets` and Java's `java.util.BitSet`.  However, Go's package caters to negative integers as well, which I do not need.  Also, I need a simple way to serialise and deserialise the sets to/from `[]byte`, because I am using them to store custom indexes in a database.

`sparsebitset` should be useful for sparse sets for which the space overhead of [BitSet](https://github.com/willf/bitset) may be high.

However, `sparsebitset` tries to be API-compatible with [BitSet](https://github.com/willf/bitset), so that users can switch between the two implementations based on the evolution (either dense --> sparse or sparse --> dense) of their data.
