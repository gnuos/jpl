package gc

import "sync"

// ============================================================================
// 托管对象接口
// ============================================================================

// ManagedObject 可被 GC 管理的堆对象
type ManagedObject interface {
	// ObjID 返回对象唯一标识
	ObjID() uint64
	// GetRefCount 返回当前引用计数
	GetRefCount() int
	// IncRef 增加引用计数
	IncRef()
	// DecRef 减少引用计数，归零时自动释放
	DecRef()
	// IsAlive 检查对象是否存活
	IsAlive() bool
	// MarkChildren 通知 GC 标记子对象（用于循环检测）
	// 参数 marker 由 GC 提供，对象调用 marker(child) 标记每个子对象
	MarkChildren(marker func(child any))
	// OnFree 释放资源时的回调（可选）
	OnFree()
}

// ============================================================================
// GC 统计
// ============================================================================

// Stats GC 统计信息
type Stats struct {
	TotalAllocated uint64 // 累计分配对象数
	TotalFreed     uint64 // 累计释放对象数
	ActiveObjects  int    // 当前存活对象数
	CyclesFreed    uint64 // 循环引用释放次数
}

// ============================================================================
// GC 核心
// ============================================================================

// GC 垃圾回收器 — 引用计数 + 循环检测
type GC struct {
	mu      sync.Mutex
	objects map[uint64]ManagedObject // 所有托管对象
	nextID  uint64                   // 下一个可用 ID
	stats   Stats                    // 统计信息
}

// New 创建新的 GC 实例
func New() *GC {
	return &GC{
		objects: make(map[uint64]ManagedObject),
	}
}

// ============================================================================
// 对象分配
// ============================================================================

// NextID 获取下一个可用的对象 ID（线程安全）
func (g *GC) NextID() uint64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nextID++
	return g.nextID
}

// Register 将托管对象注册到 GC
func (g *GC) Register(obj ManagedObject) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.objects[obj.ObjID()] = obj
	g.stats.TotalAllocated++
	g.stats.ActiveObjects++
}

// Unregister 从 GC 注销对象（不调用 OnFree）
func (g *GC) Unregister(id uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.objects[id]; ok {
		delete(g.objects, id)
		g.stats.TotalFreed++
		g.stats.ActiveObjects--
	}
}

// ============================================================================
// 引用计数
// ============================================================================

// IncRef 增加对象引用计数
func (g *GC) IncRef(obj ManagedObject) {
	if obj == nil {
		return
	}
	obj.IncRef()
}

// DecRef 减少对象引用计数，归零时自动释放
func (g *GC) DecRef(obj ManagedObject) {
	if obj == nil {
		return
	}
	obj.DecRef()
	// 检查是否需要释放
	if !obj.IsAlive() {
		g.FreeObject(obj)
	}
}

// ============================================================================
// 循环检测（三色标记 + 清扫）
// ============================================================================

// Collect 执行垃圾回收
// roots: GC 外部的根对象列表（如 VM 的寄存器、全局变量）
func (g *GC) Collect(roots []any) Stats {
	g.mu.Lock()

	// 白色集合：所有存活对象（初始状态）
	white := make(map[uint64]bool, len(g.objects))
	for id, obj := range g.objects {
		if obj.IsAlive() {
			white[id] = true
		}
	}

	// 黑色集合：已标记为可达
	black := make(map[uint64]bool)

	// 标记函数：从一个对象出发递归标记
	var mark func(objID uint64)
	mark = func(objID uint64) {
		if !white[objID] {
			return // 已标记或不存在
		}
		delete(white, objID)
		black[objID] = true

		obj := g.objects[objID]
		if obj == nil {
			return
		}
		// 让对象标记其子对象
		obj.MarkChildren(func(child any) {
			if mo, ok := child.(ManagedObject); ok {
				mark(mo.ObjID())
			}
		})
	}

	// 从根对象开始标记
	for _, root := range roots {
		if mo, ok := root.(ManagedObject); ok {
			mark(mo.ObjID())
		}
	}

	// 收集白色对象（不可达 = 循环引用）
	freed := make([]ManagedObject, 0, len(white))
	for id := range white {
		if obj, ok := g.objects[id]; ok {
			freed = append(freed, obj)
			delete(g.objects, id)
			g.stats.TotalFreed++
			g.stats.ActiveObjects--
			g.stats.CyclesFreed++
		}
	}

	stats := g.stats
	g.mu.Unlock()

	// 在锁外释放资源，避免死锁
	for _, obj := range freed {
		obj.OnFree()
	}

	return stats
}

// ============================================================================
// 根对象提供者接口
// ============================================================================

// RootProvider 根对象提供者（engine 的 VM 实现此接口）
type RootProvider interface {
	// GCRoots 返回所有 GC 根对象
	GCRoots() []any
}

// ============================================================================
// 手动释放
// ============================================================================

// FreeObject 立即释放指定对象（不通过引用计数）
func (g *GC) FreeObject(obj ManagedObject) {
	if obj == nil {
		return
	}
	g.mu.Lock()
	_, exists := g.objects[obj.ObjID()]
	if exists {
		delete(g.objects, obj.ObjID())
		g.stats.TotalFreed++
		g.stats.ActiveObjects--
	}
	g.mu.Unlock()

	if exists {
		obj.OnFree()
	}
}

// Sweep 执行循环引用检测和清理
// providers: 实现 RootProvider 接口的对象（如 VM）
func (g *GC) Sweep(providers ...RootProvider) Stats {
	// 收集所有根对象
	var roots []any
	for _, p := range providers {
		roots = append(roots, p.GCRoots()...)
	}
	return g.Collect(roots)
}

// ============================================================================
// 统计查询
// ============================================================================

// GetStats 返回当前 GC 统计信息
func (g *GC) GetStats() Stats {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.stats
}

// ActiveCount 返回当前存活对象数
func (g *GC) ActiveCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.stats.ActiveObjects
}

// Reset 重置 GC 状态（仅用于测试）
func (g *GC) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.objects = make(map[uint64]ManagedObject)
	g.nextID = 0
	g.stats = Stats{}
}

// ============================================================================
// 内部方法
// ============================================================================

// freeObjectLocked 释放对象（需持有锁）
func (g *GC) freeObjectLocked(obj ManagedObject) {
	id := obj.ObjID()
	if _, ok := g.objects[id]; ok {
		delete(g.objects, id)
		g.stats.TotalFreed++
		g.stats.ActiveObjects--
	}
}
