package gc_test

import (
	"testing"

	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/gc"
)

// ============================================================================
// 端到端内存泄漏测试
// ============================================================================

// TestIntegration_ArrayLifecycle 测试数组的完整生命周期
func TestIntegration_ArrayLifecycle(t *testing.T) {
	g := gc.New()

	// 创建数组（引用计数 = 1）
	arr := createManagedArray(g, 100)

	stats := g.GetStats()
	if stats.ActiveObjects != 1 {
		t.Errorf("expected 1 active object, got %d", stats.ActiveObjects)
	}

	// 释放引用
	releaseManaged(g, arr)

	stats = g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects after release, got %d", stats.ActiveObjects)
	}
}

// TestIntegration_ObjectLifecycle 测试对象的完整生命周期
func TestIntegration_ObjectLifecycle(t *testing.T) {
	g := gc.New()

	obj := createManagedObject(g, 50)

	stats := g.GetStats()
	if stats.ActiveObjects != 1 {
		t.Errorf("expected 1 active object, got %d", stats.ActiveObjects)
	}

	releaseManaged(g, obj)

	stats = g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects after release, got %d", stats.ActiveObjects)
	}
}

// TestIntegration_NestedArrays 测试嵌套数组的引用管理
func TestIntegration_NestedArrays(t *testing.T) {
	g := gc.New()

	// 创建外层数组
	outer := createManagedArray(g, 0)
	// 创建内层数组
	inner := createManagedArray(g, 5)

	// outer 引用 inner（增加引用计数）
	g.IncRef(inner)

	stats := g.GetStats()
	if stats.ActiveObjects != 2 {
		t.Errorf("expected 2 active objects, got %d", stats.ActiveObjects)
	}

	// 释放 outer（inner 仍有引用，refcount = 2）
	releaseManaged(g, outer)
	stats = g.GetStats()
	if stats.ActiveObjects != 1 {
		t.Errorf("expected 1 active object after releasing outer, got %d", stats.ActiveObjects)
	}

	// 释放 outer 对 inner 的引用
	g.DecRef(inner) // refcount: 2 -> 1
	stats = g.GetStats()
	if stats.ActiveObjects != 1 {
		t.Errorf("expected 1 active object (inner still in GC registry), got %d", stats.ActiveObjects)
	}

	// 释放 inner 的原始引用
	releaseManaged(g, inner) // refcount: 1 -> 0, freed
	stats = g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects after releasing inner, got %d", stats.ActiveObjects)
	}
}

// TestIntegration_CycleDetection 测试不可达对象的清理
func TestIntegration_CycleDetection(t *testing.T) {
	g := gc.New()

	// 创建两个对象（无外部引用）
	_ = createManagedObject(g, 0)
	_ = createManagedObject(g, 0)

	stats := g.GetStats()
	if stats.ActiveObjects != 2 {
		t.Errorf("expected 2 active objects, got %d", stats.ActiveObjects)
	}

	// 执行循环检测（无外部根）
	stats = g.Collect([]any{})

	// 两个对象都不可达，应被释放
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects after sweep, got %d", stats.ActiveObjects)
	}
}

// TestIntegration_CollectWithRoots 测试有根对象时的可达性
func TestIntegration_CollectWithRoots(t *testing.T) {
	g := gc.New()

	objA := createManagedObject(g, 0)
	_ = createManagedObject(g, 0)

	// objA 作为根对象
	stats := g.Collect([]any{objA})

	// objA 可达，objB 不可达
	if stats.ActiveObjects != 1 {
		t.Errorf("expected 1 active object (objA), got %d", stats.ActiveObjects)
	}
	if stats.CyclesFreed != 1 {
		t.Errorf("expected 1 cycle freed (objB), got %d", stats.CyclesFreed)
	}
}

// TestIntegration_ManyObjects 测试大量对象的分配和释放
func TestIntegration_ManyObjects(t *testing.T) {
	g := gc.New()
	count := 1000

	// 分配大量对象
	objects := make([]gc.ManagedObject, count)
	for i := range count {
		if i%2 == 0 {
			objects[i] = createManagedArray(g, i%10)
		} else {
			objects[i] = createManagedObject(g, i%10)
		}
	}

	stats := g.GetStats()
	if stats.ActiveObjects != count {
		t.Errorf("expected %d active objects, got %d", count, stats.ActiveObjects)
	}
	if stats.TotalAllocated != uint64(count) {
		t.Errorf("expected %d total allocated, got %d", count, stats.TotalAllocated)
	}

	// 释放所有对象
	for _, obj := range objects {
		releaseManaged(g, obj)
	}

	stats = g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects after release, got %d", stats.ActiveObjects)
	}
}

// TestIntegration_VMRegisters 测试 VM 寄存器中的 GC 管理
func TestIntegration_VMRegisters(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	g := gc.New()
	vm := engine.NewTestVM(e)
	vm.SetGC(g)

	// 模拟 VM 操作：创建数组并存储到寄存器
	arr := engine.NewArrayGC([]engine.Value{engine.NewInt(1), engine.NewInt(2)}, g)

	// 验证 GC 跟踪
	mo := engine.AsManagedObject(arr)
	if mo == nil {
		t.Fatal("array should be a managed object")
	}
	if mo.GetRefCount() != 1 {
		t.Errorf("expected refcount = 1, got %d", mo.GetRefCount())
	}

	stats := g.GetStats()
	if stats.ActiveObjects != 1 {
		t.Errorf("expected 1 active object, got %d", stats.ActiveObjects)
	}

	// 模拟寄存器赋值（inc new, dec old）
	g.IncRef(mo) // 新引用
	g.DecRef(mo) // 释放新引用
	g.DecRef(mo) // 释放原始引用

	stats = g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}
}

// TestIntegration_VMRootProvider 测试 VM 作为根对象提供者
func TestIntegration_VMRootProvider(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	g := gc.New()
	vm := engine.NewTestVM(e)
	vm.SetGC(g)

	// 创建对象并设置为全局变量
	arr := engine.NewArrayGC([]engine.Value{engine.NewInt(1)}, g)
	vm.SetGlobal("myArray", arr)

	// 验证 GCRoots 包含该对象
	roots := vm.GCRoots()
	found := false
	for _, root := range roots {
		if mo, ok := root.(gc.ManagedObject); ok {
			if mo == engine.AsManagedObject(arr) {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("GC roots should contain the global array")
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

// createManagedArray 创建并注册到 GC 的测试数组
func createManagedArray(g *gc.GC, size int) gc.ManagedObject {
	elems := make([]engine.Value, size)
	for i := range elems {
		elems[i] = engine.NewInt(int64(i))
	}
	arr := engine.NewArrayGC(elems, g)
	return engine.AsManagedObject(arr)
}

// createManagedObject 创建并注册到 GC 的测试对象
func createManagedObject(g *gc.GC, size int) gc.ManagedObject {
	kv := make(map[string]engine.Value, size)
	for i := range size {
		kv[string(rune('a'+i))] = engine.NewInt(int64(i))
	}
	obj := engine.NewObjectGC(kv, g)
	return engine.AsManagedObject(obj)
}

// releaseManaged 释放托管对象
func releaseManaged(g *gc.GC, obj gc.ManagedObject) {
	if obj == nil {
		return
	}
	obj.DecRef()
	if !obj.IsAlive() {
		g.FreeObject(obj)
	}
}
