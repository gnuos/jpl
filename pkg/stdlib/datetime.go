package stdlib

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gnuos/jpl/engine"
)

// RegisterDateTime 注册日期时间相关内置函数到引擎。
//
// 注册的函数：
//   - time: 返回当前 Unix 时间戳（秒级 float64）
//   - date: 格式化时间戳为日期字符串
//   - now: 返回当前时间对象或格式化字符串
//   - sleep: 暂停执行指定毫秒数
//   - microtime: 返回高精度时间戳
//   - getdate: 返回日期信息数组
//   - gettimeofday: 返回时间信息数组
//   - strftime: 格式化本地时间
//   - gmdate: 格式化 GMT 时间
//   - localtime: 返回本地时间信息
//   - mktime: 生成本地时间戳
//   - gmmktime: 生成 GMT 时间戳
//
// 参数：
//   - e: 引擎实例
func RegisterDateTime(e *engine.Engine) {
	e.RegisterFunc("time", builtinTime)
	e.RegisterFunc("date", builtinDate)
	e.RegisterFunc("now", builtinNow)
	e.RegisterFunc("sleep", builtinSleep)
	e.RegisterFunc("microtime", builtinMicrotime)

	// Phase 7.5: 扩展函数
	e.RegisterFunc("getdate", builtinGetdate)
	e.RegisterFunc("gettimeofday", builtinGettimeofday)
	e.RegisterFunc("strftime", builtinStrftime)
	e.RegisterFunc("gmdate", builtinGmdate)
	e.RegisterFunc("localtime", builtinLocaltime)
	e.RegisterFunc("mktime", builtinMktime)
	e.RegisterFunc("gmmktime", builtinGmmktime)

	// 模块注册 — import "datetime" 可用
	e.RegisterModule("datetime", map[string]engine.GoFunction{
		"time":      builtinTime,
		"date":      builtinDate,
		"now":       builtinNow,
		"sleep":     builtinSleep,
		"microtime": builtinMicrotime,
		// Phase 7.5
		"getdate":      builtinGetdate,
		"gettimeofday": builtinGettimeofday,
		"strftime":     builtinStrftime,
		"gmdate":       builtinGmdate,
		"localtime":    builtinLocaltime,
		"mktime":       builtinMktime,
		"gmmktime":     builtinGmmktime,
	})
}

// DateTimeNames 返回日期时间函数名称列表。
//
// 返回值：
//   - []string: 函数名列表
func DateTimeNames() []string {
	return []string{
		"time", "date", "now", "sleep", "microtime",
		"getdate", "gettimeofday", "strftime", "gmdate",
		"localtime", "mktime", "gmmktime",
	}
}

// builtinTime 返回当前 Unix 时间戳（秒级，float64）。
//
// 返回自 1970-01-01 00:00:00 UTC 以来的秒数，包含小数部分表示毫秒/微秒。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - float64: Unix 时间戳（秒）
//   - error: 无
//
// 使用示例：
//
//	time()              // → 1711209600.123456
//	int(time())         // → 1711209600（整数秒）
func builtinTime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("time() expects 0 arguments, got %d", len(args))
	}
	return engine.NewFloat(float64(time.Now().UnixNano()) / 1e9), nil
}

// builtinDate 格式化时间戳为日期字符串。
//
// 支持 PHP 风格的格式化字符：
//   - Y: 4位年份, y: 2位年份
//   - m: 月份(01-12), n: 月份(1-12, 无前导零)
//   - d: 日(01-31), j: 日(1-31, 无前导零)
//   - H: 时(00-23), G: 时(0-23, 无前导零)
//   - i: 分(00-59)
//   - s: 秒(00-59)
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 格式字符串
//   - args[1]: 可选的时间戳（float64），不提供则使用当前时间
//
// 返回值：
//   - string: 格式化后的日期字符串
//   - error: 参数错误
//
// 使用示例：
//
//	date("Y-m-d")               // → "2026-03-26"
//	date("Y-m-d H:i:s")         // → "2026-03-26 12:30:45"
//	date("Y/m/d", time())       // → "2026/03/26"
func builtinDate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("date() expects 1-2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("date() argument 1 must be a string, got %s", args[0].Type())
	}
	format := args[0].String()

	var t time.Time
	if len(args) == 2 {
		// 使用指定时间戳
		ts := args[1].Float()
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec).UTC()
	} else {
		t = time.Now().UTC()
	}

	result := formatDate(format, t)
	return engine.NewString(result), nil
}

// formatDate 按 JPL 格式化日期
func formatDate(format string, t time.Time) string {
	var b strings.Builder
	i := 0
	for i < len(format) {
		ch := format[i]
		switch ch {
		case 'Y':
			b.WriteString(fmt.Sprintf("%04d", t.Year()))
		case 'y':
			b.WriteString(fmt.Sprintf("%02d", t.Year()%100))
		case 'm':
			b.WriteString(fmt.Sprintf("%02d", int(t.Month())))
		case 'n':
			b.WriteString(fmt.Sprintf("%d", int(t.Month())))
		case 'd':
			b.WriteString(fmt.Sprintf("%02d", t.Day()))
		case 'j':
			b.WriteString(fmt.Sprintf("%d", t.Day()))
		case 'H':
			b.WriteString(fmt.Sprintf("%02d", t.Hour()))
		case 'G':
			b.WriteString(fmt.Sprintf("%d", t.Hour()))
		case 'i':
			b.WriteString(fmt.Sprintf("%02d", t.Minute()))
		case 's':
			b.WriteString(fmt.Sprintf("%02d", t.Second()))
		default:
			b.WriteByte(ch)
		}
		i++
	}
	return b.String()
}

// builtinNow 返回当前时间的对象表示或格式化字符串。
//
// 无参数调用返回包含完整时间信息的对象，有参数时返回格式化字符串。
//
// 返回对象字段：
//   - year: 年份
//   - month: 月份(1-12)
//   - day: 日(1-31)
//   - hour: 时(0-23)
//   - minute: 分(0-59)
//   - second: 秒(0-59)
//   - weekday: 星期(0=周日, 6=周六)
//   - millisecond: 毫秒
//   - timezone: 时区名
//   - timestamp: Unix 时间戳
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 可选的格式字符串
//
// 返回值：
//   - object/string: 时间对象或格式化字符串
//   - error: 参数错误
//
// 使用示例：
//
//	now()                  // → {year: 2026, month: 3, day: 26, ...}
//	now("Y-m-d")           // → "2026-03-26"
//	now("H:i:s")           // → "12:30:45"
func builtinNow(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("now() expects 0-1 arguments, got %d", len(args))
	}

	t := time.Now()

	// 如果提供参数，作为格式字符串返回
	if len(args) == 1 && args[0].Type() == engine.TypeString {
		format := args[0].String()
		result := formatDate(format, t.UTC())
		return engine.NewString(result), nil
	}

	// 返回完整对象
	_, offset := t.Zone()
	obj := map[string]engine.Value{
		"year":        engine.NewInt(int64(t.Year())),
		"month":       engine.NewInt(int64(t.Month())),
		"day":         engine.NewInt(int64(t.Day())),
		"hour":        engine.NewInt(int64(t.Hour())),
		"minute":      engine.NewInt(int64(t.Minute())),
		"second":      engine.NewInt(int64(t.Second())),
		"millisecond": engine.NewInt(int64(t.Nanosecond() / 1e6)),
		"weekday":     engine.NewInt(int64(t.Weekday())),
		"timezone":    engine.NewString(timezoneFromOffset(offset)),
		"timestamp":   engine.NewFloat(float64(t.UnixNano()) / 1e9),
	}
	return engine.NewObject(obj), nil
}

// timezoneFromOffset 将偏移秒数转换为时区字符串
func timezoneFromOffset(offset int) string {
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / 3600
	mins := (offset % 3600) / 60
	return fmt.Sprintf("%s%02d:%02d", sign, hours, mins)
}

// builtinSleep 暂停执行指定毫秒数。
//
// 支持中断：如果 VM 被中断（如 Ctrl+C），会提前结束等待。
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 暂停毫秒数（非负整数）
//
// 返回值：
//   - null: 总是返回 null
//   - error: 参数错误
//
// 使用示例：
//
//	sleep(1000)            // 暂停 1 秒
//	sleep(500)             // 暂停 0.5 秒
func builtinSleep(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("sleep() expects 1 argument, got %d", len(args))
	}

	ms := args[0].Int()
	if ms < 0 {
		return nil, fmt.Errorf("sleep() argument must be non-negative")
	}

	duration := time.Duration(ms) * time.Millisecond

	// 通过 interrupt channel 实现可中断的 sleep
	vm := ctx.VM()
	if vm != nil {
		select {
		case <-time.After(duration):
			// 正常等待完毕
		case <-vm.InterruptChannel():
			// 中断触发，缩短等待提前返回
		}
	} else {
		time.Sleep(duration)
	}

	return engine.NewNull(), nil
}

// builtinMicrotime 返回微秒级时间戳。
//
// 返回自 Unix 纪元以来的微秒数，用于高精度计时。
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - float64: 微秒级时间戳
//   - error: 无
//
// 使用示例：
//
//	microtime()            // → 1711209600.123456
func builtinMicrotime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("microtime() expects 0 arguments, got %d", len(args))
	}

	ns := time.Now().UnixNano()
	us := float64(ns) / 1e3
	// 截断到微秒精度
	us = math.Round(us) / 1e6
	return engine.NewFloat(us), nil
}

// ============================================================================
// Phase 7.5: 日期时间扩展函数
// ============================================================================

// builtinGetdate 返回日期信息对象（PHP 风格）。
//
// 返回包含完整日期时间信息的对象。
//
// 返回对象字段：
//   - year: 年份
//   - month: 月份(1-12)
//   - day: 日(1-31)
//   - hour: 时(0-23)
//   - minute: 分(0-59)
//   - second: 秒(0-59)
//   - weekday: 星期(0=周日, 6=周六)
//   - yday: 年中的第几天(0-365)
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 可选的时间戳，不提供则使用当前时间
//
// 返回值：
//   - object: 日期信息对象
//   - error: 参数错误
//
// 使用示例：
//
//	getdate()              // → {year: 2026, month: 3, ...}
//	getdate(time())        // → 使用指定时间戳
func builtinGetdate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	var t time.Time
	if len(args) == 0 {
		t = time.Now()
	} else if len(args) == 1 {
		ts := args[0].Float()
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec)
	} else {
		return nil, fmt.Errorf("getdate() expects 0-1 arguments, got %d", len(args))
	}

	obj := map[string]engine.Value{
		"year":    engine.NewInt(int64(t.Year())),
		"month":   engine.NewInt(int64(t.Month())),
		"day":     engine.NewInt(int64(t.Day())),
		"hour":    engine.NewInt(int64(t.Hour())),
		"minute":  engine.NewInt(int64(t.Minute())),
		"second":  engine.NewInt(int64(t.Second())),
		"weekday": engine.NewInt(int64(t.Weekday())),
		"yday":    engine.NewInt(int64(t.YearDay() - 1)), // 0-based day of year
	}
	return engine.NewObject(obj), nil
}

// builtinGettimeofday 返回时间信息对象（PHP 风格）。
//
// 返回对象字段：
//   - sec: Unix 秒
//   - usec: 微秒
//   - minuteswest: GMT 以西分钟数
//   - dsttime: 夏令时标志
//
// 参数：
//   - ctx: 执行上下文
//   - args: 无参数
//
// 返回值：
//   - object: 时间信息对象
//   - error: 无
//
// 使用示例：
//
//	gettimeofday()         // → {sec: 1711209600, usec: 123456, ...}
func builtinGettimeofday(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("gettimeofday() expects 0 arguments, got %d", len(args))
	}

	now := time.Now()
	sec := now.Unix()
	usec := (now.UnixNano() / 1e3) % 1e6

	_, offset := now.Zone()
	minuteswest := -offset / 60

	obj := map[string]engine.Value{
		"sec":         engine.NewInt(sec),
		"usec":        engine.NewInt(usec),
		"minuteswest": engine.NewInt(int64(minuteswest)),
		"dsttime":     engine.NewInt(0), // Simplified
	}
	return engine.NewObject(obj), nil
}

// builtinStrftime 按 strftime 格式化时间。
//
// 支持的格式符：
//   - %Y: 4位年份, %y: 2位年份
//   - %m: 月份(01-12), %d: 日(01-31)
//   - %H: 时(00-23), %M: 分(00-59), %S: 秒(00-59)
//   - %a: 缩写星期, %A: 完整星期
//   - %b: 缩写月份, %B: 完整月份
//
// 参数：
//   - ctx: 执行上下文
//   - args[0]: 格式字符串
//   - args[1]: 可选的时间戳
//
// 返回值：
//   - string: 格式化后的字符串
//   - error: 参数错误
//
// 使用示例：
//
//	strftime("%Y-%m-%d")          // → "2026-03-26"
//	strftime("%Y-%m-%d %H:%M:%S") // → "2026-03-26 12:30:45"
func builtinStrftime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("strftime() expects 1-2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strftime() argument 1 must be a string")
	}
	format := args[0].String()

	var t time.Time
	if len(args) == 2 {
		ts := args[1].Float()
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec)
	} else {
		t = time.Now()
	}

	result := strftimeFormat(format, t)
	return engine.NewString(result), nil
}

// strftimeFormat 实现 strftime 格式化
func strftimeFormat(format string, t time.Time) string {
	var b strings.Builder
	i := 0
	for i < len(format) {
		if i < len(format)-1 && format[i] == '%' {
			ch := format[i+1]
			switch ch {
			case 'Y':
				b.WriteString(fmt.Sprintf("%04d", t.Year()))
			case 'y':
				b.WriteString(fmt.Sprintf("%02d", t.Year()%100))
			case 'm':
				b.WriteString(fmt.Sprintf("%02d", int(t.Month())))
			case 'd':
				b.WriteString(fmt.Sprintf("%02d", t.Day()))
			case 'H':
				b.WriteString(fmt.Sprintf("%02d", t.Hour()))
			case 'M':
				b.WriteString(fmt.Sprintf("%02d", t.Minute()))
			case 'S':
				b.WriteString(fmt.Sprintf("%02d", t.Second()))
			case 'A':
				b.WriteString(t.Weekday().String())
			case 'a':
				b.WriteString(t.Weekday().String()[:3])
			case 'B':
				b.WriteString(t.Month().String())
			case 'b':
				b.WriteString(t.Month().String()[:3])
			default:
				b.WriteByte('%')
				b.WriteByte(ch)
			}
			i += 2
		} else {
			b.WriteByte(format[i])
			i++
		}
	}
	return b.String()
}

// builtinGmdate 格式化 GMT/UTC 时间
func builtinGmdate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("gmdate() expects 1-2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("gmdate() argument 1 must be a string")
	}
	format := args[0].String()

	var t time.Time
	if len(args) == 2 {
		ts := args[1].Float()
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec).UTC()
	} else {
		t = time.Now().UTC()
	}

	result := formatDate(format, t)
	return engine.NewString(result), nil
}

// builtinLocaltime 返回本地时间信息（PHP 风格）
func builtinLocaltime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	var t time.Time
	var isAssoc bool = false

	if len(args) >= 1 {
		ts := args[0].Float()
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		t = time.Unix(sec, nsec)
	} else {
		t = time.Now()
	}

	if len(args) >= 2 {
		isAssoc = engine.IsTruthy(args[1])
	}

	if isAssoc {
		obj := map[string]engine.Value{
			"tm_sec":   engine.NewInt(int64(t.Second())),
			"tm_min":   engine.NewInt(int64(t.Minute())),
			"tm_hour":  engine.NewInt(int64(t.Hour())),
			"tm_mday":  engine.NewInt(int64(t.Day())),
			"tm_mon":   engine.NewInt(int64(t.Month() - 1)), // 0-based
			"tm_year":  engine.NewInt(int64(t.Year() - 1900)),
			"tm_wday":  engine.NewInt(int64(t.Weekday())),
			"tm_yday":  engine.NewInt(int64(t.YearDay() - 1)),
			"tm_isdst": engine.NewInt(0),
		}
		return engine.NewObject(obj), nil
	}

	// Numeric indices (array)
	arr := []engine.Value{
		engine.NewInt(int64(t.Second())),      // 0: sec
		engine.NewInt(int64(t.Minute())),      // 1: min
		engine.NewInt(int64(t.Hour())),        // 2: hour
		engine.NewInt(int64(t.Day())),         // 3: mday
		engine.NewInt(int64(t.Month() - 1)),   // 4: mon (0-based)
		engine.NewInt(int64(t.Year() - 1900)), // 5: year
		engine.NewInt(int64(t.Weekday())),     // 6: wday
		engine.NewInt(int64(t.YearDay() - 1)), // 7: yday
		engine.NewInt(0),                      // 8: isdst
	}
	return engine.NewArray(arr), nil
}

// builtinMktime 生成时间戳（本地时间）
// mktime(hour, minute, second, month, day, year)
func builtinMktime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 6 {
		return nil, fmt.Errorf("mktime() expects 6 arguments, got %d", len(args))
	}

	hour := int(args[0].Int())
	minute := int(args[1].Int())
	second := int(args[2].Int())
	month := int(args[3].Int())
	day := int(args[4].Int())
	year := int(args[5].Int())

	// Handle 2-digit year
	if year < 100 {
		if year < 70 {
			year += 2000
		} else {
			year += 1900
		}
	}

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local)
	return engine.NewFloat(float64(t.Unix())), nil
}

// builtinGmmktime 生成时间戳（GMT 时间）
func builtinGmmktime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 6 {
		return nil, fmt.Errorf("gmmktime() expects 6 arguments, got %d", len(args))
	}

	hour := int(args[0].Int())
	minute := int(args[1].Int())
	second := int(args[2].Int())
	month := int(args[3].Int())
	day := int(args[4].Int())
	year := int(args[5].Int())

	// Handle 2-digit year
	if year < 100 {
		if year < 70 {
			year += 2000
		} else {
			year += 1900
		}
	}

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return engine.NewFloat(float64(t.Unix())), nil
}
