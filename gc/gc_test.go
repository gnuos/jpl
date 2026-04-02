package gc

import (
	"sync"
	"testing"
)

// ============================================================================
// 测试用 Mock 对象
// ============================================================================

// mockObject 测试用托管对象
type mockObject struct {
	id       uint64
	refCount int
	alive    bool
	gc       *GC
	children []ManagedObject
	freed    bool
}

func newMockObject(gc *GC) *mockObject {
	obj := &mockObject{
		id:    gc.NextID(),
		alive: true,
		gc:    gc,
	}
	gc.Register(obj)
	return obj
}

func (o *mockObject) ObjID() uint64    { return o.id }
func (o *mockObject) GetRefCount() int { return o.refCount }
func (o *mockObject) IsAlive() bool    { return o.alive }
func (o *mockObject) OnFree()          { o.freed = true }

func (o *mockObject) IncRef() {
	o.refCount++
}

func (o *mockObject) DecRef() {
	o.refCount--
	if o.refCount <= 0 {
		o.alive = false
	}
}

func (o *mockObject) MarkChildren(marker func(child any)) {
	for _, child := range o.children {
		marker(child)
	}
}

func (o *mockObject) addChild(child ManagedObject) {
	o.children = append(o.children, child)
}

// ============================================================================
// 基础功能测试
// ============================================================================

func TestGCNew(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}
	stats := g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}
}

func TestGCNextID(t *testing.T) {
	g := New()
	id1 := g.NextID()
	id2 := g.NextID()
	id3 := g.NextID()

	if id1 != 1 {
		t.Errorf("expected first ID = 1, got %d", id1)
	}
	if id2 != 2 {
		t.Errorf("expected second ID = 2, got %d", id2)
	}
	if id3 != 3 {
		t.Errorf("expected third ID = 3, got %d", id3)
	}
}

func TestGCRegister(t *testing.T) {
	g := New()
	obj := newMockObject(g)

	stats := g.GetStats()
	if stats.ActiveObjects != 1 {
		t.Errorf("expected 1 active object, got %d", stats.ActiveObjects)
	}
	if stats.TotalAllocated != 1 {
		t.Errorf("expected 1 total allocated, got %d", stats.TotalAllocated)
	}

	_ = obj
}

func TestGCUnregister(t *testing.T) {
	g := New()
	obj := newMockObject(g)
	g.Unregister(obj.ObjID())

	stats := g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}
	if stats.TotalFreed != 1 {
		t.Errorf("expected 1 total freed, got %d", stats.TotalFreed)
	}
}

// ============================================================================
// 引用计数测试
// ============================================================================

func TestGCIncDecRef(t *testing.T) {
	g := New()
	obj := newMockObject(g)

	g.IncRef(obj)
	if obj.GetRefCount() != 1 {
		t.Errorf("expected refcount = 1, got %d", obj.GetRefCount())
	}

	g.IncRef(obj)
	if obj.GetRefCount() != 2 {
		t.Errorf("expected refcount = 2, got %d", obj.GetRefCount())
	}

	g.DecRef(obj)
	if obj.GetRefCount() != 1 {
		t.Errorf("expected refcount = 1, got %d", obj.GetRefCount())
	}
	if !obj.IsAlive() {
		t.Error("object should still be alive")
	}
}

func TestGCDecRefToZero(t *testing.T) {
	g := New()
	obj := newMockObject(g)

	g.IncRef(obj)
	g.DecRef(obj)

	if obj.IsAlive() {
		t.Error("object should be dead after refcount reaches 0")
	}
	if !obj.freed {
		t.Error("OnFree should have been called")
	}

	stats := g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}
}

func TestGCIncRefNil(t *testing.T) {
	g := New()
	// 不应 panic
	g.IncRef(nil)
	g.DecRef(nil)
}

// ============================================================================
// 循环检测测试
// ============================================================================

func TestGCCollectNoCycles(t *testing.T) {
	g := New()

	// A -> B（无循环）
	objA := newMockObject(g)
	objB := newMockObject(g)
	objA.addChild(objB)

	// A 作为根对象
	stats := g.Collect([]any{objA})

	if stats.ActiveObjects != 2 {
		t.Errorf("expected 2 active objects, got %d", stats.ActiveObjects)
	}
	if stats.CyclesFreed != 0 {
		t.Errorf("expected 0 cycles freed, got %d", stats.CyclesFreed)
	}
}

func TestGCCollectSimpleCycle(t *testing.T) {
	g := New()

	// A <-> B（简单循环，无外部引用）
	objA := newMockObject(g)
	objB := newMockObject(g)
	objA.addChild(objB)
	objB.addChild(objA)

	// 无根对象，A 和 B 形成不可达循环
	stats := g.Collect([]any{})

	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}
	if stats.CyclesFreed != 2 {
		t.Errorf("expected 2 cycles freed, got %d", stats.CyclesFreed)
	}
}

func TestGCCollectCycleWithRoot(t *testing.T) {
	g := New()

	// A <-> B（循环），但 A 是根对象
	objA := newMockObject(g)
	objB := newMockObject(g)
	objA.addChild(objB)
	objB.addChild(objA)

	stats := g.Collect([]any{objA})

	// 从根可达，不应释放
	if stats.ActiveObjects != 2 {
		t.Errorf("expected 2 active objects, got %d", stats.ActiveObjects)
	}
	if stats.CyclesFreed != 0 {
		t.Errorf("expected 0 cycles freed, got %d", stats.CyclesFreed)
	}
}

func TestGCCollectThreeWayCycle(t *testing.T) {
	g := New()

	// A -> B -> C -> A（三元循环）
	objA := newMockObject(g)
	objB := newMockObject(g)
	objC := newMockObject(g)
	objA.addChild(objB)
	objB.addChild(objC)
	objC.addChild(objA)

	stats := g.Collect([]any{})

	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}
	if stats.CyclesFreed != 3 {
		t.Errorf("expected 3 cycles freed, got %d", stats.CyclesFreed)
	}
}

func TestGCCollectMixedReachableAndCycle(t *testing.T) {
	g := New()

	// 根 -> A -> B（可达）
	objA := newMockObject(g)
	objB := newMockObject(g)
	objA.addChild(objB)

	// C <-> D（不可达循环）
	objC := newMockObject(g)
	objD := newMockObject(g)
	objC.addChild(objD)
	objD.addChild(objC)

	stats := g.Collect([]any{objA})

	if stats.ActiveObjects != 2 {
		t.Errorf("expected 2 active objects, got %d", stats.ActiveObjects)
	}
	if stats.CyclesFreed != 2 {
		t.Errorf("expected 2 cycles freed, got %d", stats.CyclesFreed)
	}
}

func TestGCCollectEmptyRoots(t *testing.T) {
	g := New()

	obj := newMockObject(g)
	stats := g.Collect([]any{})

	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}

	_ = obj
}

// ============================================================================
// 手动释放测试
// ============================================================================

func TestGCFreeObject(t *testing.T) {
	g := New()
	obj := newMockObject(g)

	g.FreeObject(obj)

	if !obj.freed {
		t.Error("OnFree should have been called")
	}
	stats := g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects, got %d", stats.ActiveObjects)
	}
}

func TestGCFreeNilObject(t *testing.T) {
	g := New()
	// 不应 panic
	g.FreeObject(nil)
}

func TestGCFreeAlreadyFreed(t *testing.T) {
	g := New()
	obj := newMockObject(g)

	g.FreeObject(obj)
	// 再次释放不应 panic
	g.FreeObject(obj)

	stats := g.GetStats()
	if stats.TotalFreed != 1 {
		t.Errorf("expected 1 total freed, got %d", stats.TotalFreed)
	}
}

// ============================================================================
// 统计测试
// ============================================================================

func TestGCStats(t *testing.T) {
	g := New()

	obj1 := newMockObject(g)
	obj2 := newMockObject(g)
	obj3 := newMockObject(g)

	g.IncRef(obj1)
	g.DecRef(obj1) // obj1 释放

	stats := g.GetStats()
	if stats.TotalAllocated != 3 {
		t.Errorf("expected 3 total allocated, got %d", stats.TotalAllocated)
	}
	if stats.TotalFreed != 1 {
		t.Errorf("expected 1 total freed, got %d", stats.TotalFreed)
	}
	if stats.ActiveObjects != 2 {
		t.Errorf("expected 2 active objects, got %d", stats.ActiveObjects)
	}

	_ = obj2
	_ = obj3
}

func TestGCActiveCount(t *testing.T) {
	g := New()

	newMockObject(g)
	newMockObject(g)

	if g.ActiveCount() != 2 {
		t.Errorf("expected ActiveCount = 2, got %d", g.ActiveCount())
	}
}

func TestGCReset(t *testing.T) {
	g := New()

	newMockObject(g)
	newMockObject(g)
	g.Reset()

	stats := g.GetStats()
	if stats.ActiveObjects != 0 {
		t.Errorf("expected 0 active objects after reset, got %d", stats.ActiveObjects)
	}
	if stats.TotalAllocated != 0 {
		t.Errorf("expected 0 total allocated after reset, got %d", stats.TotalAllocated)
	}
}

// ============================================================================
// 并发安全测试
// ============================================================================

func TestGCConcurrentAlloc(t *testing.T) {
	g := New()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			obj := newMockObject(g)
			g.IncRef(obj)
			g.DecRef(obj)
		}()
	}

	wg.Wait()

	stats := g.GetStats()
	if stats.TotalAllocated != 100 {
		t.Errorf("expected 100 total allocated, got %d", stats.TotalAllocated)
	}
}

func TestGCConcurrentCollect(t *testing.T) {
	g := New()

	// 先分配一些对象
	for i := 0; i < 10; i++ {
		newMockObject(g)
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g.Collect([]any{})
		}()
	}

	wg.Wait()
	// 不应 panic 或死锁
}
