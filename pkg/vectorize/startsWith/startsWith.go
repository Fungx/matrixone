// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package startsWith

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/container/types"
)

var (
	startsWith           func(*types.Bytes, *types.Bytes, []uint8) []uint8
	startsWithRightConst func(*types.Bytes, *types.Bytes, []uint8) []uint8
	startsWithLeftConst  func(*types.Bytes, *types.Bytes, []uint8) []uint8
	startsWithAllConst   func(*types.Bytes, *types.Bytes, []uint8) []uint8
)

func init() {
	startsWith = StartsWith
	startsWithRightConst = StartsWithRightConst
	startsWithLeftConst = StartsWithLeftConst
	startsWithAllConst = StartsWithAllConst
}

func hasPrefix(b1, b2 []byte, offset1, offset2 uint32, len1, len2 uint32) uint8 {
	if len1 >= len2 && bytes.Equal(b1[offset1:offset1+len2], b2[offset2:offset2+len2]) {
		return 1
	}
	return 0
}

func StartsWith(lv, rv *types.Bytes, rs []uint8) []uint8 {
	for i := range lv.Offsets {
		lvCursor, lvLen := lv.Offsets[i], lv.Lengths[i]
		rvCursor, rvLen := rv.Offsets[i], rv.Lengths[i]

		rs[i] = hasPrefix(lv.Data, rv.Data, lvCursor, rvCursor, lvLen, rvLen)
	}

	return rs
}

func StartsWithRightConst(lv, rv *types.Bytes, rs []uint8) []uint8 {
	rvCursor, rvLen := rv.Offsets[0], rv.Lengths[0]
	for i := range lv.Offsets {
		lvCursor, lvLen := lv.Offsets[i], lv.Lengths[i]

		rs[i] = hasPrefix(lv.Data, rv.Data, lvCursor, rvCursor, lvLen, rvLen)
	}

	return rs
}

func StartsWithLeftConst(lv, rv *types.Bytes, rs []uint8) []uint8 {
	lvCursor, lvLen := lv.Offsets[0], lv.Lengths[0]
	for i := range rv.Offsets {
		rvCursor, rvLen := rv.Offsets[i], rv.Lengths[i]

		rs[i] = hasPrefix(lv.Data, rv.Data, lvCursor, rvCursor, lvLen, rvLen)
	}

	return rs
}

func StartsWithAllConst(lv, rv *types.Bytes, rs []uint8) []uint8 {
	lvCursor, lvLen := lv.Offsets[0], lv.Lengths[0]
	rvCursor, rvLen := rv.Offsets[0], rv.Lengths[0]
	rs[0] = hasPrefix(lv.Data, rv.Data, lvCursor, rvCursor, lvLen, rvLen)

	return rs
}
