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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

// ============================================================================
// QdrantBuilder: Builder Pattern Configuration Builder
// ============================================================================

// QdrantBuilder Qdrant configuration builder
// Uses Builder pattern to construct QdrantCustom configuration
type QdrantBuilder struct {
	custom *QdrantCustom
}

// NewQdrantBuilder creates a Qdrant configuration builder
//
// Example:
//
//	// Basic configuration
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        ScoreThreshold(0.8).
//	        Build(),
//	).Build()
//
//	// Advanced API
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        Recommend(func(rb *xb.RecommendBuilder) {
//	            rb.Positive(123, 456).Limit(20)
//	        }).
//	        Build(),
//	).Build()
//
//	// Mixed configuration
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        ScoreThreshold(0.8).
//	        Recommend(func(rb *xb.RecommendBuilder) {
//	            rb.Positive(123, 456).Limit(20)
//	        }).
//	        Build(),
//	).Build()
func NewQdrantBuilder() *QdrantBuilder {
	return &QdrantBuilder{
		custom: newQdrantCustom(),
	}
}

// HnswEf sets the ef parameter for HNSW algorithm
// Larger ef means higher query precision but slower speed
// Recommended value: 64-256
func (qb *QdrantBuilder) HnswEf(ef int) *QdrantBuilder {
	if ef < 1 {
		panic(fmt.Sprintf("HnswEf must be >= 1, got: %d", ef))
	}
	qb.custom.DefaultHnswEf = ef
	return qb
}

// ScoreThreshold sets the minimum similarity threshold
// Only returns results with similarity >= threshold
func (qb *QdrantBuilder) ScoreThreshold(threshold float32) *QdrantBuilder {
	if threshold < 0 || threshold > 1 {
		panic(fmt.Sprintf("ScoreThreshold must be in [0, 1], got: %f", threshold))
	}
	qb.custom.DefaultScoreThreshold = threshold
	return qb
}

// WithVector sets whether to return vector data
// true: return vectors (uses bandwidth)
// false: don't return vectors (saves bandwidth)
func (qb *QdrantBuilder) WithVector(withVector bool) *QdrantBuilder {
	qb.custom.DefaultWithVector = withVector
	return qb
}

// Recommend enables Qdrant Recommend API
//
// Example:
//
//	xb.NewQdrantBuilder().
//	    HnswEf(512).
//	    Recommend(func(rb *xb.RecommendBuilder) {
//	        rb.Positive(123, 456).Negative(789).Limit(20)
//	    }).
//	    Build()
func (qb *QdrantBuilder) Recommend(fn func(rb *RecommendBuilder)) *QdrantBuilder {
	if fn == nil {
		qb.custom.recommendConfig = nil
		return qb
	}

	builder := &RecommendBuilder{}
	fn(builder)

	if len(builder.positive) == 0 {
		panic("Recommend() requires at least one Positive() id")
	}
	if builder.limit <= 0 {
		panic("Recommend() requires Limit() > 0")
	}

	qb.custom.recommendConfig = &qdrantRecommendConfig{
		positive: append([]int64(nil), builder.positive...),
		negative: append([]int64(nil), builder.negative...),
		limit:    builder.limit,
	}
	return qb
}

// Discover enables Qdrant Discover API
//
// Example:
//
//	xb.NewQdrantBuilder().
//	    HnswEf(512).
//	    Discover(func(db *xb.DiscoverBuilder) {
//	        db.Context(101, 102, 103).Limit(20)
//	    }).
//	    Build()
func (qb *QdrantBuilder) Discover(fn func(db *DiscoverBuilder)) *QdrantBuilder {
	if fn == nil {
		qb.custom.discoverConfig = nil
		return qb
	}

	builder := &DiscoverBuilder{}
	fn(builder)

	if len(builder.context) == 0 {
		panic("Discover() requires Context() with at least one id")
	}
	if builder.limit <= 0 {
		panic("Discover() requires Limit() > 0")
	}

	qb.custom.discoverConfig = &qdrantDiscoverConfig{
		context: append([]int64(nil), builder.context...),
		limit:   builder.limit,
	}
	return qb
}

// ScrollID enables Qdrant Scroll API
//
// Example:
//
//	xb.NewQdrantBuilder().
//	    HnswEf(512).
//	    ScrollID("scroll-abc123").
//	    Build()
func (qb *QdrantBuilder) ScrollID(scrollID string) *QdrantBuilder {
	if scrollID == "" {
		panic("ScrollID() requires a non-empty id")
	}
	qb.custom.scrollID = scrollID
	return qb
}

// Build constructs and returns QdrantCustom configuration
func (qb *QdrantBuilder) Build() *QdrantCustom {
	return qb.custom
}

// ============================================================================
// QdrantCustom: Qdrant Database-Specific Configuration
// ============================================================================

// QdrantCustom Qdrant-specific configuration implementation
//
// Implements Custom interface, provides Qdrant default configuration and preset modes
type QdrantCustom struct {
	// Default parameters (used if user doesn't explicitly specify)
	DefaultHnswEf         int     // Default HNSW EF parameter
	DefaultScoreThreshold float32 // Default similarity threshold
	DefaultWithVector     bool    // Default whether to return vectors

	// Advanced API configuration (Recommend / Discover / Scroll)
	recommendConfig *qdrantRecommendConfig
	discoverConfig  *qdrantDiscoverConfig
	scrollID        string
}

// newQdrantCustom internal function: creates Qdrant Custom (default configuration)
func newQdrantCustom() *QdrantCustom {
	return &QdrantCustom{
		DefaultHnswEf:         128,
		DefaultScoreThreshold: 0.0,
		DefaultWithVector:     false, // Backward compatible: default to not returning vectors
	}
}

// Generate implements Custom interface
// ⭐ Returns different JSON based on operation type
func (c *QdrantCustom) Generate(built *Built) (interface{}, error) {
	built = c.applyAdvancedConfig(built)

	// ⭐ INSERT: generate Qdrant upsert JSON
	if built.Inserts != nil && len(*built.Inserts) > 0 {
		return c.generateInsertJSON(built)
	}

	// ⭐ UPDATE: generate Qdrant update payload JSON
	if built.Updates != nil && len(*built.Updates) > 0 {
		return c.generateUpdateJSON(built)
	}

	// ⭐ DELETE: generate Qdrant delete JSON
	if built.Delete {
		return c.generateDeleteJSON(built)
	}

	// ⭐ SELECT: generate Qdrant search JSON
	switch {
	case hasBbWithOp(built.Conds, QDRANT_RECOMMEND):
		return built.toQdrantRecommendJSON()
	case hasBbWithOp(built.Conds, QDRANT_DISCOVER):
		return built.toQdrantDiscoverJSON()
	case hasBbWithOp(built.Conds, QDRANT_SCROLL):
		return built.toQdrantScrollJSON()
	default:
		json, err := built.toQdrantJSON()
		return json, err
	}
}

// ============================================================================
// Usage Instructions
// ============================================================================
//
// Configuration methods:
//
// Using QdrantBuilder (unified Builder pattern)
//
//	// Basic configuration
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        ScoreThreshold(0.85).
//	        Build(),
//	).Build()
//
//	// Advanced API
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        Recommend(func(rb *xb.RecommendBuilder) {
//	            rb.Positive(123, 456).Limit(20)
//	        }).
//	        Build(),
//	).Build()
//
//	// Mixed configuration (basic + advanced)
//	xb.Of(...).Custom(
//	    xb.NewQdrantBuilder().
//	        HnswEf(512).
//	        ScoreThreshold(0.85).
//	        Recommend(func(rb *xb.RecommendBuilder) {
//	            rb.Positive(123, 456).Limit(20)
//	        }).
//	        Build(),
//	).Build()
//
//	// Configuration reuse (same configuration can be used for multiple queries)
//	highPrecision := xb.NewQdrantBuilder().HnswEf(512).Build()
//	xb.Of(...).Custom(highPrecision).Build()
//	xb.Of(...).Custom(highPrecision).Build()  // Reuse configuration
//

// ============================================================================
// Insert/Update/Delete JSON Generation
// ============================================================================

// QdrantPoint Qdrant point structure
type QdrantPoint struct {
	ID      interface{}            `json:"id"`
	Vector  interface{}            `json:"vector"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// generateInsertJSON generates Qdrant upsert JSON
// PUT /collections/{collection_name}/points
func (c *QdrantCustom) generateInsertJSON(built *Built) (string, error) {
	inserts := *built.Inserts
	if len(inserts) == 0 {
		return "", fmt.Errorf("no insert data")
	}

	// Qdrant upsert request structure
	type QdrantUpsertRequest struct {
		Points []QdrantPoint `json:"points"`
	}

	points := []QdrantPoint{}

	// ⭐ Using Insert(func(ib)) format
	// Multiple bbs (field-value pairs) form one point
	point, err := c.extractPointFromBbs(inserts)
	if err != nil {
		return "", err
	}
	points = append(points, point)

	req := QdrantUpsertRequest{Points: points}
	bytes, err := json.Marshal(req, jsontext.WithIndent("  "))
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant upsert request: %w", err)
	}

	return string(bytes), nil
}

// generateUpdateJSON generates Qdrant update payload JSON
// POST /collections/{collection_name}/points/payload
func (c *QdrantCustom) generateUpdateJSON(built *Built) (string, error) {
	updates := *built.Updates
	if len(updates) == 0 {
		return "", fmt.Errorf("no update data")
	}

	// Qdrant update payload request structure
	type QdrantUpdateRequest struct {
		Points  []interface{}          `json:"points,omitempty"` // Point ID list
		Filter  *QdrantFilter          `json:"filter,omitempty"` // Or use filter
		Payload map[string]interface{} `json:"payload"`          // Payload to update
	}

	// Extract payload
	payload := make(map[string]interface{})
	for _, bb := range updates {
		payload[bb.Key] = bb.Value
	}

	req := QdrantUpdateRequest{
		Payload: payload,
	}

	// Extract point IDs from conditions or build filter
	ids, filter := c.extractIdsOrFilter(built.Conds)
	if len(ids) > 0 {
		req.Points = ids
	} else if filter != nil {
		req.Filter = filter
	}

	bytes, err := json.Marshal(req, jsontext.WithIndent("  "))
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant update request: %w", err)
	}

	return string(bytes), nil
}

// generateDeleteJSON generates Qdrant delete JSON
// POST /collections/{collection_name}/points/delete
func (c *QdrantCustom) generateDeleteJSON(built *Built) (string, error) {
	// Qdrant delete request structure
	type QdrantDeleteRequest struct {
		Points []interface{} `json:"points,omitempty"` // Point ID list
		Filter *QdrantFilter `json:"filter,omitempty"` // Or use filter
	}

	req := QdrantDeleteRequest{}

	// Extract point IDs from conditions or build filter
	ids, filter := c.extractIdsOrFilter(built.Conds)
	if len(ids) > 0 {
		req.Points = ids
	} else if filter != nil {
		req.Filter = filter
	} else {
		return "", fmt.Errorf("no delete conditions (points or filter)")
	}

	bytes, err := json.Marshal(req, jsontext.WithIndent("  "))
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant delete request: %w", err)
	}

	return string(bytes), nil
}

// ============================================================================
// Helper Methods
// ============================================================================

// extractPointFromBbs extracts Qdrant Point from InsertBuilder's bbs
// Used for Insert(func(ib *InsertBuilder)) format
func (c *QdrantCustom) extractPointFromBbs(bbs []Bb) (QdrantPoint, error) {
	point := QdrantPoint{
		Payload: make(map[string]interface{}),
	}

	for _, bb := range bbs {
		switch bb.Key {
		case "id":
			point.ID = bb.Value
		case "vector":
			point.Vector = bb.Value
		default:
			// Other fields go into payload
			point.Payload[bb.Key] = bb.Value
		}
	}

	// Validate required fields
	if point.ID == nil {
		return QdrantPoint{}, fmt.Errorf("point must have 'id' field")
	}
	if point.Vector == nil {
		return QdrantPoint{}, fmt.Errorf("point must have 'vector' field")
	}

	return point, nil
}

// extractIdsOrFilter extracts point IDs from conditions or builds filter
func (c *QdrantCustom) extractIdsOrFilter(conds []Bb) ([]interface{}, *QdrantFilter) {
	ids := []interface{}{}

	// Find id IN (...) condition
	for _, bb := range conds {
		if bb.Key == "id" {
			if bb.Op == IN {
				// IN condition: extract ID list
				if arr, ok := bb.Value.(*[]string); ok {
					for _, id := range *arr {
						ids = append(ids, id)
					}
					return ids, nil
				}
			} else if bb.Op == EQ {
				// Single ID
				ids = append(ids, bb.Value)
				return ids, nil
			}
		}
	}

	// If no id condition, build filter
	if len(conds) > 0 {
		filter := &QdrantFilter{}
		filter.Must = []QdrantCondition{}

		for _, bb := range conds {
			cond := QdrantCondition{
				Key: bb.Key,
			}

			switch bb.Op {
			case EQ:
				cond.Match = &QdrantMatchCondition{Value: bb.Value}
			case GT, GTE, LT, LTE:
				cond.Range = &QdrantRangeCondition{}
				switch bb.Op {
				case GT:
					cond.Range.Gt = toFloat64Ptr(bb.Value)
				case GTE:
					cond.Range.Gte = toFloat64Ptr(bb.Value)
				case LT:
					cond.Range.Lt = toFloat64Ptr(bb.Value)
				case LTE:
					cond.Range.Lte = toFloat64Ptr(bb.Value)
				}
			}

			filter.Must = append(filter.Must, cond)
		}

		return nil, filter
	}

	return nil, nil
}

func toFloat64Ptr(v interface{}) *float64 {
	switch val := v.(type) {
	case float64:
		return &val
	case float32:
		f := float64(val)
		return &f
	case int:
		f := float64(val)
		return &f
	case int64:
		f := float64(val)
		return &f
	}
	return nil
}

// qdrantRecommendConfig Recommend API configuration
type qdrantRecommendConfig struct {
	positive []int64
	negative []int64
	limit    int
}

// qdrantDiscoverConfig Discover API configuration
type qdrantDiscoverConfig struct {
	context []int64
	limit   int
}

// RecommendBuilder Recommend API builder
type RecommendBuilder struct {
	positive []int64
	negative []int64
	limit    int
}

// Positive sets positive sample IDs
func (rb *RecommendBuilder) Positive(ids ...int64) *RecommendBuilder {
	if len(ids) == 0 {
		return rb
	}
	rb.positive = append(rb.positive, ids...)
	return rb
}

// Negative sets negative sample IDs
func (rb *RecommendBuilder) Negative(ids ...int64) *RecommendBuilder {
	if len(ids) == 0 {
		return rb
	}
	rb.negative = append(rb.negative, ids...)
	return rb
}

// Limit sets the number of results to return
func (rb *RecommendBuilder) Limit(limit int) *RecommendBuilder {
	rb.limit = limit
	return rb
}

// DiscoverBuilder Discover API builder
type DiscoverBuilder struct {
	context []int64
	limit   int
}

// Context sets context IDs
func (db *DiscoverBuilder) Context(ids ...int64) *DiscoverBuilder {
	if len(ids) == 0 {
		return db
	}
	db.context = append(db.context, ids...)
	return db
}

// Limit sets the number of results to return
func (db *DiscoverBuilder) Limit(limit int) *DiscoverBuilder {
	db.limit = limit
	return db
}

// ensureAdvancedConds injects advanced configuration into condition list
func (c *QdrantCustom) ensureAdvancedConds(conds []Bb) []Bb {
	if c == nil {
		return conds
	}

	if c.recommendConfig != nil && !hasBbWithOp(conds, QDRANT_RECOMMEND) {
		value := map[string]interface{}{
			"positive": append([]int64(nil), c.recommendConfig.positive...),
			"limit":    c.recommendConfig.limit,
		}
		if len(c.recommendConfig.negative) > 0 {
			value["negative"] = append([]int64(nil), c.recommendConfig.negative...)
		}
		conds = append(conds, Bb{
			Op:    QDRANT_RECOMMEND,
			Value: value,
		})
	}

	if c.discoverConfig != nil && !hasBbWithOp(conds, QDRANT_DISCOVER) {
		value := map[string]interface{}{
			"context": append([]int64(nil), c.discoverConfig.context...),
			"limit":   c.discoverConfig.limit,
		}
		conds = append(conds, Bb{
			Op:    QDRANT_DISCOVER,
			Value: value,
		})
	}

	if c.scrollID != "" && !hasBbWithOp(conds, QDRANT_SCROLL) {
		conds = append(conds, Bb{
			Op:    QDRANT_SCROLL,
			Value: c.scrollID,
		})
	}

	return conds
}

func hasBbWithOp(bbs []Bb, op string) bool {
	for _, bb := range bbs {
		if bb.Op == op {
			return true
		}
	}
	return false
}

func (c *QdrantCustom) applyAdvancedConfig(built *Built) *Built {
	if c == nil || built == nil {
		return built
	}

	origLen := len(built.Conds)
	condsCopy := cloneBbs(built.Conds)
	newConds := c.ensureAdvancedConds(condsCopy)
	if len(newConds) == origLen {
		return built
	}

	cloned := *built
	cloned.Conds = newConds
	return &cloned
}

func cloneBbs(bbs []Bb) []Bb {
	if len(bbs) == 0 {
		return nil
	}
	cloned := make([]Bb, len(bbs))
	copy(cloned, bbs)
	return cloned
}
