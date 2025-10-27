package model

import (
	"testing"
)

// 测试括号匹配函数
func TestFindMatchingBracket(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		leftPos  int
		expected int
		wantErr  bool
	}{
		{
			name:     "简单括号",
			input:    "(test)",
			leftPos:  0,
			expected: 5,
			wantErr:  false,
		},
		{
			name:     "嵌套括号",
			input:    "(a(b)c)",
			leftPos:  0,
			expected: 6,
			wantErr:  false,
		},
		{
			name:     "多层嵌套",
			input:    "(a(b(c)d)e)",
			leftPos:  0,
			expected: 10,
			wantErr:  false,
		},
		{
			name:     "不匹配的括号",
			input:    "(test",
			leftPos:  0,
			expected: -1,
			wantErr:  true,
		},
		{
			name:     "起始位置不是左括号",
			input:    "test)",
			leftPos:  0,
			expected: -1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findMatchingBracket(tt.input, tt.leftPos)
			if (err != nil) != tt.wantErr {
				t.Errorf("findMatchingBracket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("findMatchingBracket() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// 测试字符串重复优化
func TestCharToStringOptimized(t *testing.T) {
	tests := []struct {
		name     string
		char     string
		count    int
		expected string
	}{
		{"正常情况", "a", 5, "aaaaa"},
		{"零次重复", "b", 0, ""},
		{"负数重复", "c", -1, ""},
		{"空字符串", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CharToStringOptimized(tt.char, tt.count)
			if got != tt.expected {
				t.Errorf("CharToStringOptimized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// 测试属性名称检查优化
func TestCheckAttrNameOptimized(t *testing.T) {
	tests := []struct {
		name          string
		attrName      string
		authorityName string
		expected      bool
	}{
		{"正确格式", "Authority:Attr1", "Authority", true},
		{"错误的权威名称", "Authority:Attr1", "WrongAuth", false},
		{"缺少冒号", "AuthorityAttr1", "Authority", false},
		{"空属性名", "", "Authority", false},
		{"空权威名", "Authority:Attr1", "", false},
		{"冒号在开头", ":Attr1", "", false},
		{"冒号在末尾", "Authority:", "Authority", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckAttrNameOptimized(tt.attrName, tt.authorityName)
			if got != tt.expected {
				t.Errorf("CheckAttrNameOptimized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// 测试权威名称提取优化
func TestGetAuthorityNameFromAttrNameOptimized(t *testing.T) {
	tests := []struct {
		name     string
		attrName string
		expected string
		wantErr  bool
	}{
		{"正确格式", "Authority:Attr1", "Authority", false},
		{"多个冒号", "Auth:Attr:Value", "Auth", false},
		{"缺少冒号", "AuthorityAttr1", "", true},
		{"空字符串", "", "", true},
		{"冒号在开头", ":Attr1", "", true},
		{"冒号在末尾", "Authority:", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAuthorityNameFromAttrNameOptimized(tt.attrName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAuthorityNameFromAttrNameOptimized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("GetAuthorityNameFromAttrNameOptimized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// 测试 GetPadding 优化
func TestGetPaddingOptimized(t *testing.T) {
	tests := []struct {
		name    string
		m       int
		l       int
		depth   int
		wantErr bool
	}{
		{"正常情况1", 0, 2, 5, false},
		{"正常情况2", 5, 4, 8, false},
		{"无效深度", 1, 2, 0, true},
		{"l大于depth", 1, 5, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPaddingOptimized(tt.m, tt.l, tt.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPaddingOptimized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != tt.depth-1 {
				t.Errorf("GetPaddingOptimized() length = %v, want %v", len(got), tt.depth-1)
			}
		})
	}
}

// 性能基准测试：括号匹配
func BenchmarkFindMatchingBracket(b *testing.B) {
	input := "(a(b(c(d(e)f)g)h)i)"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findMatchingBracket(input, 0)
	}
}

// 性能基准测试：旧版本 LookForMyRightBraket（需要修改才能比较）
func BenchmarkCharToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CharToString("*", 100)
	}
}

func BenchmarkCharToStringOptimized(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CharToStringOptimized("*", 100)
	}
}

// 集成测试：完整策略解析
func TestParsePolicyStringToTreeOptimized(t *testing.T) {
	tests := []struct {
		name    string
		policy  string
		wantErr bool
	}{
		{
			name:    "简单AND策略",
			policy:  "(Attr1 AND Attr2)",
			wantErr: false,
		},
		{
			name:    "简单OR策略",
			policy:  "(Attr1 OR Attr2)",
			wantErr: false,
		},
		{
			name:    "嵌套策略",
			policy:  "((Attr1 AND Attr2) OR (Attr3 AND Attr4))",
			wantErr: false,
		},
		{
			name:    "空字符串",
			policy:  "",
			wantErr: true,
		},
		{
			name:    "只有操作符",
			policy:  "AND OR",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ParsePolicyStringToTreeOptimized(tt.policy)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePolicyStringToTreeOptimized() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// 测试 AccessStruct 映射解析优化
func TestAccessStructParseOptimized(t *testing.T) {
	tests := []struct {
		name       string
		policy     string
		wantErr    bool
		expectAttr []string
	}{
		{
			name:       "简单策略",
			policy:     "Attr1 AND Attr2",
			wantErr:    false,
			expectAttr: []string{"Attr1", "Attr2"},
		},
		{
			name:       "带括号策略",
			policy:     "(Attr1 OR Attr2) AND Attr3",
			wantErr:    false,
			expectAttr: []string{"Attr1", "Attr2", "Attr3"},
		},
		{
			name:       "空策略",
			policy:     "",
			wantErr:    true,
			expectAttr: nil,
		},
		{
			name:       "重复属性",
			policy:     "Attr1 AND Attr1 AND Attr2",
			wantErr:    false,
			expectAttr: []string{"Attr1", "Attr2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := NewAccessStruct()
			err := as.ParsePolicyStringtoMapOptimized(tt.policy)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePolicyStringtoMapOptimized() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(as.PolicyMap) != len(tt.expectAttr) {
					t.Errorf("PolicyMap length = %v, want %v", len(as.PolicyMap), len(tt.expectAttr))
				}

				for _, attr := range tt.expectAttr {
					if _, exists := as.PolicyMap[attr]; !exists {
						t.Errorf("Expected attribute %v not found in PolicyMap", attr)
					}
				}
			}
		})
	}
}
