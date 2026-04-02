package engine

import (
	"fmt"
	"sync"
	"testing"
)

// ============================================================================
// 并发编译测试
// ============================================================================

func TestConcurrentCompile(t *testing.T) {
	goroutines := 100
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			script := fmt.Sprintf("$a = %d + %d; $a", id, id)
			prog, err := CompileString(script)
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: compile: %w", id, err)
				return
			}
			eng := NewEngine()
			defer eng.Close()
			vm := newVMWithProgram(eng, prog)
			if err := vm.Execute(); err != nil {
				errs <- fmt.Errorf("goroutine %d: execute: %w", id, err)
				return
			}
			result := vm.GetResult()
			if result.Int() != int64(id*2) {
				errs <- fmt.Errorf("goroutine %d: expected %d, got %d", id, id*2, result.Int())
				return
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// ============================================================================
// 共享 Engine 并发执行测试
// ============================================================================

func TestConcurrentSharedEngine(t *testing.T) {
	eng := NewEngine()
	defer eng.Close()

	goroutines := 100
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			script := fmt.Sprintf(`
				fn compute($x) {
					$r = 0
					for ($i = 0; $i < $x; $i = $i + 1) {
						$r = $r + $i
					}
					return $r
				}
				compute(%d)
			`, id)
			prog, err := CompileString(script)
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: compile: %w", id, err)
				return
			}
			vm := NewVMWithProgram(eng, prog)
			if err := vm.Execute(); err != nil {
				errs <- fmt.Errorf("goroutine %d: execute: %w", id, err)
				return
			}
			// sum of 0..(id-1) = id*(id-1)/2
			expected := int64(id) * int64(id-1) / 2
			result := vm.GetResult()
			if result.Int() != expected {
				errs <- fmt.Errorf("goroutine %d: expected %d, got %d", id, expected, result.Int())
				return
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// ============================================================================
// 并发递归函数测试
// ============================================================================

func TestConcurrentRecursion(t *testing.T) {
	goroutines := 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			n := 20 + id%5 // fib(20) ~ fib(24)
			script := fmt.Sprintf(`
				fn fib($n) {
					if ($n <= 1) { return $n }
					return fib($n - 1) + fib($n - 2)
				}
				fib(%d)
			`, n)
			eng := NewEngine()
			defer eng.Close()
			result, err := compileAndRun(script)
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", id, err)
				return
			}
			// 验证结果非零
			if result.Int() <= 0 {
				errs <- fmt.Errorf("goroutine %d: fib(%d) = %d, expected > 0", id, n, result.Int())
				return
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// ============================================================================
// 并发异常处理测试
// ============================================================================

func TestConcurrentTryCatch(t *testing.T) {
	goroutines := 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			script := `
				$caught = false
				try {
					throw "error"
				} catch ($e) {
					$caught = true
				}
				$caught
			`
			eng := NewEngine()
			defer eng.Close()
			result, err := compileAndRun(script)
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", id, err)
				return
			}
			if !result.Bool() {
				errs <- fmt.Errorf("goroutine %d: expected true, got %v", id, result)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// ============================================================================
// 并发闭包测试
// ============================================================================

func TestConcurrentClosures(t *testing.T) {
	goroutines := 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			script := fmt.Sprintf(`
				fn makeAdder($n) {
					return fn($x) { return $x + $n }
				}
				$add5 = makeAdder(5)
				$add10 = makeAdder(10)
				$add5(%d) + $add10(%d)
			`, id, id)
			eng := NewEngine()
			defer eng.Close()
			result, err := compileAndRun(script)
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: %w", id, err)
				return
			}
			expected := int64(id + 5 + id + 10)
			if result.Int() != expected {
				errs <- fmt.Errorf("goroutine %d: expected %d, got %d", id, expected, result.Int())
				return
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
