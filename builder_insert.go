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

type InsertBuilder struct {
	bbs []Bb
}

func (b *InsertBuilder) Set(k string, v interface{}) *InsertBuilder {
	// Handle nil values early
	if v == nil {
		return b
	}

	buffer, ok := v.([]byte)
	if ok {
		b.bbs = append(b.bbs, Bb{
			Key:   k,
			Value: buffer,
		})
		return b
	}

	defer func() *InsertBuilder {
		if s := recover(); s != nil {
			bytes, _ := json.Marshal(v)
			b.bbs = append(b.bbs, Bb{
				Key:   k,
				Value: string(bytes),
			})
		}
		return b
	}()

	switch v.(type) {
	case string:
	case uint64, uint, int64, int, int32, int16, int8, bool, byte, float64, float32:
		if v == 0 {
			return b
		}
	case *uint64, *uint, *int64, *int, *int32, *int16, *int8, *bool, *byte, *float64, *float32:
		isNil, n := NilOrNumber(v)
		if isNil {
			return b
		}
		v = n
	case uuid.UUID:
		// Handle uuid.UUID type - convert to string
		uuidVal := v.(uuid.UUID)
		if uuidVal == uuid.Nil() {
			return b
		}
		b.bbs = append(b.bbs, Bb{
			Key:   k,
			Value: uuidVal.String(),
		})
		return b
	case *time.Time:
		// Handle *time.Time pointer type
		timePtr := v.(*time.Time)
		if timePtr == nil {
			return b
		}
		ts := timePtr.Format("2006-01-02 15:04:05")
		v = ts
	case *string:
		// 场景：调用方传入 nil *string（如 models.User 的 Phone、Email 未赋值）。
		// 问题环节：v 的类型是 *string、值是 nil 时，v == nil 为 false（interface 里存的是 (type=*string, value=nil)），
		// 不会在第 32 行 return；switch 里又没有 *string 分支，会落到 case interface{}，执行 json.Marshal(v)。
		// json.Marshal(nil指针) 得到字节 "null"，string(bytes) 即字符串 "null"，被当作普通字符串插入 DB。
		// 修复：单独处理 *string，nil 则忽略不插入，非 nil 则取 *sptr 插入。
		sptr := v.(*string)
		if sptr == nil {
			return b
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
		// 不添加 case interface{}：实现 driver.Valuer 的结构体（如 BubbleKeywords）应原样传递，
		// 由 database/sql 调用 Value()；若落入 interface{} 会被 json.Marshal 错误处理
	}

	b.bbs = append(b.bbs, Bb{
		Key:   k,
		Value: v,
	})
	return b
}
