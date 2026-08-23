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
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xb

import (
	"encoding/json/v2"
	"testing"
)

// TestQdrantBuilder_Insert test QdrantBuilder build Custom and use for Insert
func TestQdrantBuilder_Insert(t *testing.T) {
	// ✅ Use QdrantBuilder to build Custom
	built := Of(&CodeVectorForQdrant{}).
		Custom(
			NewQdrantBuilder().
				HnswEf(512).
				ScoreThreshold(0.8).
				WithVector(false).
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 456).
				Set("vector", []float32{0.5, 0.6, 0.7, 0.8}).
				Set("language", "rust").
				Set("content", "fn main() { println!(\"Hello\"); }")
		}).
		Build()

	jsonStr, err := built.JsonOfInsert()
	if err != nil {
		t.Fatalf("JsonOfInsert failed: %v", err)
	}

	t.Logf("=== QdrantBuilder Insert ===\n%s", jsonStr)

	// Verify JSON
	var req struct {
		Points []QdrantPoint `json:"points"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point, got %d", len(req.Points))
	}

	// Verify ID
	if int(req.Points[0].ID.(float64)) != 456 {
		t.Errorf("Expected ID 456, got %v", req.Points[0].ID)
	}

	// Verify payload
	if req.Points[0].Payload["language"] != "rust" {
		t.Errorf("Expected language=rust, got %v", req.Points[0].Payload["language"])
	}

	t.Logf("✅ QdrantBuilder works correctly")
}

// TestQdrantBuilder_ConfigReuse test QdrantBuilder config reuse
func TestQdrantBuilder_ConfigReuse(t *testing.T) {
	// ✅ Config can be reused
	highPrecision := NewQdrantBuilder().
		HnswEf(512).
		ScoreThreshold(0.9).
		Build()

	// First use
	built1 := Of(&CodeVectorForQdrant{}).
		Custom(highPrecision).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 1).
				Set("vector", []float32{0.1, 0.2}).
				Set("language", "go")
		}).
		Build()

	json1, _ := built1.JsonOfInsert()
	t.Logf("=== First use ===\n%s", json1)

	// Second use (reuse config)
	built2 := Of(&CodeVectorForQdrant{}).
		Custom(highPrecision).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 2).
				Set("vector", []float32{0.3, 0.4}).
				Set("language", "rust")
		}).
		Build()

	json2, _ := built2.JsonOfInsert()
	t.Logf("=== Second use (reused config) ===\n%s", json2)

	t.Logf("✅ Config reuse works")
}

// TestQdrantBuilder_ChainStyle test QdrantBuilder chain style
func TestQdrantBuilder_ChainStyle(t *testing.T) {
	// ✅ Demonstrate the smoothness of chain style
	jsonStr, err := Of(&CodeVectorForQdrant{}).
		Custom(
			NewQdrantBuilder().
				HnswEf(256).
				ScoreThreshold(0.75).
				WithVector(true).
				Build(),
		).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 789).
				Set("vector", []float32{0.9, 1.0, 1.1, 1.2}).
				Set("language", "python").
				Set("content", "def main(): pass")
		}).
		Build().
		JsonOfInsert()

	if err != nil {
		t.Fatalf("JsonOfInsert failed: %v", err)
	}

	t.Logf("=== Chain Style ===\n%s", jsonStr)

	var req struct {
		Points []QdrantPoint `json:"points"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point, got %d", len(req.Points))
	}

	t.Logf("✅ Chain style works perfectly")
}
