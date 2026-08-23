// Copyright 2025 me.fndo.xb
//
// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package xb

import (
	"encoding/json/v2"
	"time"
	"uuid"
)

type UpdateBuilder struct {
	bbs []Bb
}

func (ub *UpdateBuilder) Set(k string, v interface{}) *UpdateBuilder {
	// Handle nil values early
	if v == nil {
		return ub
	}

	buffer, ok := v.([]byte)
	if ok {
		ub.bbs = append(ub.bbs, Bb{
			Key:   k,
			Value: buffer,
		})
		return ub
	}

	defer func() *UpdateBuilder {
		if s := recover(); s != nil {
			bytes, _ := json.Marshal(v)
			ub.bbs = append(ub.bbs, Bb{
				Key:   k,
				Value: string(bytes),
			})
		}
		return ub
	}()

	switch v.(type) {
	case string:
		// "" 忽略，不参与更新；显式置空请用 X("col = ''")
		if v == "" {
			return ub
		}
	case uint64, uint, int64, int, int32, int16, int8, bool, byte, float64, float32:
		if v == 0 {
			return ub
		}
	case *uint64, *uint, *int64, *int, *int32, *int16, *int8, *bool, *byte, *float64, *float32:
		isNil, n := NilOrNumber(v)
		if isNil {
			return ub
		}
		v = n
	case uuid.UUID:
		// Handle uuid.UUID type - convert to string
		uuidVal := v.(uuid.UUID)
		if uuidVal == uuid.Nil() {
			return ub
		}
		ub.bbs = append(ub.bbs, Bb{
			Key:   k,
			Value: uuidVal.String(),
		})
		return ub
	case *time.Time:
		// Handle *time.Time pointer type
		timePtr := v.(*time.Time)
		if timePtr == nil {
			return ub
		}
		ts := timePtr.Format("2006-01-02 15:04:05")
		v = ts
	case *string:
		// nil 忽略；*sptr 为 "" 也忽略（与 case string 一致）；非 nil 且非 "" 则取 *sptr 参与更新
		sptr := v.(*string)
		if sptr == nil || *sptr == "" {
			return ub
		}
		v = *sptr
	case time.Time:
		// Handle time.Time value type
		ts := v.(time.Time).Format("2006-01-02 15:04:05")
		v = ts
	case Vector:
		// Vector type: no processing, keep as is
		// Let database/sql call driver.Valuer interface
		// Vector.Value() will return the correct database format
	case []float32, []float64:
		// ⭐ Vector array: keep as is (for Qdrant/Milvus)
		// No JSON serialization
		// 不添加 case interface{}：实现 driver.Valuer 的结构体应原样传递，由 database/sql 调用 Value()
	}

	ub.bbs = append(ub.bbs, Bb{
		Key:   k,
		Value: v,
	})
	return ub
}

func (ub *UpdateBuilder) X(s string, vs ...interface{}) *UpdateBuilder {
	bb := Bb{
		Op:  "SET",
		Key: s,
	}
	// Only set Value if there are actual parameters
	if len(vs) > 0 {
		bb.Value = vs
	}
	ub.bbs = append(ub.bbs, bb)
	return ub
}

func (ub *UpdateBuilder) Any(f func(*UpdateBuilder)) *UpdateBuilder {
	f(ub)
	return ub
}

func (ub *UpdateBuilder) Bool(preCond BoolFunc, f func(cb *UpdateBuilder)) *UpdateBuilder {
	if preCond == nil {
		panic("UpdateBuilder.Bool para of BoolFunc can not nil")
	}
	if !preCond() {
		return ub
	}
	if f == nil {
		panic("UpdateBuilder.Bool para of func(k string, vs... interface{}) can not nil")
	}
	f(ub)
	return ub
}
