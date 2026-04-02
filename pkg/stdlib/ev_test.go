package stdlib

import (
	"testing"

	"github.com/gnuos/jpl/engine"
)

// TestEvRegistryNew 测试创建注册表
func TestEvRegistryNew(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterEv(eng)

	result, err := builtinEvRegistryNew(nil, []engine.Value{})
	if err != nil {
		t.Fatalf("ev_registry_new() error = %v", err)
	}

	registry, ok := result.(*EvRegistryValue)
	if !ok {
		t.Fatalf("ev_registry_new() returned %T, expected *EvRegistryValue", result)
	}

	if registry.Count() != 0 {
		t.Errorf("new registry count = %d, expected 0", registry.Count())
	}
}

// TestEvLoopNew 测试创建事件循环
func TestEvLoopNew(t *testing.T) {
	eng := engine.NewEngine()
	defer eng.Close()
	RegisterEv(eng)

	result, err := builtinEvLoopNew(nil, []engine.Value{})
	if err != nil {
		t.Fatalf("ev_loop_new() error = %v", err)
	}

	loop, ok := result.(*EvLoopValue)
	if !ok {
		t.Fatalf("ev_loop_new() returned %T, expected *EvLoopValue", result)
	}

	if loop.running {
		t.Error("new loop should not be running")
	}
}

// TestEvTimerNow 测试微秒级时间
func TestEvTimerNow(t *testing.T) {
	result, err := builtinEvTimerNow(nil, []engine.Value{})
	if err != nil {
		t.Fatalf("ev_timer_now() error = %v", err)
	}

	us := result.Int()
	if us <= 0 {
		t.Errorf("ev_timer_now() = %d, expected positive value", us)
	}
}

// TestEvRegistryTimer 测试定时器注册
func TestEvRegistryTimer(t *testing.T) {
	registry := NewEvRegistry()

	args := []engine.Value{
		registry,
		engine.NewInt(1000),
		engine.NewInt(1),
	}

	timerID, err := builtinRegistryOnTimer(nil, args)
	if err != nil {
		t.Fatalf("on_timer() error = %v", err)
	}

	if timerID.Int() != 1 {
		t.Errorf("timer id = %d, expected 1", timerID.Int())
	}

	if registry.Count() != 1 {
		t.Errorf("registry count = %d, expected 1", registry.Count())
	}
}

// TestEvRegistrySignal 测试信号注册
func TestEvRegistrySignal(t *testing.T) {
	registry := NewEvRegistry()

	args := []engine.Value{
		registry,
		engine.NewInt(2),
		engine.NewInt(1),
	}

	result, err := builtinRegistryOnSignal(nil, args)
	if err != nil {
		t.Fatalf("on_signal() error = %v", err)
	}

	if !result.Bool() {
		t.Error("on_signal() should return true")
	}

	if registry.Count() != 1 {
		t.Errorf("registry count = %d, expected 1", registry.Count())
	}
}

// TestEvRegistryClear 测试清空注册表
func TestEvRegistryClear(t *testing.T) {
	registry := NewEvRegistry()

	registry.timers[1] = &evTimer{active: true}
	registry.signals[1] = &evSignalHandler{}

	if registry.Count() != 2 {
		t.Errorf("registry count before clear = %d, expected 2", registry.Count())
	}

	registry.cancelAll()
	registry.handlers = make(map[int]*evHandler)
	registry.timers = make(map[int]*evTimer)
	registry.signals = make(map[int]*evSignalHandler)

	if registry.Count() != 0 {
		t.Errorf("registry count after clear = %d, expected 0", registry.Count())
	}
}
