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
	"testing"
)

// TestQdrantCustom_OnlyCustom 测试只用 Custom（推荐用法）
func TestQdrantCustom_OnlyCustomNoQdrantX(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// 只使用 Custom，不使用 QdrantX
	built := Of(&CodeVectorForQdrant{}).
		Custom(
			NewQdrantBuilder().
				HnswEf(512).
				ScoreThreshold(0.75).
				WithVector(true).
				Build(),
		).
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== 只用 Custom ===\n%s", jsonStr)

	var req QdrantSearchRequest
	json.Unmarshal([]byte(jsonStr), &req)

	// ✅ 验证：使用 Custom 的默认值
	if req.Params.HnswEf != 512 {
		t.Errorf("Expected HnswEf 512, got %d", req.Params.HnswEf)
	}

	if req.ScoreThreshold == nil || *req.ScoreThreshold != 0.75 {
		t.Errorf("Expected ScoreThreshold 0.75, got %v", req.ScoreThreshold)
	}

	if !req.WithVector {
		t.Errorf("Expected WithVector true, got false")
	}

	t.Logf("✅ Custom 默认值正确应用：HnswEf=512, ScoreThreshold=0.75, WithVector=true")
}
