package stdlib

import (
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gnuos/jpl/engine"
)

func TestBuiltinTime(t *testing.T) {
	result, err := callBuiltin("time")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ts := result.Float()
	now := float64(time.Now().UnixNano()) / 1e9
	if math.Abs(ts-now) > 1.0 {
		t.Errorf("time() returned %v, expected ~%v", ts, now)
	}
}

func TestBuiltinTimeNoArgs(t *testing.T) {
	_, err := callBuiltin("time", engine.NewInt(1))
	if err == nil {
		t.Error("time() should reject arguments")
	}
}

func TestBuiltinDate(t *testing.T) {
	// 固定时间戳测试: 2025-01-15 10:30:45 UTC
	tm := time.Date(2025, 1, 15, 10, 30, 45, 0, time.UTC)
	ts := engine.NewFloat(float64(tm.Unix()))

	result, err := callBuiltin("date", engine.NewString("Y-m-d H:i:s"), ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "2025-01-15 10:30:45"
	if result.String() != expected {
		t.Errorf("date('Y-m-d H:i:s', %d) = %q, expected %q", tm.Unix(), result.String(), expected)
	}
}

func TestBuiltinDatePartial(t *testing.T) {
	tm := time.Date(2025, 3, 5, 8, 5, 9, 0, time.UTC)
	ts := engine.NewFloat(float64(tm.Unix()))

	result, err := callBuiltin("date", engine.NewString("j/n/Y G:i:s"), ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "5/3/2025 8:05:09"
	if result.String() != expected {
		t.Errorf("date('j/n/Y G:i:s') = %q, expected %q", result.String(), expected)
	}
}

func TestBuiltinDateNoTimestamp(t *testing.T) {
	result, err := callBuiltin("date", engine.NewString("Y-m-d"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 应该返回当前日期，格式为 YYYY-MM-DD
	// 使用正则表达式验证格式，避免时区或运行时间差异
	resultStr := result.String()
	if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, resultStr); !matched {
		t.Errorf("date('Y-m-d') = %q, expected format YYYY-MM-DD", resultStr)
	}
}

func TestBuiltinDateLiteralChars(t *testing.T) {
	result, err := callBuiltin("date", engine.NewString("Y年m月d日"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 验证格式：YYYY年MM月DD日
	resultStr := result.String()
	if matched, _ := regexp.MatchString(`^\d{4}年\d{2}月\d{2}日$`, resultStr); !matched {
		t.Errorf("date('Y年m月d日') = %q, expected format YYYY年MM月DD日", resultStr)
	}
}

func TestBuiltinNow(t *testing.T) {
	result, err := callBuiltin("now")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Type() != engine.TypeObject {
		t.Fatalf("now() should return object, got %s", result.Type())
	}

	obj := result.Object()

	// 检查所有字段存在且有有效的数值
	fields := []string{"year", "month", "day", "hour", "minute", "second", "weekday"}
	for _, f := range fields {
		v, ok := obj[f]
		if !ok {
			t.Errorf("now() missing field: %s", f)
			continue
		}
		// 验证字段值是有效的整数
		if v.Type() != engine.TypeInt {
			t.Errorf("now() field %s should be int, got %s", f, v.Type())
		}
	}

	// 验证日期字段在合理范围内
	if v, ok := obj["month"]; ok {
		month := v.Int()
		if month < 1 || month > 12 {
			t.Errorf("month out of range: %d", month)
		}
	}
	if v, ok := obj["day"]; ok {
		day := v.Int()
		if day < 1 || day > 31 {
			t.Errorf("day out of range: %d", day)
		}
	}
	if v, ok := obj["hour"]; ok {
		hour := v.Int()
		if hour < 0 || hour > 23 {
			t.Errorf("hour out of range: %d", hour)
		}
	}
	if v, ok := obj["minute"]; ok {
		minute := v.Int()
		if minute < 0 || minute > 59 {
			t.Errorf("minute out of range: %d", minute)
		}
	}
	if v, ok := obj["second"]; ok {
		second := v.Int()
		if second < 0 || second > 59 {
			t.Errorf("second out of range: %d", second)
		}
	}
}

func TestBuiltinSleep(t *testing.T) {
	start := time.Now()
	result, err := callBuiltin("sleep", engine.NewInt(100))
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsNull() {
		t.Errorf("sleep() should return null, got %v", result)
	}
	if elapsed < 90*time.Millisecond {
		t.Errorf("sleep(100) only waited %v", elapsed)
	}
}

func TestBuiltinSleepNegative(t *testing.T) {
	_, err := callBuiltin("sleep", engine.NewInt(-1))
	if err == nil {
		t.Error("sleep(-1) should return error")
	}
}

func TestBuiltinMicrotime(t *testing.T) {
	result, err := callBuiltin("microtime")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ts := result.Float()
	now := float64(time.Now().UnixNano()) / 1e9
	if math.Abs(ts-now) > 1.0 {
		t.Errorf("microtime() returned %v, expected ~%v", ts, now)
	}
}

func TestDateTimeIntegration(t *testing.T) {
	script := `$ts = time();
$result = date("Y-m-d H:i:s", $ts);
$result`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 应该返回当前时间的格式化字符串
	s := result.String()
	if len(s) != 19 { // "2006-01-02 15:04:05"
		t.Errorf("date() returned %q (len=%d), expected 19 chars", s, len(s))
	}
}

func TestNowIntegration(t *testing.T) {
	script := `$n = now();
$n.year`

	result, err := compileAndRunBuiltins(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	year := result.Int()
	nowYear := int64(time.Now().Year())
	if year != nowYear {
		t.Errorf("now().year = %d, expected %d", year, nowYear)
	}
}

// TestDateTimeNames 测试日期时间函数列表
func TestDateTimeNames(t *testing.T) {
	names := DateTimeNames()
	if len(names) != 12 {
		t.Errorf("expected 12 datetime function names, got %d", len(names))
	}
}

// TestGetdate 测试 getdate 函数
func TestGetdate(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	// Test getdate() without args
	fn := e.GetRegisteredFunc("getdate")
	if fn == nil {
		t.Fatal("getdate function not registered")
	}
}

// TestGettimeofday 测试 gettimeofday 函数
func TestGettimeofday(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	fn := e.GetRegisteredFunc("gettimeofday")
	if fn == nil {
		t.Fatal("gettimeofday function not registered")
	}
}

// TestStrftime 测试 strftime 函数
func TestStrftime(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	fn := e.GetRegisteredFunc("strftime")
	if fn == nil {
		t.Fatal("strftime function not registered")
	}
}

// TestGmdate 测试 gmdate 函数
func TestGmdate(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	fn := e.GetRegisteredFunc("gmdate")
	if fn == nil {
		t.Fatal("gmdate function not registered")
	}
}

// TestLocaltime 测试 localtime 函数
func TestLocaltime(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	fn := e.GetRegisteredFunc("localtime")
	if fn == nil {
		t.Fatal("localtime function not registered")
	}
}

// TestMktime 测试 mktime 函数
func TestMktime(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	fn := e.GetRegisteredFunc("mktime")
	if fn == nil {
		t.Fatal("mktime function not registered")
	}
}

// TestGmmktime 测试 gmmktime 函数
func TestGmmktime(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()
	RegisterDateTime(e)

	fn := e.GetRegisteredFunc("gmmktime")
	if fn == nil {
		t.Fatal("gmmktime function not registered")
	}
}

// TestStrftimeFormat 测试 strftime 格式化
func TestStrftimeFormat(t *testing.T) {
	tests := []struct {
		format   string
		contains []string
	}{
		{"%Y", []string{"202"}},              // Year should contain 202x
		{"%m", []string{"0", "1"}},           // Month
		{"%d", []string{"0", "1", "2", "3"}}, // Day
		{"%H:%M:%S", []string{":"}},          // Time format
	}

	for _, tt := range tests {
		result := strftimeFormat(tt.format, time.Now())
		found := false
		for _, substr := range tt.contains {
			if strings.Contains(result, substr) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("strftime(%q) = %q, expected to contain one of %v",
				tt.format, result, tt.contains)
		}
	}
}
