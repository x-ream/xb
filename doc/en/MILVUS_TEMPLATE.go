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

//go:build ignore
// +build ignore

// ============================================================================
// Milvus Extension Template (copy this file to to_milvus_json.go to start implementation)
// ============================================================================
//
// ⭐ Note: This file is only a template reference and will not be compiled (build ignore)
//
// This is a complete Milvus support template showing how to quickly implement
// Milvus vector database support based on the VectorDBRequest interface.
//
// Implementation steps:
//  1. Copy this file to to_milvus_json.go (remove build ignore tag)
//  2. Add Milvus operator constants in oper.go
//  3. Add Builder methods in cond_builder_milvus.go
//  4. Run tests to verify
//
// ============================================================================

package xb

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"

	. "github.com/fndome/xb"
)

// ============================================================================
// Step 1: Add Milvus-specific operators in oper.go
// ============================================================================

/*
Add in oper.go file:

const (
	MILVUS_NPROBE     = "MILVUS_NPROBE"      // Search parameter nprobe
	MILVUS_ROUND_DEC  = "MILVUS_ROUND_DEC"   // Round decimal places
	MILVUS_EF         = "MILVUS_EF"          // HNSW search parameter
	MILVUS_EXPR       = "MILVUS_EXPR"        // Filter expression
	MILVUS_XX         = "MILVUS_XX"          // Custom parameter
)
*/

// ============================================================================
// Step 2: Define Milvus-specific interface (inherits VectorDBRequest)
// ============================================================================

// MilvusRequest Milvus-specific request interface
// Inherits VectorDBRequest, automatically supports common parameters (ScoreThreshold, WithVector)
type MilvusRequest interface {
	VectorDBRequest // ⭐ Inherit common interface

	// Milvus-specific methods
	GetSearchParams() **MilvusSearchParams
	GetExpr() *string
}

// ============================================================================
// Step 3: Define Milvus request struct
// ============================================================================

// MilvusSearchRequest Milvus search request
type MilvusSearchRequest struct {
	CollectionName string      `json:"collection_name"`
	Vectors        [][]float32 `json:"vectors"`
	TopK           int         `json:"topk"`
	MetricType     string      `json:"metric_type"`

	// ⭐ Common fields (automatically supported)
	ScoreThreshold *float32 `json:"score_threshold,omitempty"`
	OutputFields   []string `json:"output_fields,omitempty"` // WithVector controls this field

	// ⭐ Milvus-specific fields
	SearchParams *MilvusSearchParams `json:"search_params,omitempty"`
	Expr         string              `json:"expr,omitempty"`
}

// MilvusSearchParams Milvus search parameters
type MilvusSearchParams struct {
	NProbe   int `json:"nprobe,omitempty"`
	RoundDec int `json:"round_decimal,omitempty"`
	Ef       int `json:"ef,omitempty"` // HNSW parameter
}

// ============================================================================
// Step 4: Implement interface methods
// ============================================================================

// ⭐ Implement VectorDBRequest (common interface)

func (r *MilvusSearchRequest) GetScoreThreshold() **float32 {
	return &r.ScoreThreshold
}

func (r *MilvusSearchRequest) GetWithVector() *bool {
	// Milvus controls whether to return vectors through OutputFields
	// Returning nil here means direct bool setting is not supported
	// In actual applications, need to handle in applyMilvusParams
	return nil
}

func (r *MilvusSearchRequest) GetFilter() interface{} {
	return &r.Expr // Milvus uses Expr string for filtering
}

// ⭐ Implement MilvusRequest (Milvus-specific interface)

func (r *MilvusSearchRequest) GetSearchParams() **MilvusSearchParams {
	return &r.SearchParams
}

func (r *MilvusSearchRequest) GetExpr() *string {
	return &r.Expr
}

// ============================================================================
// Step 5: Parameter application function (reuse common logic)
// ============================================================================

// applyMilvusParams applies Milvus-specific parameters
func applyMilvusParams(bbs []Bb, req MilvusRequest) {
	// ⭐ First layer: reuse common parameter application
	ApplyCommonVectorParams(bbs, req)

	// ⭐ Second layer: apply Milvus-specific parameters
	for _, bb := range bbs {
		switch bb.Op {
		case "MILVUS_NPROBE": // Need to define in oper.go
			ensureMilvusParams(req)
			(*req.GetSearchParams()).NProbe = bb.Value.(int)

		case "MILVUS_ROUND_DEC":
			ensureMilvusParams(req)
			(*req.GetSearchParams()).RoundDec = bb.Value.(int)

		case "MILVUS_EF":
			ensureMilvusParams(req)
			(*req.GetSearchParams()).Ef = bb.Value.(int)

		case "MILVUS_EXPR":
			expr := bb.Value.(string)
			*req.GetExpr() = expr
		}
	}
}

// ensureMilvusParams ensures SearchParams is initialized
func ensureMilvusParams(req MilvusRequest) {
	params := req.GetSearchParams()
	if *params == nil {
		*params = &MilvusSearchParams{}
	}
}

// ============================================================================
// Step 6: JSON conversion function (on Built, consistent with Qdrant)
// ============================================================================

// JsonOfMilvusSelect converts to Milvus search JSON
// ⭐ Naming consistent with SQL: JsonOfSelect (Milvus version)
// ⭐ Design consistent with Qdrant: called on Built, get information from VectorSearch parameters
//
// Returns:
//   - JSON string
//   - error
//
// Example:
//
//	built := C().
//	    VectorScoreThreshold(0.8).      // Common parameter
//	    MilvusNProbe(64).               // Milvus-specific
//	    MilvusExpr("age > 18").         // Filter expression
//	    MilvusX("consistency_level", "Strong"). // Custom parameter
//	    VectorSearch("users", "embedding", []float32{0.1, 0.2}, 10, L2Distance).
//	    Build()
//
//	json, err := built.JsonOfMilvusSelect()
func (built *Built) JsonOfMilvusSelect() (string, error) {
	// 1️⃣ Find VECTOR_SEARCH parameter from Built.Conds
	vectorBb := findVectorSearchBb(built.Conds)
	if vectorBb == nil {
		return "", fmt.Errorf("no VECTOR_SEARCH found, use VectorSearch() to specify search parameters")
	}

	params := vectorBb.Value.(VectorSearchParams)

	// 2️⃣ Create Milvus request object
	req := &MilvusSearchRequest{
		CollectionName: params.TableName,
		Vectors:        [][]float32{params.Vector},
		TopK:           params.Limit,
		MetricType:     milvusDistanceMetric(params.Distance),
	}

	// 3️⃣ Apply parameters (automatically handle common + specific parameters)
	applyMilvusParams(built.Conds, req)

	// 4️⃣ Serialize to JSON (reuse common logic)
	return milvusMergeAndSerialize(req, built.Conds)
}

// findVectorSearchBb finds VECTOR_SEARCH from Bb array
func findVectorSearchBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].Op == VECTOR_SEARCH {
			return &bbs[i]
		}
	}
	return nil
}

// milvusDistanceMetric converts distance metric
func milvusDistanceMetric(metric VectorDistance) string {
	switch metric {
	case CosineDistance:
		return "COSINE"
	case L2Distance:
		return "L2"
	case InnerProduct:
		return "IP"
	default:
		return "L2"
	}
}

// milvusMergeAndSerialize merges custom parameters and serializes
// ⭐ This function is completely consistent with Qdrant's mergeAndSerialize logic
func milvusMergeAndSerialize(req interface{}, bbs []Bb) (string, error) {
	// ⭐ Reuse common extraction function
	customParams := ExtractCustomParams(bbs, "MILVUS_XX")

	if len(customParams) == 0 {
		// No custom parameters, serialize directly
		bytes, err := json.Marshal(req, jsontext.WithIndent("  "))
		if err != nil {
			return "", fmt.Errorf("failed to marshal Milvus request: %w", err)
		}
		return string(bytes), nil
	}

	// Has custom parameters, first serialize to map, then add custom fields
	bytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Milvus request: %w", err)
	}

	var reqMap map[string]interface{}
	if err := json.Unmarshal(bytes, &reqMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add user custom parameters
	for k, v := range customParams {
		reqMap[k] = v
	}

	// Re-serialize
	finalBytes, err := json.Marshal(reqMap, jsontext.WithIndent("  "))
	if err != nil {
		return "", fmt.Errorf("failed to marshal final JSON: %w", err)
	}

	return string(finalBytes), nil
}

// ============================================================================
// Step 7: Add Builder methods in cond_builder_milvus.go
// ============================================================================

/*
Create cond_builder_milvus.go file:

package xb

// ⭐ Common parameters (already implemented in cond_builder.go)
// func (b *CondBuilder) VectorScoreThreshold(threshold float32)
// func (b *CondBuilder) VectorWithVector(withVector bool)

// MilvusNProbe sets Milvus nprobe search parameter
func (b *CondBuilder) MilvusNProbe(nprobe int) *CondBuilder {
	return b.append(Bb{op: MILVUS_NPROBE, value: nprobe})
}

// MilvusRoundDec sets Milvus decimal rounding
func (b *CondBuilder) MilvusRoundDec(dec int) *CondBuilder {
	return b.append(Bb{op: MILVUS_ROUND_DEC, value: dec})
}

// MilvusEf sets Milvus HNSW ef parameter
func (b *CondBuilder) MilvusEf(ef int) *CondBuilder {
	return b.append(Bb{op: MILVUS_EF, value: ef})
}

// MilvusExpr sets Milvus filter expression
func (b *CondBuilder) MilvusExpr(expr string) *CondBuilder {
	return b.append(Bb{op: MILVUS_EXPR, value: expr})
}

// MilvusX user-defined Milvus parameter (similar to Qdrant's QdrantBuilder)
//
// Example:
//   MilvusX("consistency_level", "Strong")
//   MilvusX("travel_timestamp", 12345)
func (b *CondBuilder) MilvusX(key string, value interface{}) *CondBuilder {
	return b.append(Bb{op: MILVUS_XX, key: key, value: value})
}
*/

// ============================================================================
// Step 8: Test example
// ============================================================================

/*
Create to_milvus_json_test.go file:

package xb

import (
	"encoding/json/v2"
	"testing"
)

func TestMilvusSearchRequest_Interface(t *testing.T) {
	req := &MilvusSearchRequest{}

	// ✅ Verify implements VectorDBRequest
	var _ VectorDBRequest = req

	// ✅ Verify implements MilvusRequest
	var _ MilvusRequest = req
}

func TestJsonOfMilvusSelect(t *testing.T) {
	// ⭐ Call style consistent with SQL naming
	built := C().
		VectorScoreThreshold(0.8).
		MilvusNProbe(64).
		MilvusExpr("age > 18").
		MilvusX("consistency_level", "Strong").
		VectorSearch("users", "embedding", []float32{0.1, 0.2, 0.3}, 10, L2Distance).
		Build()

	jsonStr, err := built.JsonOfMilvusSelect()
	if err != nil {
		t.Fatalf("JsonOfMilvusSelect failed: %v", err)
	}

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Verify basic fields
	if result["collection_name"] != "users" {
		t.Errorf("collection_name = %v, want 'users'", result["collection_name"])
	}

	if result["topk"] != 10 {
		t.Errorf("topk = %v, want 10", result["topk"])
	}

	if result["metric_type"] != "L2" {
		t.Errorf("metric_type = %v, want 'L2'", result["metric_type"])
	}

	// Verify common parameters
	if result["score_threshold"] != 0.8 {
		t.Errorf("score_threshold = %v, want 0.8", result["score_threshold"])
	}

	// Verify Milvus-specific parameters
	searchParams := result["search_params"].(map[string]interface{})
	if searchParams["nprobe"] != 64 {
		t.Errorf("nprobe = %v, want 64", searchParams["nprobe"])
	}

	if result["expr"] != "age > 18" {
		t.Errorf("expr = %v, want 'age > 18'", result["expr"])
	}

	// Verify custom parameters
	if result["consistency_level"] != "Strong" {
		t.Errorf("consistency_level = %v, want 'Strong'", result["consistency_level"])
	}
}
*/

// ============================================================================
// Summary
// ============================================================================

/*
Through this template, Milvus users only need:

✅ 5 steps (define operators → define interface → implement methods → apply parameters → serialize)
✅ Automatically reuse common logic (ApplyCommonVectorParams, extractCustomParams)
✅ Zero code duplication (common parameters, custom parameters, JSON serialization all reused)
✅ Type safety (compile-time checking)
✅ Elegant API (as fluent as Qdrant)

Estimated code size:
- to_milvus_json.go: ~200 lines
- cond_builder_milvus.go: ~50 lines
- to_milvus_json_test.go: ~100 lines
Total: ~350 lines (vs Qdrant's 800 lines, 56% code reduction)

Core reason: reused common interfaces and functions! 🎉
*/
