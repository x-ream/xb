# Use `omitzero` JSON tag option for Qdrant request structs

Replace `omitempty` with `omitzero` (or add alongside) on Qdrant request struct fields, aligning with the Go 1.24 `encoding/json` best practice for distinguishing "unset" from "zero value" — without requiring pointer types for non-nil-able scalars.

## Motivation

Introduced in Go 1.24, `omitzero` is a JSON struct tag option that omits a field when its value is the zero value for its type. Unlike `omitempty`, it correctly handles `time.Time` and makes intent explicit: the field should be omitted because the value is the *default/unset* state, not just "empty".

Currently, Qdrant request structs in `to_qdrant_json.go` use a mix of strategies:

| Strategy | Example | Drawback |
|---|---|---|
| `*float32` + `omitempty` | `ScoreThreshold *float32 \`json:"score_threshold,omitempty"\`` | Requires pointer allocation, GC pressure, extra nil-check |
| `bool` + `omitempty` | `WithVector bool \`json:"with_vector,omitempty"\`` | Works, but `omitzero` is semantically clearer |
| `int` + `omitempty` | `HnswEf int \`json:"hnsw_ef,omitempty"\`` | Same as above |

**This change does NOT affect the builder-layer pointer requirement** (`NilOrNumber`, `**float32` interface methods, condition filtering). The builder itself still needs pointers to distinguish "not provided" from "explicitly set to 0". This PR targets **JSON serialization only**.

## Proposed Change

In `to_qdrant_json.go`, update the following struct fields:

### 1. Value-type fields: `omitempty` → `omitzero`

| Struct | Field | Before | After |
|---|---|---|---|
| `QdrantSearchParams` | `HnswEf` | `json:"hnsw_ef,omitempty"` | `json:"hnsw_ef,omitzero"` |
| `QdrantSearchParams` | `Exact` | `json:"exact,omitempty"` | `json:"exact,omitzero"` |
| `QdrantSearchParams` | `IndexedOnly` | `json:"indexed_only,omitempty"` | `json:"indexed_only,omitzero"` |
| `QdrantSearchRequest` | `WithVector` | `json:"with_vector,omitempty"` | `json:"with_vector,omitzero"` |
| `QdrantSearchRequest` | `Offset` | `json:"offset,omitempty"` | `json:"offset,omitzero"` |
| `QdrantRecommendRequest` | `WithVector` | `json:"with_vector,omitempty"` | `json:"with_vector,omitzero"` |
| `QdrantRecommendRequest` | `Offset` | `json:"offset,omitempty"` | `json:"offset,omitzero"` |
| `QdrantDiscoverRequest` | `WithVector` | `json:"with_vector,omitempty"` | `json:"with_vector,omitzero"` |
| `QdrantDiscoverRequest` | `Offset` | `json:"offset,omitempty"` | `json:"offset,omitzero"` |
| `QdrantScrollRequest` | `WithVector` | `json:"with_vector,omitempty"` | `json:"with_vector,omitzero"` |

### 2. Pointer-type fields: add `omitzero` alongside `omitempty`

| Struct | Field | Before | After |
|---|---|---|---|
| `QdrantSearchRequest` | `ScoreThreshold` | `json:"score_threshold,omitempty"` | `json:"score_threshold,omitempty,omitzero"` |
| `QdrantRecommendRequest` | `ScoreThreshold` | `json:"score_threshold,omitempty"` | `json:"score_threshold,omitempty,omitzero"` |
| `QdrantDiscoverRequest` | `ScoreThreshold` | `json:"score_threshold,omitempty"` | `json:"score_threshold,omitempty,omitzero"` |
| `QdrantRangeCondition` | `Gt`, `Gte`, `Lt`, `Lte` | `json:"...,omitempty"` | `json:"...,omitempty,omitzero"` |
| All Qdrant request structs | `Filter` | `json:"filter,omitempty"` | `json:"filter,omitempty,omitzero"` |
| All Qdrant request structs | `Params` | `json:"params,omitempty"` | `json:"params,omitempty,omitzero"` |

### 3. Unchanged fields

- `WithPayload interface{}` — not a scalar, stays `omitempty`
- `Positive []int64`, `Negative []int64`, `Context []int64` — slices, stays `omitempty`
- `Must []QdrantCondition`, `Should []QdrantCondition`, `MustNot []QdrantCondition` — slices
- `QdrantMatchCondition.Value interface{}`, `Any []interface{}` — not scalars

## Impact

- **Breaking change:** No. `omitzero` on `bool`/`int` behaves identically to `omitempty`. On pointer types, `omitzero` is additive and does not change behavior.
- **Behavioral change:** None. JSON output is unchanged for all existing inputs.
- **Tests:** All existing tests must continue passing (`go test ./...`).

## Non-Goals

- This issue does NOT propose removing pointer types from the builder/condition layer.
- This issue does NOT propose changing the `VectorDBRequest` interface (`GetScoreThreshold() **float32` etc.).
- This issue does NOT propose removing the `nil_able.go` helper functions.

## Related

- Go 1.24 release notes (omitzero): https://go.dev/doc/go1.24#encoding/json
- Go 1.24+: `omitzero` is the recommended way to omit zero-valued fields
