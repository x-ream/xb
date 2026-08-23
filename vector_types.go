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
	"database/sql/driver"
	"encoding/json/v2"
	"fmt"
	"math"
)

// Vector vector type (compatible with PostgreSQL pgvector)
type Vector []float32

// Value implements driver.Valuer interface
func (v Vector) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}

	// PostgreSQL pgvector format: '[1,2,3]'
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// Convert [1,2,3] to '[1,2,3]' string format
	return string(bytes), nil
}

// Scan implements sql.Scanner interface
func (v *Vector) Scan(value interface{}) error {
	if value == nil {
		*v = nil
		return nil
	}

	switch value := value.(type) {
	case []byte:
		return json.Unmarshal(value, v)
	case string:
		return json.Unmarshal([]byte(value), v)
	default:
		return fmt.Errorf("unsupported vector type: %T", value)
	}
}

// VectorDistance vector distance metric type
type VectorDistance string

const (
	// CosineDistance cosine distance (most commonly used)
	// PostgreSQL: <->
	CosineDistance VectorDistance = "<->"

	// L2Distance Euclidean distance (L2 norm)
	// PostgreSQL: <#>
	L2Distance VectorDistance = "<#>"

	// InnerProduct inner product distance (dot product)
	// PostgreSQL: <=>
	InnerProduct VectorDistance = "<=>"
)

// Distance calculates the distance between two vectors
func (v Vector) Distance(other Vector, metric VectorDistance) float32 {
	if len(v) != len(other) {
		panic("vectors must have same dimension")
	}

	switch metric {
	case CosineDistance:
		return cosineDistance(v, other)
	case L2Distance:
		return l2Distance(v, other)
	case InnerProduct:
		return innerProduct(v, other)
	default:
		return cosineDistance(v, other)
	}
}

// cosineDistance calculates cosine distance
// distance = 1 - (dot(a,b) / (||a|| * ||b||))
func cosineDistance(a, b Vector) float32 {
	var dotProduct, normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 1.0 // Completely dissimilar
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	return 1.0 - (dotProduct / (normA * normB))
}

// l2Distance calculates Euclidean distance (L2 norm)
// distance = sqrt(sum((a[i] - b[i])^2))
func l2Distance(a, b Vector) float32 {
	var sum float32

	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}

// innerProduct calculates inner product distance
// distance = -sum(a[i] * b[i])
// Note: negative sign because larger values mean more similar when sorting
func innerProduct(a, b Vector) float32 {
	var sum float32

	for i := range a {
		sum += a[i] * b[i]
	}

	return -sum
}

// Normalize vector normalization (L2 norm)
func (v Vector) Normalize() Vector {
	var norm float32
	for _, val := range v {
		norm += val * val
	}

	if norm == 0 {
		return v
	}

	norm = float32(math.Sqrt(float64(norm)))
	normalized := make(Vector, len(v))
	for i, val := range v {
		normalized[i] = val / norm
	}

	return normalized
}

// Dim returns vector dimension
func (v Vector) Dim() int {
	return len(v)
}

// DiversityStrategy diversity strategy
type DiversityStrategy string

const (
	// DiversityByHash deduplication based on semantic hash
	DiversityByHash DiversityStrategy = "hash"

	// DiversityByDistance deduplication based on vector distance
	DiversityByDistance DiversityStrategy = "distance"

	// DiversityByMMR uses MMR (Maximal Marginal Relevance) algorithm
	DiversityByMMR DiversityStrategy = "mmr"
)

// DiversityParams diversity query parameters
type DiversityParams struct {
	// Enabled whether diversity is enabled
	Enabled bool

	// Strategy diversity strategy
	Strategy DiversityStrategy

	// HashField semantic hash field name (for DiversityByHash)
	// Example: "semantic_hash", "content_hash"
	HashField string

	// MinDistance minimum distance between results (for DiversityByDistance)
	// Example: 0.3 means distance between results is at least 0.3
	MinDistance float32

	// Lambda MMR balance parameter (for DiversityByMMR)
	// Range: 0-1
	// 0 = complete diversity
	// 1 = complete relevance
	// 0.5 = balanced (recommended)
	Lambda float32

	// OverFetchFactor over-fetch factor
	// First fetch TopK * OverFetchFactor results, then apply diversity filtering
	// Default: 5 (fetch 5x results then filter)
	OverFetchFactor int
}
