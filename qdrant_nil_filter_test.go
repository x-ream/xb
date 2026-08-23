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

// Test nil/0 automatic filtering
func TestQdrant_NilZeroFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// Build query containing nil/0 values
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang"). // ✅ Valid
		Eq("category", "").       // ⭐ Should be filtered (empty string)
		Gt("score", 0.8).         // ✅ Valid
		Gt("rank", 0).            // ⭐ Should be filtered (0)
		Lt("complexity", 100).    // ✅ Valid
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== nil/0 Filtering Test ===\n%s", jsonStr)

	// Parse JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ Key verification: filter should only have 3 conditions
	// language=golang, score>0.8, complexity<100
	// category="" and rank=0 should be filtered out
	if req.Filter == nil {
		t.Errorf("Expected filter, got nil")
	} else {
		mustCount := len(req.Filter.Must)

		// Should only have 3 conditions (excluding category="" and rank=0)
		if mustCount != 3 {
			t.Errorf("Expected 3 must conditions (filtered out empty string and 0), got %d", mustCount)
			t.Logf("Conditions:")
			for i, cond := range req.Filter.Must {
				t.Logf("  %d: key=%s, match=%v, range=%v", i, cond.Key, cond.Match, cond.Range)
			}
		} else {
			t.Logf("✅ nil/0 filtering successful: only 3 valid conditions")
		}

		// Verify does not contain category and rank
		for _, cond := range req.Filter.Must {
			if cond.Key == "category" {
				t.Errorf("category with empty value should be filtered")
			}
			if cond.Key == "rank" {
				t.Errorf("rank with value 0 should be filtered")
			}
		}
	}
}

// Test case when all are nil/0
func TestQdrant_AllNilZero(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// All conditions are nil/0
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("category", "").
		Gt("rank", 0).
		Lt("count", 0).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== All nil/0 Test ===\n%s", jsonStr)

	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// filter should be empty or have no must conditions
	if req.Filter != nil && len(req.Filter.Must) > 0 {
		t.Errorf("Expected no filter conditions, got %d", len(req.Filter.Must))
	} else {
		t.Logf("✅ All nil/0 conditions filtered")
	}
}

// Test PostgreSQL also has the same filtering
func TestPostgreSQL_NilZeroFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// Same query
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		Eq("category", ""). // ⭐ Should be filtered
		Gt("score", 0.8).
		Gt("rank", 0). // ⭐ Should be filtered
		VectorSearch("embedding", queryVector, 20).
		Build()

	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== PostgreSQL nil/0 Filtering Test ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// SQL should not contain category and rank
	if containsString(sql, "category") {
		t.Errorf("category with empty value should be filtered from SQL")
	}
	if containsString(sql, "rank") {
		t.Errorf("rank with value 0 should be filtered from SQL")
	}

	// Should only have 3 args: queryVector, "golang", and 0.8
	if len(args) != 3 {
		t.Errorf("Expected 3 args (vector, language, score), got %d: %v", len(args), args)
	} else {
		t.Logf("✅ PostgreSQL nil/0 filtering successful")
	}
}
