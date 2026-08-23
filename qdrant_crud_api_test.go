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

// ============================================================================
// Qdrant INSERT test (using real API)
// ============================================================================

// TestQdrantAPI_Insert test Qdrant Insert API
func TestQdrantAPI_Insert(t *testing.T) {
	// ⭐ Use Insert(func(ib *InsertBuilder)) - API consistent with SQL
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 123).
				Set("vector", []float32{0.1, 0.2, 0.3, 0.4}).
				Set("language", "golang").
				Set("content", "func main() {...}")
		}).
		Build()

	jsonStr, err := built.JsonOfInsert()
	if err != nil {
		t.Fatalf("JsonOfInsert failed: %v", err)
	}

	t.Logf("=== Qdrant Insert API ===\n%s", jsonStr)

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
	if int(req.Points[0].ID.(float64)) != 123 {
		t.Errorf("Expected ID 123, got %v", req.Points[0].ID)
	}

	// Verify payload
	if req.Points[0].Payload["language"] != "golang" {
		t.Errorf("Expected language=golang, got %v", req.Points[0].Payload["language"])
	}

	t.Logf("✅ Qdrant Insert API works")
}

// ============================================================================
// Qdrant UPDATE test (using real API)
// ============================================================================

// TestQdrantAPI_Update test Qdrant Update API
func TestQdrantAPI_Update(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("id", 123).
		Update(func(ub *UpdateBuilder) {
			ub.Set("language", "rust").
				Set("version", "1.75")
		}).
		Build()

	jsonStr, err := built.JsonOfUpdate()
	if err != nil {
		t.Fatalf("JsonOfUpdate failed: %v", err)
	}

	t.Logf("=== Qdrant Update API ===\n%s", jsonStr)

	var req struct {
		Points  []interface{}          `json:"points,omitempty"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point ID, got %d", len(req.Points))
	}

	if req.Payload["language"] != "rust" {
		t.Errorf("Expected language=rust, got %v", req.Payload["language"])
	}

	t.Logf("✅ Qdrant Update API works")
}

// TestQdrantAPI_UpdateByFilter test update by filter
func TestQdrantAPI_UpdateByFilter(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("language", "golang").
		Gt("quality_score", 0.8).
		Update(func(ub *UpdateBuilder) {
			ub.Set("verified", true)
		}).
		Build()

	jsonStr, err := built.JsonOfUpdate()
	if err != nil {
		t.Fatalf("JsonOfUpdate failed: %v", err)
	}

	t.Logf("=== Qdrant Update by Filter ===\n%s", jsonStr)

	var req struct {
		Filter  *QdrantFilter          `json:"filter,omitempty"`
		Payload map[string]interface{} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if req.Filter == nil {
		t.Errorf("Expected filter, got nil")
	}

	if len(req.Filter.Must) < 2 {
		t.Errorf("Expected at least 2 conditions, got %d", len(req.Filter.Must))
	}

	t.Logf("✅ Qdrant Update by Filter works: %d conditions", len(req.Filter.Must))
}

// ============================================================================
// Qdrant DELETE test (using real API)
// ============================================================================

// TestQdrantAPI_Delete test Qdrant Delete API
func TestQdrantAPI_Delete(t *testing.T) {
	built := Of(&CodeVectorForQdrant{}).
		Custom(NewQdrantBuilder().Build()).
		Eq("id", 456).
		Build()

	jsonStr, err := built.JsonOfDelete()
	if err != nil {
		t.Fatalf("JsonOfDelete failed: %v", err)
	}

	t.Logf("=== Qdrant Delete API ===\n%s", jsonStr)

	var req struct {
		Points []interface{} `json:"points,omitempty"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Errorf("Invalid JSON: %v", err)
	}

	if len(req.Points) != 1 {
		t.Errorf("Expected 1 point ID, got %d", len(req.Points))
	}

	t.Logf("✅ Qdrant Delete API works")
}

// ============================================================================
// Full CRUD workflow test (using real API)
// ============================================================================

// TestQdrantAPI_FullCRUD test full CRUD workflow
func TestQdrantAPI_FullCRUD(t *testing.T) {
	qdrant := NewQdrantBuilder().Build()

	// 1. INSERT
	t.Log("\n=== 1. INSERT ===")
	point := map[string]interface{}{
		"id":     999,
		"vector": []float32{0.9, 0.9, 0.9},
		"payload": map[string]interface{}{
			"language": "golang",
			"status":   "active",
		},
	}

	insertBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", point["id"]).
				Set("vector", point["vector"]).
				Set("language", point["payload"].(map[string]interface{})["language"]).
				Set("status", point["payload"].(map[string]interface{})["status"])
		}).
		Build()

	insertJSON, err := insertBuilt.JsonOfInsert()
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	t.Logf("Insert JSON:\n%s", insertJSON)

	// 2. SELECT (Search)
	t.Log("\n=== 2. SELECT (Search) ===")
	queryVector := Vector{0.9, 0.9, 0.9}
	searchBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("status", "active").
		VectorSearch("embedding", queryVector, 10).
		Build()

	searchJSON, err := searchBuilt.JsonOfSelect()
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	t.Logf("Search JSON (truncated):\n%s", searchJSON[:200]+"...")

	// 3. UPDATE
	t.Log("\n=== 3. UPDATE ===")
	updateBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 999).
		Update(func(ub *UpdateBuilder) {
			ub.Set("status", "archived")
		}).
		Build()

	updateJSON, err := updateBuilt.JsonOfUpdate()
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	t.Logf("Update JSON:\n%s", updateJSON)

	// 4. DELETE
	t.Log("\n=== 4. DELETE ===")
	deleteBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 999).
		Build()

	deleteJSON, err := deleteBuilt.JsonOfDelete()
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	t.Logf("Delete JSON:\n%s", deleteJSON)

	t.Log("\n✅ CRUD workflow test successful (using real API)!")
}

// ============================================================================
// Architecture validation: Custom interface unification
// ============================================================================

// TestQdrantAPI_CustomInterfaceUnification test Custom interface unification
func TestQdrantAPI_CustomInterfaceUnification(t *testing.T) {
	qdrant := NewQdrantBuilder().Build()

	t.Log("=== Verify: same Custom, handle all operations ===")

	// INSERT
	insertBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Insert(func(ib *InsertBuilder) {
			ib.Set("id", 1).
				Set("vector", []float32{0.1, 0.2}).
				Set("test", "value")
		}).
		Build()

	_, err := insertBuilt.JsonOfInsert()
	if err != nil {
		t.Errorf("INSERT failed: %v", err)
	}

	// UPDATE
	updateBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 1).
		Update(func(ub *UpdateBuilder) {
			ub.Set("test", "updated")
		}).
		Build()

	_, err = updateBuilt.JsonOfUpdate()
	if err != nil {
		t.Errorf("UPDATE failed: %v", err)
	}

	// DELETE
	deleteBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		Eq("id", 1).
		Build()

	_, err = deleteBuilt.JsonOfDelete()
	if err != nil {
		t.Errorf("DELETE failed: %v", err)
	}

	// SELECT
	selectBuilt := Of(&CodeVectorForQdrant{}).
		Custom(qdrant).
		VectorSearch("embedding", Vector{0.1, 0.2}, 10).
		Build()

	_, err = selectBuilt.JsonOfSelect()
	if err != nil {
		t.Errorf("SELECT failed: %v", err)
	}

	t.Log("✅ Same Custom instance handles all CRUD operations")
	t.Log("✅ Custom interface design verification successful")
}
