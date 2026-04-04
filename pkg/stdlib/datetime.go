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
	e.RegisterFunc("strtotime", builtinStrtotime)
	e.RegisterFunc("checkdate", builtinCheckdate)

	// P1
	e.RegisterFunc("date_diff", builtinDateDiff)
	e.RegisterFunc("date_add", builtinDateAdd)

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
		"strtotime":    builtinStrtotime,
		"checkdate":    builtinCheckdate,
		// P1
		"date_diff": builtinDateDiff,
		"date_add":  builtinDateAdd,
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
		"strtotime", "checkdate",
		"date_diff", "date_add",
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

// builtinStrtotime 解析日期时间字符串为 Unix 时间戳。
// 支持多种常见格式：
//   - "2006-01-02 15:04:05"
//   - "2006-01-02"
//   - "01/02/2006"
//   - "Jan 2, 2006"
//   - "2006-01-02T15:04:05Z" (ISO 8601)
//   - 相对时间："+1 day", "-2 weeks", "+3 months", "+1 year"
func builtinStrtotime(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("strtotime() expects 1-2 arguments, got %d", len(args))
	}

	if args[0].Type() != engine.TypeString {
		return nil, fmt.Errorf("strtotime() argument 1 must be a string, got %s", args[0].Type())
	}

	str := args[0].String()

	var base time.Time
	if len(args) == 2 {
		ts := args[1].Float()
		sec := int64(ts)
		nsec := int64((ts - float64(sec)) * 1e9)
		base = time.Unix(sec, nsec)
	} else {
		base = time.Now()
	}

	// 尝试相对时间格式：+1 day, -2 weeks, +3 months, +1 year
	if relTime, ok := parseRelativeTime(str, base); ok {
		return engine.NewFloat(float64(relTime.Unix())), nil
	}

	// 尝试多种日期格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04",
		"2006-01-02",
		"01/02/2006 15:04:05",
		"01/02/2006",
		"Jan 2, 2006 15:04:05",
		"Jan 2, 2006",
		"02-Jan-2006 15:04:05",
		"02-Jan-2006",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return engine.NewFloat(float64(t.Unix())), nil
		}
	}

	return engine.NewNull(), nil
}

// parseRelativeTime 解析相对时间字符串
func parseRelativeTime(str string, base time.Time) (time.Time, bool) {
	str = strings.TrimSpace(strings.ToLower(str))
	if !strings.HasPrefix(str, "+") && !strings.HasPrefix(str, "-") {
		return base, false
	}

	parts := strings.Fields(str)
	if len(parts) != 2 {
		return base, false
	}

	sign := 1
	numStr := parts[0]
	if numStr[0] == '-' {
		sign = -1
		numStr = numStr[1:]
	} else if numStr[0] == '+' {
		numStr = numStr[1:]
	}

	num := 0
	for _, c := range numStr {
		if c < '0' || c > '9' {
			return base, false
		}
		num = num*10 + int(c-'0')
	}
	num *= sign

	unit := parts[1]
	switch unit {
	case "second", "seconds", "sec", "secs":
		return base.Add(time.Duration(num) * time.Second), true
	case "minute", "minutes", "min", "mins":
		return base.Add(time.Duration(num) * time.Minute), true
	case "hour", "hours", "hr", "hrs":
		return base.Add(time.Duration(num) * time.Hour), true
	case "day", "days":
		return base.AddDate(0, 0, num), true
	case "week", "weeks":
		return base.AddDate(0, 0, num*7), true
	case "month", "months":
		return base.AddDate(0, num, 0), true
	case "year", "years":
		return base.AddDate(num, 0, 0), true
	default:
		return base, false
	}
}

// builtinCheckdate 验证公历日期是否有效。
func builtinCheckdate(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("checkdate() expects 3 arguments (month, day, year), got %d", len(args))
	}

	month := int(args[0].Int())
	day := int(args[1].Int())
	year := int(args[2].Int())

	if month < 1 || month > 12 {
		return engine.NewBool(false), nil
	}
	if year < 1 || year > 32767 {
		return engine.NewBool(false), nil
	}

	// 使用 time.Date 验证日期有效性
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	valid := t.Month() == time.Month(month) && t.Day() == day && t.Year() == year
	return engine.NewBool(valid), nil
}

// DateTimeSigs returns function signatures for REPL :doc command.

// builtinDateDiff 计算两个时间戳之间的差异。
func builtinDateDiff(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("date_diff() expects 2 arguments, got %d", len(args))
	}
	ts1 := args[0].Float()
	ts2 := args[1].Float()
	diff := ts2 - ts1
	if diff < 0 {
		diff = -diff
	}
	result := map[string]engine.Value{
		"seconds": engine.NewFloat(diff),
		"minutes": engine.NewFloat(diff / 60),
		"hours":   engine.NewFloat(diff / 3600),
		"days":    engine.NewFloat(diff / 86400),
	}
	return engine.NewObject(result), nil
}

// builtinDateAdd 给时间戳添加时间。
func builtinDateAdd(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("date_add() expects 2 arguments (timestamp, duration_str), got %d", len(args))
	}
	ts := args[0].Float()
	durStr := args[1].String()

	sec := int64(ts)
	nsec := int64((ts - float64(sec)) * 1e9)
	base := time.Unix(sec, nsec)

	parts := strings.Fields(durStr)
	if len(parts) != 2 {
		return nil, fmt.Errorf("date_add() duration must be like '1 day', '2 hours'")
	}
	num := 0
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return nil, fmt.Errorf("date_add() invalid duration number")
		}
		num = num*10 + int(c-'0')
	}

	unit := strings.ToLower(parts[1])
	var result time.Time
	switch unit {
	case "second", "seconds", "sec", "secs":
		result = base.Add(time.Duration(num) * time.Second)
	case "minute", "minutes", "min", "mins":
		result = base.Add(time.Duration(num) * time.Minute)
	case "hour", "hours", "hr", "hrs":
		result = base.Add(time.Duration(num) * time.Hour)
	case "day", "days":
		result = base.AddDate(0, 0, num)
	case "week", "weeks":
		result = base.AddDate(0, 0, num*7)
	case "month", "months":
		result = base.AddDate(0, num, 0)
	case "year", "years":
		result = base.AddDate(num, 0, 0)
	default:
		return nil, fmt.Errorf("date_add() unknown unit: %s", unit)
	}

	return engine.NewFloat(float64(result.Unix())), nil
}
func DateTimeSigs() map[string]string {
	return map[string]string{
		"time":         "time() → float  — Current Unix timestamp (seconds)",
		"now":          "now([format]) → object/string  — Current time or formatted string",
		"date":         "date(format, [timestamp]) → string  — Format timestamp",
		"sleep":        "sleep(ms) → null  — Sleep for milliseconds",
		"microtime":    "microtime() → float  — High-precision timestamp",
		"getdate":      "getdate([timestamp]) → object  — Get date info",
		"gettimeofday": "gettimeofday() → object  — Get time info with microseconds",
		"strftime":     "strftime(format, [timestamp]) → string  — Format with strftime",
		"gmdate":       "gmdate(format, [timestamp]) → string  — Format GMT time",
		"localtime":    "localtime([timestamp], [is_assoc]) → array/object  — Local time info",
		"mktime":       "mktime(hour, minute, second, month, day, year) → float  — Create timestamp",
		"gmmktime":     "gmmktime(hour, minute, second, month, day, year) → float  — Create GMT timestamp",
		"strtotime":    "strtotime(datetime_string, [base_timestamp]) → float  — Parse date string to timestamp",
		"checkdate":    "checkdate(month, day, year) → bool  — Validate Gregorian date",
		"date_diff":    "date_diff(ts1, ts2) → object  — Difference between two timestamps",
		"date_add":     "date_add(timestamp, duration_str) → float  — Add time to timestamp",
	}
}
