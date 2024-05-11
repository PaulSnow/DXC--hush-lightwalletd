// Copyright (c) 2019-2024 Duke Leto and The Hush developers
// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the GPLv3 software license
// +build gofuzz

package parser

func Fuzz(data []byte) int {
	block := NewBlock()
	_, err := block.ParseFromSlice(data)
	if err != nil {
		return 0
	}
	return 1
}
