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
	"encoding/json"
	"fmt"
)

// ============================================================================
// Qdrant-Specific Interface (extends common interface)
// ============================================================================

// QdrantRequest Qdrant-specific request interface
// Extends common interface with Qdrant-specific HNSW parameters
type QdrantRequest interface {
	VectorDBRequest // ⭐ Extends common interface

	// GetParams gets Qdrant-specific search parameters (HNSW, Exact, etc.)
	GetParams() **QdrantSearchParams

	// GetQdrantFilter gets Qdrant-specific filter (type-safe)
	GetQdrantFilter() **QdrantFilter
}

// QdrantSearchRequest Qdrant search request structure
// Documentation: https://qdrant.tech/documentation/concepts/search/
type QdrantSearchRequest struct {
	Vector         []float32           `json:"vector"`
	Limit          int                 `json:"limit"`
	Filter         *QdrantFilter       `json:"filter,omitempty,omitzero"`
	WithPayload    interface{}         `json:"with_payload,omitempty"` // true, false, or []string
	WithVector     bool                `json:"with_vector,omitzero"`
	ScoreThreshold *float32            `json:"score_threshold,omitempty,omitzero"`
	Offset         int                 `json:"offset,omitzero"`
	Params         *QdrantSearchParams `json:"params,omitempty,omitzero"`
}

// Implements VectorDBRequest interface (common)
func (r *QdrantSearchRequest) GetScoreThreshold() **float32 {
	return &r.ScoreThreshold
}

func (r *QdrantSearchRequest) GetWithVector() *bool {
	return &r.WithVector
}

func (r *QdrantSearchRequest) GetFilter() interface{} {
	return &r.Filter
}

// Implements QdrantRequest interface (specific)
func (r *QdrantSearchRequest) GetParams() **QdrantSearchParams {
	return &r.Params
}

func (r *QdrantSearchRequest) GetQdrantFilter() **QdrantFilter {
	return &r.Filter
}

// QdrantFilter Qdrant filter
type QdrantFilter struct {
	Must    []QdrantCondition `json:"must,omitempty"`
	Should  []QdrantCondition `json:"should,omitempty"`
	MustNot []QdrantCondition `json:"must_not,omitempty"`
}

// QdrantCondition Qdrant condition
type QdrantCondition struct {
	Key   string                `json:"key,omitempty"`
	Match *QdrantMatchCondition `json:"match,omitempty"`
	Range *QdrantRangeCondition `json:"range,omitempty"`
}

// QdrantMatchCondition Qdrant exact match condition
type QdrantMatchCondition struct {
	Value interface{}   `json:"value,omitempty"`
	Any   []interface{} `json:"any,omitempty"`
}

// QdrantRangeCondition Qdrant range condition
type QdrantRangeCondition struct {
	Gt  *float64 `json:"gt,omitempty,omitzero"`
	Gte *float64 `json:"gte,omitempty,omitzero"`
	Lt  *float64 `json:"lt,omitempty,omitzero"`
	Lte *float64 `json:"lte,omitempty,omitzero"`
}

// QdrantSearchParams Qdrant search parameters
type QdrantSearchParams struct {
	HnswEf      int  `json:"hnsw_ef,omitzero"`
	Exact       bool `json:"exact,omitzero"`
	IndexedOnly bool `json:"indexed_only,omitzero"`
}

// QdrantRecommendRequest Qdrant recommend request structure (v0.10.0)
// Documentation: https://qdrant.tech/documentation/concepts/explore/#recommendation-api
type QdrantRecommendRequest struct {
	Positive       []int64             `json:"positive"`           // Positive sample ID list
	Negative       []int64             `json:"negative,omitempty"` // Negative sample ID list (optional)
	Limit          int                 `json:"limit"`
	Filter         *QdrantFilter       `json:"filter,omitempty,omitzero"`
	WithPayload    interface{}         `json:"with_payload,omitempty"` // true, false, or []string
	WithVector     bool                `json:"with_vector,omitzero"`
	ScoreThreshold *float32            `json:"score_threshold,omitempty,omitzero"`
	Offset         int                 `json:"offset,omitzero"`
	Params         *QdrantSearchParams `json:"params,omitempty,omitzero"`
	Strategy       string              `json:"strategy,omitempty"` // "average_vector" or "best_score"
}

// Implements VectorDBRequest interface (common)
func (r *QdrantRecommendRequest) GetScoreThreshold() **float32 {
	return &r.ScoreThreshold
}

func (r *QdrantRecommendRequest) GetWithVector() *bool {
	return &r.WithVector
}

func (r *QdrantRecommendRequest) GetFilter() interface{} {
	return &r.Filter
}

// Implements QdrantRequest interface (specific)
func (r *QdrantRecommendRequest) GetParams() **QdrantSearchParams {
	return &r.Params
}

func (r *QdrantRecommendRequest) GetQdrantFilter() **QdrantFilter {
	return &r.Filter
}

// QdrantScrollRequest Qdrant Scroll request structure (v0.10.0)
// Documentation: https://qdrant.tech/documentation/concepts/points/#scroll-points
type QdrantScrollRequest struct {
	ScrollID    string        `json:"scroll_id,omitempty"`
	Limit       int           `json:"limit,omitempty"`
	Filter      *QdrantFilter `json:"filter,omitempty,omitzero"`
	WithPayload interface{}   `json:"with_payload,omitempty"`
	WithVector  bool          `json:"with_vector,omitzero"`
}

// Implements VectorDBRequest interface (common)
func (r *QdrantScrollRequest) GetScoreThreshold() **float32 {
	return nil // Scroll does not support score threshold
}

func (r *QdrantScrollRequest) GetWithVector() *bool {
	return &r.WithVector
}

func (r *QdrantScrollRequest) GetFilter() interface{} {
	return &r.Filter
}

// Implements QdrantRequest interface (specific)
func (r *QdrantScrollRequest) GetParams() **QdrantSearchParams {
	return nil // Scroll does not support search parameters
}

func (r *QdrantScrollRequest) GetQdrantFilter() **QdrantFilter {
	return &r.Filter
}

// QdrantDiscoverRequest Qdrant Discover request structure (v0.10.0)
// Documentation: https://qdrant.tech/documentation/concepts/explore/#discovery-api
type QdrantDiscoverRequest struct {
	Context        []int64             `json:"context"` // Context sample ID list
	Limit          int                 `json:"limit"`
	Filter         *QdrantFilter       `json:"filter,omitempty,omitzero"`
	WithPayload    interface{}         `json:"with_payload,omitempty"` // true, false, or []string
	WithVector     bool                `json:"with_vector,omitzero"`
	ScoreThreshold *float32            `json:"score_threshold,omitempty,omitzero"`
	Offset         int                 `json:"offset,omitzero"`
	Params         *QdrantSearchParams `json:"params,omitempty,omitzero"`
}

// Implements VectorDBRequest interface (common)
func (r *QdrantDiscoverRequest) GetScoreThreshold() **float32 {
	return &r.ScoreThreshold
}

func (r *QdrantDiscoverRequest) GetWithVector() *bool {
	return &r.WithVector
}

func (r *QdrantDiscoverRequest) GetFilter() interface{} {
	return &r.Filter
}

// Implements QdrantRequest interface (specific)
func (r *QdrantDiscoverRequest) GetParams() **QdrantSearchParams {
	return &r.Params
}

func (r *QdrantDiscoverRequest) GetQdrantFilter() **QdrantFilter {
	return &r.Filter
}

func ensureQdrantAdvanced(built *Built) *Built {
	if built == nil {
		return nil
	}
	if qdrantCustom, ok := built.Custom.(*QdrantCustom); ok {
		return qdrantCustom.applyAdvancedConfig(built)
	}
	return built
}

// toQdrantJSON internal implementation
func (built *Built) toQdrantJSON() (string, error) {
	built = ensureQdrantAdvanced(built)
	req, err := built.ToQdrantRequest()
	if err != nil {
		return "", err
	}

	// ⭐ Check if there are user-defined parameters (QDRANT_XX)
	customParams := extractQdrantCustomParams(built.Conds)

	if len(customParams) == 0 {
		// No custom parameters, serialize directly
		bytes, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal Qdrant request: %w", err)
		}
		return string(bytes), nil
	}

	// Has custom parameters, first serialize to map, then add custom fields
	bytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant request: %w", err)
	}

	var reqMap map[string]interface{}
	if err := json.Unmarshal(bytes, &reqMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// ⭐ Add user-defined parameters
	for k, v := range customParams {
		reqMap[k] = v
	}

	// Re-serialize
	finalBytes, err := json.MarshalIndent(reqMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal final JSON: %w", err)
	}

	return string(finalBytes), nil
}

// extractQdrantCustomParams extracts Qdrant custom parameters (backward compatible)
// ⭐ Recommended to use general version extractCustomParams (defined in vector_db_request.go)
func extractQdrantCustomParams(bbs []Bb) map[string]interface{} {
	return ExtractCustomParams(bbs, QDRANT_XX)
}

// toQdrantRecommendJSON internal implementation
func (built *Built) toQdrantRecommendJSON() (string, error) {
	built = ensureQdrantAdvanced(built)
	// Find recommend parameters
	recommendBb := findRecommendBb(built.Conds)
	if recommendBb == nil {
		return "", fmt.Errorf("no recommend configuration found")
	}

	recommendData := recommendBb.Value.(map[string]interface{})

	// Build recommend request
	req := &QdrantRecommendRequest{
		Positive:    recommendData["positive"].([]int64),
		Negative:    []int64{},
		Limit:       recommendData["limit"].(int),
		WithPayload: true,
		WithVector:  false,
	}

	// Handle negative samples (optional)
	if negative, ok := recommendData["negative"].([]int64); ok && len(negative) > 0 {
		req.Negative = negative
	}

	// ⭐ Use unified parameter application function
	applyQdrantParams(built.Conds, req)

	// Apply filter
	filter, err := buildQdrantFilter(built.Conds)
	if err == nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// ⭐ Use unified serialization function
	return mergeAndSerialize(req, built.Conds)
}

// JsonOfSelect converts to Qdrant Scroll JSON (v0.10.0)
// Returns: JSON string, error
//
// Example output:
//
//	{
//	  "scroll_id": "xxxx-yyyy-zzzz",
//	  "limit": 100,
//	  "filter": {...}
//	}
//
// toQdrantScrollJSON internal implementation
func (built *Built) toQdrantScrollJSON() (string, error) {
	built = ensureQdrantAdvanced(built)
	// Find Scroll ID
	scrollBb := findScrollBb(built.Conds)
	if scrollBb == nil {
		return "", fmt.Errorf("no scroll_id found")
	}

	// Build Scroll request
	req := &QdrantScrollRequest{
		ScrollID:    scrollBb.Value.(string),
		Limit:       100, // Default value
		WithPayload: true,
		WithVector:  false,
	}

	// ⭐ Use unified parameter application function (Scroll supports WithVector)
	applyQdrantParams(built.Conds, req)

	// Apply filter
	filter, err := buildQdrantFilter(built.Conds)
	if err == nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// ⭐ Use unified serialization function
	return mergeAndSerialize(req, built.Conds)
}

// findRecommendBb finds recommend configuration
func findRecommendBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].Op == QDRANT_RECOMMEND {
			return &bbs[i]
		}
	}
	return nil
}

// findScrollBb finds Scroll ID
func findScrollBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].Op == QDRANT_SCROLL {
			return &bbs[i]
		}
	}
	return nil
}

// findDiscoverBb finds Discover configuration
func findDiscoverBb(bbs []Bb) *Bb {
	for i := range bbs {
		if bbs[i].Op == QDRANT_DISCOVER {
			return &bbs[i]
		}
	}
	return nil
}

// JsonOfSelect converts to Qdrant Discover JSON (v0.10.0)
// Returns: JSON string, error
//
// Example output:
//
//	{
//	  "context": [101, 102, 103],
//	  "limit": 20,
//	  "filter": {...}
//	}
//
// toQdrantDiscoverJSON internal implementation
func (built *Built) toQdrantDiscoverJSON() (string, error) {
	built = ensureQdrantAdvanced(built)
	// Find discover configuration
	discoverBb := findDiscoverBb(built.Conds)
	if discoverBb == nil {
		return "", fmt.Errorf("no discover configuration found")
	}

	discoverData := discoverBb.Value.(map[string]interface{})

	// Build discover request
	req := &QdrantDiscoverRequest{
		Context:     discoverData["context"].([]int64),
		Limit:       discoverData["limit"].(int),
		WithPayload: true,
		WithVector:  false,
	}

	// ⭐ Use unified parameter application function
	applyQdrantParams(built.Conds, req)

	// Apply filter
	filter, err := buildQdrantFilter(built.Conds)
	if err == nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// ⭐ Use unified serialization function
	return mergeAndSerialize(req, built.Conds)
}

// ToQdrantRequest builds Qdrant request object
// ⭐ Public method: for testing and advanced usage
func (built *Built) ToQdrantRequest() (*QdrantSearchRequest, error) {
	built = ensureQdrantAdvanced(built)
	// Find vector search parameters
	vectorBb := findVectorSearchBb(built.Conds)
	if vectorBb == nil {
		return nil, fmt.Errorf("no vector search found")
	}

	params := vectorBb.Value.(VectorSearchParams)

	// Build request
	req := &QdrantSearchRequest{
		Vector:      params.QueryVector,
		Limit:       params.TopK,
		WithPayload: true,
		WithVector:  false,
	}

	// ⭐ Diversity handling: if diversity is enabled, need to over-fetch
	if params.Diversity != nil && params.Diversity.Enabled {
		factor := params.Diversity.OverFetchFactor
		if factor <= 0 {
			factor = 5 // Default 5x
		}
		req.Limit = params.TopK * factor

		// Note: Qdrant doesn't natively support diversity, needs application-layer processing
		// Here we just fetch more results, application layer needs to filter later
	}

	// Build filter
	filter, err := buildQdrantFilter(built.Conds)
	if err != nil {
		return nil, err
	}
	if filter != nil && (len(filter.Must) > 0 || len(filter.Should) > 0 || len(filter.MustNot) > 0) {
		req.Filter = filter
	}

	// Set search parameters (use Custom's default values if available)
	defaultHnswEf := 128
	defaultScoreThreshold := float32(0.0)
	defaultWithVector := false

	// ⭐ Read default values from Custom (implementation approach 1)
	if built.Custom != nil {
		if qdrantCustom, ok := built.Custom.(*QdrantCustom); ok {
			defaultHnswEf = qdrantCustom.DefaultHnswEf
			defaultScoreThreshold = qdrantCustom.DefaultScoreThreshold
			defaultWithVector = qdrantCustom.DefaultWithVector
		}
	}

	req.Params = &QdrantSearchParams{
		HnswEf:      defaultHnswEf,
		Exact:       false,
		IndexedOnly: false,
	}

	// Apply Custom's default values
	if defaultScoreThreshold > 0 {
		req.ScoreThreshold = &defaultScoreThreshold
	}
	req.WithVector = defaultWithVector

	// ⭐ Apply Qdrant-specific configuration (extracted from Conds, will override defaults)
	applyQdrantSpecificConfig(built.Conds, req)

	// Apply pagination configuration (if any)
	if built.PageCondition != nil {
		req.Limit = int(built.PageCondition.Rows)
		if built.PageCondition.Page > 1 {
			req.Offset = int((built.PageCondition.Page - 1) * built.PageCondition.Rows)
		}

		// If diversity is enabled, need to override limit
		if params.Diversity != nil && params.Diversity.Enabled {
			factor := params.Diversity.OverFetchFactor
			if factor <= 0 {
				factor = 5
			}
			req.Limit = int(built.PageCondition.Rows) * factor
		}
	}

	return req, nil
}

// applyQdrantSpecificConfig extracts Qdrant-specific configuration from Bb
func applyQdrantSpecificConfig(bbs []Bb, req *QdrantSearchRequest) {
	for _, bb := range bbs {
		switch bb.Op {
		case QDRANT_HNSW_EF:
			if ef, ok := bb.Value.(int); ok {
				req.Params.HnswEf = ef
			}
		case QDRANT_EXACT:
			if exact, ok := bb.Value.(bool); ok {
				req.Params.Exact = exact
			}
		case QDRANT_SCORE_THRESHOLD:
			if threshold, ok := bb.Value.(float32); ok {
				req.ScoreThreshold = &threshold
			}
		case QDRANT_WITH_VECTOR:
			if withVec, ok := bb.Value.(bool); ok {
				req.WithVector = withVec
			}
		case QDRANT_XX:
			// ⭐ User-defined parameters
			// Note: These parameters will be added to the top level of JSON
			// Since QdrantSearchRequest is a fixed structure,
			// QDRANT_XX parameters are specially handled in JsonOfSelect()
			// This is just a marker, actual processing is in JsonOfSelect()
		}
	}
}

// buildQdrantFilter builds Qdrant filter
func buildQdrantFilter(bbs []Bb) (*QdrantFilter, error) {
	filter := &QdrantFilter{
		Must: []QdrantCondition{},
	}

	for _, bb := range bbs {
		// ⭐ Skip vector-specific operators (handled separately)
		if isVectorOp(bb.Op) {
			continue
		}

		// ⭐ Skip Qdrant-specific operators (handled separately)
		if isQdrantOp(bb.Op) {
			continue
		}

		cond, err := bbToQdrantCondition(bb)
		if err != nil {
			// ⭐ Key: unsupported operations don't error, just ignore
			continue
		}

		if cond != nil {
			filter.Must = append(filter.Must, *cond)
		}
	}

	return filter, nil
}

// isVectorOp checks if operator is vector operator
func isVectorOp(op string) bool {
	return op == VECTOR_SEARCH || op == VECTOR_DISTANCE_FILTER
}

// isQdrantOp checks if operator is Qdrant-specific operator
func isQdrantOp(op string) bool {
	return op == QDRANT_HNSW_EF ||
		op == QDRANT_EXACT ||
		op == QDRANT_SCORE_THRESHOLD ||
		op == QDRANT_WITH_VECTOR ||
		op == QDRANT_XX
}

// bbToQdrantCondition converts Bb to Qdrant condition
func bbToQdrantCondition(bb Bb) (*QdrantCondition, error) {
	switch bb.Op {
	case EQ:
		return &QdrantCondition{
			Key: bb.Key,
			Match: &QdrantMatchCondition{
				Value: bb.Value,
			},
		}, nil

	case NE:
		// Note: Qdrant's != needs to use must_not
		// Simplified handling here, caller needs to handle at upper level
		return nil, fmt.Errorf("NE not directly supported, use must_not")

	case IN:
		// IN converts to match.any
		// Note: IN's value is *[]string
		var anyValues []interface{}

		switch v := bb.Value.(type) {
		case *[]string:
			if v == nil {
				return nil, nil
			}
			for _, s := range *v {
				anyValues = append(anyValues, s)
			}
		case []interface{}:
			anyValues = v
		case []string:
			for _, s := range v {
				anyValues = append(anyValues, s)
			}
		default:
			return nil, fmt.Errorf("IN operator expects []string or []interface{}, got %T", bb.Value)
		}

		if len(anyValues) == 0 {
			return nil, nil
		}

		return &QdrantCondition{
			Key: bb.Key,
			Match: &QdrantMatchCondition{
				Any: anyValues,
			},
		}, nil

	case GT:
		val, err := toFloat64(bb.Value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.Key,
			Range: &QdrantRangeCondition{
				Gt: &val,
			},
		}, nil

	case GTE:
		val, err := toFloat64(bb.Value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.Key,
			Range: &QdrantRangeCondition{
				Gte: &val,
			},
		}, nil

	case LT:
		val, err := toFloat64(bb.Value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.Key,
			Range: &QdrantRangeCondition{
				Lt: &val,
			},
		}, nil

	case LTE:
		val, err := toFloat64(bb.Value)
		if err != nil {
			return nil, err
		}
		return &QdrantCondition{
			Key: bb.Key,
			Range: &QdrantRangeCondition{
				Lte: &val,
			},
		}, nil

	case LIKE:
		// Qdrant doesn't natively support LIKE, ignore
		return nil, fmt.Errorf("LIKE not supported in Qdrant")

	default:
		// Unsupported operations, return nil (ignore)
		return nil, nil
	}
}

// toFloat64 helper function: converts to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// QdrantDistanceMetric converts distance metric
func QdrantDistanceMetric(metric VectorDistance) string {
	switch metric {
	case CosineDistance:
		return "Cosine"
	case L2Distance:
		return "Euclid"
	case InnerProduct:
		return "Dot"
	default:
		return "Cosine"
	}
}

// ============================================================================
// Unified Parameter Application and Serialization Functions (Eliminate Duplicate Code)
// ============================================================================

// applyQdrantParams unified application of Qdrant-specific parameters
// Used to replace applyQdrantParamsToRecommend and applyQdrantParamsToDiscover
//
// Optimization: divided into two layers
//  1. Common parameters (ScoreThreshold, WithVector) → ApplyCommonVectorParams
//  2. Qdrant-specific parameters (HnswEf, Exact) → this function
func applyQdrantParams(bbs []Bb, req QdrantRequest) {
	// ⭐ First layer: apply common parameters (reuse cross-database logic)
	ApplyCommonVectorParams(bbs, req)

	// ⭐ Second layer: apply Qdrant-specific parameters
	for _, bb := range bbs {
		switch bb.Op {
		case QDRANT_HNSW_EF:
			if req.GetParams() != nil {
				ensureParams(req)
				(*req.GetParams()).HnswEf = bb.Value.(int)
			}

		case QDRANT_EXACT:
			if req.GetParams() != nil {
				ensureParams(req)
				(*req.GetParams()).Exact = bb.Value.(bool)
			}

			// ⭐ QDRANT_SCORE_THRESHOLD and QDRANT_WITH_VECTOR are already handled in ApplyCommonVectorParams
		}
	}
}

// ensureParams ensures Params field is initialized
func ensureParams(req QdrantRequest) {
	params := req.GetParams()
	if params != nil && *params == nil {
		*params = &QdrantSearchParams{}
	}
}

// mergeAndSerialize merges custom parameters and serializes to JSON
func mergeAndSerialize(req interface{}, bbs []Bb) (string, error) {
	// Extract custom parameters
	customParams := extractQdrantCustomParams(bbs)

	if len(customParams) == 0 {
		// No custom parameters, serialize directly
		bytes, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal Qdrant request: %w", err)
		}
		return string(bytes), nil
	}

	// Has custom parameters, first serialize to map, then add custom fields
	bytes, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Qdrant request: %w", err)
	}

	var reqMap map[string]interface{}
	if err := json.Unmarshal(bytes, &reqMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add user-defined parameters
	for k, v := range customParams {
		reqMap[k] = v
	}

	// Re-serialize
	finalBytes, err := json.MarshalIndent(reqMap, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal final JSON: %w", err)
	}

	return string(finalBytes), nil
}
