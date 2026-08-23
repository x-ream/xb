package xb

import (
	"testing"
	"uuid"
)

// 测试 InsertBuilder 处理 uuid.UUID
func TestInsertBuilderUUID(t *testing.T) {
	// 测试1: 正常 UUID
	t.Run("normal uuid", func(t *testing.T) {
		builder := &InsertBuilder{}
		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		builder.Set("id", testUUID)

		if len(builder.bbs) != 1 {
			t.Fatalf("Expected 1 field, got %d", len(builder.bbs))
		}

		if builder.bbs[0].Key != "id" {
			t.Errorf("Expected key 'id', got '%s'", builder.bbs[0].Key)
		}

		expected := "550e8400-e29b-41d4-a716-446655440000"
		if builder.bbs[0].Value != expected {
			t.Errorf("Expected value '%s', got '%v'", expected, builder.bbs[0].Value)
		}
		t.Logf("✓ UUID 被正确转换为字符串: %v", builder.bbs[0].Value)
	})

	// 测试2: Nil UUID
	t.Run("nil uuid", func(t *testing.T) {
		builder := &InsertBuilder{}
		var nilUUID uuid.UUID
		builder.Set("id", nilUUID)

		if len(builder.bbs) != 0 {
			t.Errorf("Expected 0 fields for nil UUID, got %d", len(builder.bbs))
		}
		t.Log("✓ Nil UUID 被正确忽略")
	})

	// 测试3: 新生成的 UUID
	t.Run("new uuid", func(t *testing.T) {
		builder := &InsertBuilder{}
		newUUID := uuid.NewV7()
		builder.Set("id", newUUID)

		if len(builder.bbs) != 1 {
			t.Fatalf("Expected 1 field, got %d", len(builder.bbs))
		}

		// 验证是有效的 UUID 字符串
		_, err := uuid.Parse(builder.bbs[0].Value.(string))
		if err != nil {
			t.Errorf("Value is not a valid UUID string: %v", err)
		}
		t.Logf("✓ 新生成的 UUID 被正确转换: %v", builder.bbs[0].Value)
	})
}

// 测试 UpdateBuilder 处理 uuid.UUID
func TestUpdateBuilderUUID(t *testing.T) {
	// 测试1: 正常 UUID
	t.Run("normal uuid", func(t *testing.T) {
		builder := &UpdateBuilder{}
		testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		builder.Set("id", testUUID)

		if len(builder.bbs) != 1 {
			t.Fatalf("Expected 1 field, got %d", len(builder.bbs))
		}

		expected := "550e8400-e29b-41d4-a716-446655440000"
		if builder.bbs[0].Value != expected {
			t.Errorf("Expected value '%s', got '%v'", expected, builder.bbs[0].Value)
		}
		t.Logf("✓ UUID 被正确转换为字符串: %v", builder.bbs[0].Value)
	})

	// 测试2: Nil UUID
	t.Run("nil uuid", func(t *testing.T) {
		builder := &UpdateBuilder{}
		var nilUUID uuid.UUID
		builder.Set("id", nilUUID)

		if len(builder.bbs) != 0 {
			t.Errorf("Expected 0 fields for nil UUID, got %d", len(builder.bbs))
		}
		t.Log("✓ Nil UUID 被正确忽略")
	})
}
