package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

/* <Policy Parser SECTION - Optimized Version */

// ParsePolicyStringToTree 将策略字符串解析为策略树
// 优化点：
// 1. 添加错误返回
// 2. 不修改原始字符串，使用副本
// 3. 添加输入验证
func ParsePolicyStringToTreeOptimized(s string) (*PolicyNode, *AccessStruct, error) {
	if s == "" {
		return nil, nil, errors.New("策略字符串不能为空")
	}

	// 使用副本，不修改原始字符串
	original := s
	AS := NewAccessStruct()
	if err := AS.ParsePolicyStringtoMapOptimized(s); err != nil {
		return nil, nil, fmt.Errorf("解析策略映射失败: %w", err)
	}

	// 标准化策略字符串
	normalized := strings.ReplaceAll(original, "AND", "&&")
	normalized = strings.ReplaceAll(normalized, "OR", "||")
	normalized = strings.ReplaceAll(normalized, " ", "")

	mainPolicy, _, err := ParsePolicyStringOptimized(AS, normalized, 0, len(normalized)-1)
	if err != nil {
		return nil, nil, fmt.Errorf("解析策略字符串失败: %w", err)
	}

	return mainPolicy, AS, nil
}

// ParsePolicyStringOptimized 递归解析策略字符串
// 优化点：
// 1. 添加错误处理
// 2. 添加递归深度限制
// 3. 改进逻辑可读性
func ParsePolicyStringOptimized(A *AccessStruct, s string, startPos, stopPos int) (*PolicyNode, int, error) {
	// 输入验证
	if startPos < 0 || stopPos >= len(s) || startPos > stopPos {
		return nil, 0, fmt.Errorf("无效的位置参数: start=%d, stop=%d, len=%d", startPos, stopPos, len(s))
	}

	// 检查递归深度（防止栈溢出）
	if A.CurrentPointer > 1000 {
		return nil, 0, errors.New("策略嵌套层级过深")
	}

	this := NewPolicyNode("ThreshHold", 0)

	// 初始化访问结构
	A.A = append(A.A, make([]int, 2))
	ID := A.CurrentPointer
	A.A[ID][0] = 0
	A.A[ID][1] = 0
	A.CurrentPointer++

	policyChildren := make([]*PolicyNode, 0)
	var trueChildBuilder strings.Builder // 使用 Builder 提高效率
	childCount := 0
	thresholdCount := 0

	i := startPos + 1
	for i <= stopPos {
		// 查找左括号
		searchEnd := stopPos + 1
		if searchEnd > len(s) {
			searchEnd = len(s)
		}
		leftPos := strings.Index(s[i:searchEnd], "(")

		if leftPos != -1 {
			// 处理括号前的内容
			trueChildBuilder.WriteString(s[i : i+leftPos])

			// 查找匹配的右括号
			rightPos, err := findMatchingBracket(s, i+leftPos)
			if err != nil {
				return nil, 0, fmt.Errorf("括号不匹配: %w", err)
			}
			if rightPos > stopPos {
				return nil, 0, errors.New("括号超出范围")
			}

			// 递归解析子策略
			tmpPolicy, tmpID, err := ParsePolicyStringOptimized(A, s, i+leftPos, rightPos)
			if err != nil {
				return nil, 0, fmt.Errorf("解析子策略失败: %w", err)
			}

			policyChildren = append(policyChildren, tmpPolicy)
			A.A[ID] = append(A.A[ID], tmpID)
			childCount++
			i = rightPos + 1
		} else {
			// 没有更多括号，处理剩余内容（不包括最外层的右括号）
			if stopPos < len(s) && s[stopPos] == ')' {
				trueChildBuilder.WriteString(s[i:stopPos])
			} else {
				trueChildBuilder.WriteString(s[i : stopPos+1])
			}
			break
		}
	}

	trueChild := trueChildBuilder.String()

	// 处理 AND 和 OR 操作
	var childAttr []string
	if strings.Contains(trueChild, "&&") {
		childAttr = strings.Split(trueChild, "&&")

		for _, attr := range childAttr {
			attr = strings.TrimSpace(attr)
			if attr != "" {
				attrID, exists := A.PolicyMap[attr]
				if !exists {
					return nil, 0, fmt.Errorf("未知的属性: %s", attr)
				}
				policyChildren = append(policyChildren, NewPolicyNode(attr, 1).SetMax(1).SetMin(1))
				A.A[ID] = append(A.A[ID], -attrID)
				A.LeafID--
				childCount++
			}
		}
		thresholdCount = childCount // AND 需要所有子节点
		this.SetOperation(1)
	} else if strings.Contains(trueChild, "||") {
		childAttr = strings.Split(trueChild, "||")

		for _, attr := range childAttr {
			attr = strings.TrimSpace(attr)
			if attr != "" {
				attrID, exists := A.PolicyMap[attr]
				if !exists {
					return nil, 0, fmt.Errorf("未知的属性: %s", attr)
				}
				policyChildren = append(policyChildren, NewPolicyNode(attr, 1).SetMax(1).SetMin(1))
				A.A[ID] = append(A.A[ID], -attrID)
				A.LeafID--
				childCount++
			}
		}
		thresholdCount = 1 // OR 只需要一个子节点
		this.SetOperation(2)
	}

	if childCount == 0 {
		return nil, 0, errors.New("策略描述错误：没有有效的子节点")
	}

	this.SetChildren(policyChildren)
	this.SetMax(childCount)
	this.SetMin(thresholdCount)
	A.A[ID][0] = childCount
	A.A[ID][1] = thresholdCount

	return this, ID, nil
}

// findMatchingBracket 查找匹配的右括号
// 优化点：
// 1. 使用计数器代替递归，避免栈溢出
// 2. 添加错误返回
// 3. 改进逻辑清晰度
func findMatchingBracket(s string, leftPos int) (int, error) {
	if leftPos >= len(s) || s[leftPos] != '(' {
		return -1, errors.New("起始位置不是左括号")
	}

	depth := 1
	for i := leftPos + 1; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}

	return -1, errors.New("未找到匹配的右括号")
}

// ParsePolicyStringtoMapOptimized 优化的策略字符串映射解析
func (as *AccessStruct) ParsePolicyStringtoMapOptimized(s string) error {
	if s == "" {
		return errors.New("策略字符串不能为空")
	}

	as.PolicyMap = make(map[string]int)

	// 标准化字符串
	normalized := strings.ReplaceAll(s, "AND", ",")
	normalized = strings.ReplaceAll(normalized, "OR", ",")
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "(", "")
	normalized = strings.ReplaceAll(normalized, ")", "")

	attrs := strings.Split(normalized, ",")

	// 预分配切片容量
	as.PolicyMaps = make([]string, 0, len(attrs)+1)
	as.PolicyMaps = append(as.PolicyMaps, "") // 索引0为空

	index := 1
	for _, attr := range attrs {
		attr = strings.TrimSpace(attr)
		if attr != "" {
			// 避免重复属性
			if _, exists := as.PolicyMap[attr]; !exists {
				as.PolicyMap[attr] = index
				as.PolicyMaps = append(as.PolicyMaps, attr)
				index++
			}
		}
	}

	if len(as.PolicyMap) == 0 {
		return errors.New("策略字符串中没有有效的属性")
	}

	return nil
}

/* Policy Parser SECTION> */

/* <Utility SECTION - Optimized */

// CharToStringOptimized 优化的字符重复函数
// 优化点：使用 strings.Repeat 代替循环拼接
func CharToStringOptimized(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}

// GetPaddingOptimized 优化的填充生成函数
// 优化点：
// 1. 使用 strings.Builder
// 2. 减少字符串拼接次数
// 3. 添加参数验证
func GetPaddingOptimized(m, l, depth int) (string, error) {
	if depth < 1 || l < 1 || l > depth {
		return "", fmt.Errorf("无效的参数: m=%d, l=%d, depth=%d", m, l, depth)
	}

	var builder strings.Builder
	builder.Grow(depth - 1) // 预分配容量

	// 生成二进制表示或零填充
	if m == 0 {
		builder.WriteString(strings.Repeat("0", l-1))
	} else {
		binaryStr := strconv.FormatUint(uint64(m), 2)
		zerosNeeded := l - 2 - len(binaryStr)
		if zerosNeeded > 0 {
			builder.WriteString(strings.Repeat("0", zerosNeeded))
		}
		builder.WriteString(binaryStr)
	}

	// 添加星号填充
	builder.WriteString(strings.Repeat("*", depth-l))

	result := builder.String()
	if len(result) < depth-1 {
		return "", errors.New("生成的填充长度不足")
	}

	return result[len(result)-(depth-1):], nil
}

// CheckAttrNameOptimized 优化的属性名称检查
// 优化点：避免重复分割
func CheckAttrNameOptimized(attrName, authorityName string) bool {
	if attrName == "" || authorityName == "" {
		return false
	}

	colonIndex := strings.Index(attrName, ":")
	if colonIndex == -1 || colonIndex == 0 || colonIndex == len(attrName)-1 {
		return false
	}

	return attrName[:colonIndex] == authorityName
}

// GetAuthorityNameFromAttrNameOptimized 优化的权威名称提取
// 优化点：
// 1. 使用 Index 代替 SplitN
// 2. 添加更多验证
func GetAuthorityNameFromAttrNameOptimized(attrName string) (string, error) {
	if attrName == "" {
		return "", errors.New("属性名称不能为空")
	}

	colonIndex := strings.Index(attrName, ":")
	if colonIndex == -1 {
		return "", errors.New("属性名称格式错误：缺少冒号分隔符")
	}
	if colonIndex == 0 {
		return "", errors.New("属性名称格式错误：权威名称为空")
	}
	if colonIndex == len(attrName)-1 {
		return "", errors.New("属性名称格式错误：属性值为空")
	}

	return attrName[:colonIndex], nil
}

/* Utility SECTION> */
