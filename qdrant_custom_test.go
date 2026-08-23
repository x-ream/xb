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
// Custom interface test
// ============================================================================

func TestQdrantCustom_ImplementsCustomInterface(t *testing.T) {
	// ✅ Verify QdrantCustom implements Custom interface
	var _ Custom = (*QdrantCustom)(nil)
	t.Log("✅ QdrantCustom implements Custom interface")
}

// ============================================================================
// JsonOfSelect unified interface test
// ============================================================================

func TestJsonOfSelect_WithQdrantCustom(t *testing.T) {
	// Test automatically calling Qdrant logic through Custom
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of(&CodeVector{}).
		Custom(NewQdrantBuilder().Recommend(func(rb *RecommendBuilder) {
			rb.Positive(123, 456).Limit(20)
		}).Build()).
		VectorSearch("embedding", queryVector, 20).
		Eq("language", "golang").
		Build()

	// ⭐ Unified interface: JsonOfSelect()
	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	// Verify JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Verify basic fields
	if result["limit"] != float64(20) {
		t.Errorf("limit = %v, want 20", result["limit"])
	}

	if result["with_payload"] != true {
		t.Errorf("with_payload = %v, want true", result["with_payload"])
	}

	t.Logf("✅ JsonOfSelect with QdrantCustom works")
}

// ============================================================================
// Custom preset mode test
// ============================================================================

func TestQdrantCustom_PresetModes(t *testing.T) {
	tests := []struct {
		name         string
		custom       *QdrantCustom
		expectHnswEf int
		expectScore  float32
		expectVector bool
	}{
		{"Default", NewQdrantBuilder().Build(), 128, 0.0, false}, // Default: don't return vector (save bandwidth)
		{"CustomHnswEf", func() *QdrantCustom {
			return NewQdrantBuilder().
				HnswEf(512).
				ScoreThreshold(0.85).
				Build()
		}(), 512, 0.85, false}, // Inherit default value false
		{"CustomHighSpeed", func() *QdrantCustom {
			return NewQdrantBuilder().
				HnswEf(32).
				ScoreThreshold(0.5).
				WithVector(false).
				Build()
		}(), 32, 0.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.custom.DefaultHnswEf != tt.expectHnswEf {
				t.Errorf("DefaultHnswEf = %d, want %d", tt.custom.DefaultHnswEf, tt.expectHnswEf)
			}

			if tt.custom.DefaultScoreThreshold != tt.expectScore {
				t.Errorf("DefaultScoreThreshold = %f, want %f", tt.custom.DefaultScoreThreshold, tt.expectScore)
			}

			if tt.custom.DefaultWithVector != tt.expectVector {
				t.Errorf("DefaultWithVector = %v, want %v", tt.custom.DefaultWithVector, tt.expectVector)
			}

			t.Logf("✅ %s mode: HnswEf=%d, ScoreThreshold=%f, WithVector=%v",
				tt.name, tt.custom.DefaultHnswEf, tt.custom.DefaultScoreThreshold, tt.custom.DefaultWithVector)
		})
	}
}

// ============================================================================
// Case when Custom is nil
// ============================================================================

func TestJsonOfSelect_CustomIsNil(t *testing.T) {
	// Case when Custom is not set
	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of("code_vectors").
		VectorSearch("embedding", queryVector, 10).
		Build()

	// built.Custom should be nil
	if built.Custom != nil {
		t.Fatalf("Custom should be nil, got %T", built.Custom)
	}

	// JsonOfSelect should return error
	_, err := built.JsonOfSelect()
	if err == nil {
		t.Fatalf("JsonOfSelect should return error when Custom is nil")
	}

	t.Logf("✅ JsonOfSelect correctly returns error when Custom is nil: %v", err)
}

// ============================================================================
// Custom switch test (runtime switching)
// ============================================================================

func TestCustomSwitch_RuntimeSelection(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	// Simulate selecting different Custom based on configuration
	// ⭐ Users can manually configure as needed
	customs := map[string]*QdrantCustom{
		"default": NewQdrantBuilder().Build(),
		"high_precision": func() *QdrantCustom {
			return NewQdrantBuilder().
				HnswEf(512).
				ScoreThreshold(0.85).
				Build()
		}(),
		"high_speed": func() *QdrantCustom {
			return NewQdrantBuilder().
				HnswEf(32).
				ScoreThreshold(0.5).
				WithVector(false).
				Build()
		}(),
	}

	for name, custom := range customs {
		t.Run(name, func(t *testing.T) {
			built := Of("code_vectors").
				Custom(custom).
				VectorSearch("embedding", queryVector, 10).
				Build()

			jsonStr, err := built.JsonOfSelect()
			if err != nil {
				t.Fatalf("JsonOfSelect failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				t.Fatalf("JSON unmarshal failed: %v", err)
			}

			t.Logf("✅ %s mode works", name)
		})
	}
}

// ============================================================================
// Custom configuration test
// ============================================================================

func TestQdrantCustom_CustomConfiguration(t *testing.T) {
	// User custom configuration
	customConfig := &QdrantCustom{
		DefaultHnswEf:         256,
		DefaultScoreThreshold: 0.75,
		DefaultWithVector:     true,
	}

	queryVector := Vector{0.1, 0.2, 0.3}

	built := Of("code_vectors").
		Custom(customConfig).
		VectorSearch("embedding", queryVector, 10).
		Build()

	jsonStr, err := built.JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	t.Logf("✅ Custom configuration applied")
}

// ============================================================================
// JsonOfSelect advanced API regression test
// ============================================================================

func TestJsonOfSelect_WithRecommendConfig(t *testing.T) {
	queryVector := Vector{0.01, 0.02, 0.03}

	jsonStr, err := Of(&CodeVector{}).
		Custom(NewQdrantBuilder().Recommend(func(rb *RecommendBuilder) {
			rb.Positive(11, 22).Negative(33).Limit(5)
		}).Build()).
		VectorSearch("embedding", queryVector, 5).
		Build().
		JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		t.Fatalf("failed to parse recommend JSON: %v", err)
	}

	if limit := payload["limit"]; limit != float64(5) {
		t.Fatalf("recommend limit want 5, got %v", limit)
	}
	if _, ok := payload["positive"]; !ok {
		t.Fatalf("recommend payload missing positive ids")
	}
	if _, ok := payload["negative"]; !ok {
		t.Fatalf("recommend payload missing negative ids")
	}
}

func TestJsonOfSelect_WithDiscoverConfig(t *testing.T) {
	queryVector := Vector{0.1, 0.2, 0.3}

	jsonStr, err := Of(&CodeVector{}).
		Custom(NewQdrantBuilder().Discover(func(db *DiscoverBuilder) {
			db.Context(101, 102, 103).Limit(8)
		}).Build()).
		VectorSearch("embedding", queryVector, 8).
		Build().
		JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		t.Fatalf("failed to parse discover JSON: %v", err)
	}

	if limit := payload["limit"]; limit != float64(8) {
		t.Fatalf("discover limit want 8, got %v", limit)
	}
	if _, ok := payload["context"]; !ok {
		t.Fatalf("discover payload missing context ids")
	}
}

func TestJsonOfSelect_WithScrollConfig(t *testing.T) {
	scrollID := "scroll-xyz"

	jsonStr, err := Of(&CodeVector{}).
		Custom(NewQdrantBuilder().ScrollID(scrollID).Build()).
		Eq("language", "golang").
		Build().
		JsonOfSelect()
	if err != nil {
		t.Fatalf("JsonOfSelect failed: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		t.Fatalf("failed to parse scroll JSON: %v", err)
	}

	if got := payload["scroll_id"]; got != scrollID {
		t.Fatalf("scroll_id want %s, got %v", scrollID, got)
	}
}
