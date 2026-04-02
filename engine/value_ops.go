package engine

// IsNumeric 检查值是否为数值类型（int, float, bigint, bigdecimal）
func IsNumeric(v Value) bool {
	t := v.Type()
	return t == TypeInt || t == TypeFloat || t == TypeBigInt || t == TypeBigDecimal
}

// IsTruthy 检查值是否为真值
// truthy: 非 null、非零数字、非空字符串、非空数组、非空对象
func IsTruthy(v Value) bool {
	if v.IsNull() {
		return false
	}
	switch v.Type() {
	case TypeBool:
		return v.Bool()
	case TypeInt:
		return v.Int() != 0
	case TypeFloat:
		return v.Float() != 0.0
	case TypeString:
		return v.String() != ""
	case TypeArray:
		return len(v.Array()) > 0
	case TypeObject:
		return len(v.Object()) > 0
	case TypeFunc:
		return true
	case TypeBigDecimal:
		return v.(*BigDecimalValue).value.Sign() != 0
	case TypeRange:
		return true
	case TypeRegex:
		return true
	default:
		return false
	}
}

// IsComparable 检查两个值是否可以比较
func IsComparable(a, b Value) bool {
	// 数值类型之间可比较
	if IsNumeric(a) && IsNumeric(b) {
		return true
	}
	// 同类型可比较
	if a.Type() == b.Type() {
		return true
	}
	// null 和 bool 可比较
	if (a.Type() == TypeNull || a.Type() == TypeBool) &&
		(b.Type() == TypeNull || b.Type() == TypeBool) {
		return true
	}
	return false
}

// CoerceToInt 将值强制转换为 int64
func CoerceToInt(v Value) int64 {
	return v.Int()
}

// CoerceToFloat 将值强制转换为 float64
// int → float, 其他类型返回 0
func CoerceToFloat(v Value) float64 {
	return v.Float()
}

// CoerceToString 将值强制转换为字符串
func CoerceToString(v Value) string {
	return v.String()
}

// CoerceToBool 将值强制转换为布尔值
func CoerceToBool(v Value) bool {
	return IsTruthy(v)
}

// ValueAdd 执行加法，处理类型提升
func ValueAdd(a, b Value) Value {
	return a.Add(b)
}

// ValueSub 执行减法
func ValueSub(a, b Value) Value {
	return a.Sub(b)
}

// ValueMul 执行乘法
func ValueMul(a, b Value) Value {
	return a.Mul(b)
}

// ValueDiv 执行除法
func ValueDiv(a, b Value) Value {
	return a.Div(b)
}

// ValueMod 执行取模
func ValueMod(a, b Value) Value {
	return a.Mod(b)
}

// ValueNegate 执行取反
func ValueNegate(a Value) Value {
	return a.Negate()
}

// ValueLess 小于比较
func ValueLess(a, b Value) bool {
	return a.Less(b)
}

// ValueGreater 大于比较
func ValueGreater(a, b Value) bool {
	return a.Greater(b)
}

// ValueLessEqual 小于等于比较
func ValueLessEqual(a, b Value) bool {
	return a.LessEqual(b)
}

// ValueGreaterEqual 大于等于比较
func ValueGreaterEqual(a, b Value) bool {
	return a.GreaterEqual(b)
}

// ConcatValues 字符串拼接（.. 运算符）
func ConcatValues(a, b Value) Value {
	return NewString(a.String() + b.String())
}
