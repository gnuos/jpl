package stdlib

import (
	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/gc"
)

// RegisterGC 注册 GC 相关内置函数
func RegisterGC(e *engine.Engine) {
	e.RegisterFunc("gc", gcCollect)
	e.RegisterFunc("gc_info", gcInfo)
}

// GCNames 返回 GC 内置函数名称列表
func GCNames() []string {
	return []string{"gc", "gc_info"}
}

// gcCollect 手动触发垃圾回收
// 用法: gc()
func gcCollect(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	vm := ctx.VM()
	g := vm.GetGC()
	if g == nil {
		// GC 未启用，静默返回
		return engine.NewNull(), nil
	}

	// 执行循环引用检测和清理
	stats := g.Sweep(vm)

	return engine.NewObject(map[string]engine.Value{
		"cycles_freed": engine.NewInt(int64(stats.CyclesFreed)),
		"active":       engine.NewInt(int64(stats.ActiveObjects)),
	}), nil
}

// gcInfo 返回 GC 统计信息
// 用法: gc_info()  ->  {total: N, active: N, freed: N, cycles_freed: N}
func gcInfo(ctx *engine.Context, args []engine.Value) (engine.Value, error) {
	vm := ctx.VM()
	g := vm.GetGC()
	if g == nil {
		return engine.NewObject(map[string]engine.Value{
			"enabled": engine.NewBool(false),
		}), nil
	}

	stats := g.GetStats()
	return engine.NewObject(map[string]engine.Value{
		"enabled":      engine.NewBool(true),
		"total":        engine.NewInt(int64(stats.TotalAllocated)),
		"active":       engine.NewInt(int64(stats.ActiveObjects)),
		"freed":        engine.NewInt(int64(stats.TotalFreed)),
		"cycles_freed": engine.NewInt(int64(stats.CyclesFreed)),
	}), nil
}

// RegisterGCInfo 将 GC 统计信息注册为引擎常量（供脚本查询）
func RegisterGCInfo(e *engine.Engine, g *gc.GC) {
	if g == nil {
		return
	}
	e.RegisterConst("__gc_enabled__", engine.NewBool(true))
}
