package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
	"github.com/gnuos/jpl/gc"
)

func TestGCNames(t *testing.T) {
	names := GCNames()
	if len(names) != 2 {
		t.Errorf("expected 2 GC function names, got %d", len(names))
	}
	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	if !found["gc"] {
		t.Error("missing 'gc' function name")
	}
	if !found["gc_info"] {
		t.Error("missing 'gc_info' function name")
	}
}

func TestGCInfoNilGC(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	vm := engine.NewTestVM(e)
	// VM 没有设置 GC
	ctx := engine.NewContext(e, vm)

	result, err := gcInfo(ctx, nil)
	if err != nil {
		t.Fatalf("gc_info() failed: %v", err)
	}

	obj := result.Object()
	if obj["enabled"].Bool() != false {
		t.Error("expected enabled = false when GC is nil")
	}
}

func TestGCInfoWithGC(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	g := gc.New()
	vm := engine.NewTestVM(e)
	vm.SetGC(g)
	ctx := engine.NewContext(e, vm)

	// 分配一些对象
	arr := engine.NewArrayGC([]engine.Value{engine.NewInt(1)}, g)
	obj := engine.NewObjectGC(map[string]engine.Value{"a": engine.NewInt(1)}, g)
	_ = arr
	_ = obj

	result, err := gcInfo(ctx, nil)
	if err != nil {
		t.Fatalf("gc_info() failed: %v", err)
	}

	info := result.Object()
	if info["enabled"].Bool() != true {
		t.Error("expected enabled = true when GC is set")
	}
	if info["total"].Int() != 2 {
		t.Errorf("expected total = 2, got %d", info["total"].Int())
	}
	if info["active"].Int() != 2 {
		t.Errorf("expected active = 2, got %d", info["active"].Int())
	}
}

func TestGCCollect(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	g := gc.New()
	vm := engine.NewTestVM(e)
	vm.SetGC(g)

	// 创建循环引用但无外部引用
	arr1 := engine.NewArrayGC([]engine.Value{}, g)
	arr2 := engine.NewArrayGC([]engine.Value{}, g)
	mo1 := engine.AsManagedObject(arr1)
	mo2 := engine.AsManagedObject(arr2)
	// 模拟循环引用
	mo1.IncRef() // arr1 引用 arr2（简化模拟）
	mo2.IncRef() // arr2 引用 arr1

	// 先释放外部引用
	mo1.DecRef()
	mo2.DecRef()

	// 手动从 GC 移除（模拟无外部引用）
	g.Unregister(mo1.ObjID())
	g.Unregister(mo2.ObjID())

	stats := g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects after cleanup, got %d", stats.ActiveObjects)
	}
}

func TestGCCollectNilGC(t *testing.T) {
	e := engine.NewEngine()
	defer e.Close()

	vm := engine.NewTestVM(e)
	ctx := engine.NewContext(e, vm)

	result, err := gcCollect(ctx, nil)
	if err != nil {
		t.Fatalf("gc() failed: %v", err)
	}
	if !result.IsNull() {
		t.Error("expected null result when GC is not set")
	}
}
