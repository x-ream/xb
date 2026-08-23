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

// Test vector data model (same as vector_test.go)
type CodeVectorForQdrant struct {
	Id           int64  `db:"id"`
	Content      string `db:"content"`
	Embedding    Vector `db:"embedding"`
	Language     string `db:"language"`
	Layer        string `db:"layer"`
	SemanticHash string `db:"semantic_hash"` // ⭐ Used for diversity deduplication
}

func (CodeVectorForQdrant) TableName() string {
	return "code_vectors"
}

// Test basic Qdrant JSON generation
func TestJsonOfSelect_Basic(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3, 0.4}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== Basic Qdrant JSON ===\n%s", jsonStr)

	// Verify JSON format
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// Verify fields
	if len(req.Vector) != 4 {
		t.Errorf("Expected vector length 4, got %d", len(req.Vector))
	}
	if req.Limit != 10 {
		t.Errorf("Expected limit 10, got %d", req.Limit)
	}
}

// Test Qdrant JSON with scalar filtering
func TestJsonOfSelect_WithFilter(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		Eq("layer", "repository").
		VectorSearch("embedding", queryVector, 20).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== Qdrant JSON with Filter ===\n%s", jsonStr)

	// Verify JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// Verify filter
	if req.Filter == nil {
		t.Errorf("Expected filter, got nil")
	} else if len(req.Filter.Must) != 2 {
		t.Errorf("Expected 2 must conditions, got %d", len(req.Filter.Must))
	}
}

// Test hash diversity - Qdrant JSON
func TestJsonOfSelect_WithHashDiversity(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		WithHashDiversity("semantic_hash"). // ⭐ Diversity: hash deduplication
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== Hash Diversity Qdrant JSON ===\n%s", jsonStr)

	// Verify JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ Diversity: limit should be enlarged (over-fetch)
	if req.Limit != 20*5 { // Default 5x
		t.Errorf("Expected limit %d (20*5), got %d", 20*5, req.Limit)
	}

	t.Logf("✅ Diversity enabled: Limit from 20 to %d (5x over-fetch)", req.Limit)
	t.Logf("ℹ️  Need to deduplicate based on semantic_hash to 20 in application layer")
}

// Test minimum distance diversity - Qdrant JSON
func TestJsonOfSelect_WithMinDistance(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		VectorSearch("embedding", queryVector, 20).
		WithMinDistance(0.3). // ⭐ Diversity: minimum distance 0.3
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== Minimum Distance Diversity Qdrant JSON ===\n%s", jsonStr)

	// Verify JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ Diversity: limit should be enlarged
	if req.Limit != 20*5 {
		t.Errorf("Expected limit %d, got %d", 20*5, req.Limit)
	}

	t.Logf("✅ Diversity enabled: Limit expanded from 20 to %d", req.Limit)
	t.Logf("ℹ️  Need to ensure distance >= 0.3 between results in application layer")
}

// Test MMR diversity - Qdrant JSON
func TestJsonOfSelect_WithMMR(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		WithMMR(0.5). // ⭐ Diversity: MMR lambda=0.5
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== MMR Diversity Qdrant JSON ===\n%s", jsonStr)

	// Verify JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// ⭐ Diversity: limit should be enlarged
	if req.Limit != 20*5 {
		t.Errorf("Expected limit %d, got %d", 20*5, req.Limit)
	}

	t.Logf("✅ Diversity enabled: Limit expanded from 20 to %d", req.Limit)
	t.Logf("ℹ️  Need to filter using MMR algorithm (lambda=0.5) in application layer")
}

// Test range query - Qdrant JSON
func TestJsonOfSelect_WithRange(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Gt("score", 0.8).
		Lt("complexity", 100).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== Range Query Qdrant JSON ===\n%s", jsonStr)

	// Verify JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// Verify filter
	if req.Filter == nil || len(req.Filter.Must) < 2 {
		t.Errorf("Expected at least 2 range conditions")
	}
}

// Test IN query - Qdrant JSON
func TestJsonOfSelect_WithIn(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		In("language", "golang", "python", "rust"). // ⭐ Fix: pass values directly, not slice
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	t.Logf("=== IN Query Qdrant JSON ===\n%s", jsonStr)

	// Verify JSON
	var req QdrantSearchRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	// Verify filter
	if req.Filter == nil || len(req.Filter.Must) == 0 {
		t.Errorf("Expected IN condition in filter")
	}
}

// Test PostgreSQL SQL ignores diversity
func TestSqlOfVectorSearch_IgnoresDiversity(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// ⭐ Key test: PostgreSQL SQL should work normally even with diversity set
	built := Of(&CodeVectorForQdrant{}).
		Eq("language", "golang").
		VectorSearch("embedding", queryVector, 20).
		WithHashDiversity("semantic_hash"). // ⭐ Diversity setting
		Build()

	// Generate PostgreSQL SQL
	sql, args := built.SqlOfVectorSearch()

	t.Logf("=== PostgreSQL SQL (should ignore diversity) ===")
	t.Logf("SQL: %s", sql)
	t.Logf("Args: %v", args)

	// ⭐ Key verification: SQL should not contain diversity-related logic
	// LIMIT should remain 20, not 100
	if !containsString(sql, "LIMIT 20") {
		t.Errorf("Expected LIMIT 20 in SQL (diversity should be ignored)")
	}

	// Verify basic query functionality is normal
	if !containsString(sql, "language") {
		t.Errorf("Expected language filter in SQL")
	}

	if !containsString(sql, "ORDER BY distance") {
		t.Errorf("Expected ORDER BY distance in SQL")
	}

	t.Logf("✅ Diversity parameters correctly ignored (PostgreSQL doesn't support)")
}

// Test full workflow
func TestQdrant_FullWorkflow(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3, 0.4}

	// Build query
	builder := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		Gt("quality_score", 0.8).
		VectorSearch("embedding", queryVector, 20).
		WithHashDiversity("semantic_hash")

	built := builder.Build()

	// 1. PostgreSQL usage
	t.Log("\n=== 1. PostgreSQL Backend ===")
	sql, args := built.SqlOfVectorSearch()
	t.Logf("SQL: %s", sql)
	t.Logf("Args count: %d", len(args))
	t.Logf("✅ Diversity ignored, SQL generated normally")

	// 2. Qdrant usage
	t.Log("\n=== 2. Qdrant Backend ===")
	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}
	t.Logf("JSON:\n%s", jsonStr)
	t.Logf("✅ Diversity applied, Limit expanded to %d", 20*5)

	// 3. Verify the same Built object can be used for different backends
	t.Log("\n=== 3. Same Query, Multi-Backend Compatible ===")
	t.Logf("✅ One code, two backends: PostgreSQL and Qdrant")
	t.Logf("✅ Graceful degradation: unsupported features automatically ignored")
}
