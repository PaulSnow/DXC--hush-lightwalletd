// Copyright (c) 2019-2023 Duke Leto and The Hush developers
// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the GPLv3 software license
// Package parser deserializes (full) transactions from hushd

package parser

// Reverse the given byte slice, returning a slice pointing to new data;
// the input slice is unchanged.
func Reverse(a []byte) []byte {
	r := make([]byte, len(a), len(a))
	for left, right := 0, len(a)-1; left <= right; left, right = left+1, right-1 {
		r[left], r[right] = a[right], a[left]
	}
	return r
}
