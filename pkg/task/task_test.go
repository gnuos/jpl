package task

import (
	"encoding/json"
	"testing"
)

// ============================================================================
// TaskDef JSON 解析测试
// ============================================================================

func TestTaskDef_UnmarshalJSON_String(t *testing.T) {
	data := []byte(`"jpl run build.jpl"`)
	var td TaskDef
	if err := json.Unmarshal(data, &td); err != nil {
		t.Fatalf("UnmarshalJSON() failed: %v", err)
	}
	if td.Cmd != "jpl run build.jpl" {
		t.Errorf("Cmd = %q, want %q", td.Cmd, "jpl run build.jpl")
	}
	if len(td.Deps) != 0 {
		t.Errorf("Deps = %v, want empty", td.Deps)
	}
}

func TestTaskDef_UnmarshalJSON_Object(t *testing.T) {
	data := []byte(`{"cmd": "jpl run build.jpl", "deps": ["clean", "lint"]}`)
	var td TaskDef
	if err := json.Unmarshal(data, &td); err != nil {
		t.Fatalf("UnmarshalJSON() failed: %v", err)
	}
	if td.Cmd != "jpl run build.jpl" {
		t.Errorf("Cmd = %q, want %q", td.Cmd, "jpl run build.jpl")
	}
	if len(td.Deps) != 2 || td.Deps[0] != "clean" || td.Deps[1] != "lint" {
		t.Errorf("Deps = %v, want [clean lint]", td.Deps)
	}
}

func TestTaskDef_UnmarshalJSON_ObjectNoCmd(t *testing.T) {
	data := []byte(`{"deps": ["clean"]}`)
	var td TaskDef
	err := json.Unmarshal(data, &td)
	if err == nil {
		t.Error("UnmarshalJSON() should fail for empty cmd")
	}
}

func TestTaskDef_UnmarshalJSON_Invalid(t *testing.T) {
	data := []byte(`123`)
	var td TaskDef
	err := json.Unmarshal(data, &td)
	if err == nil {
		t.Error("UnmarshalJSON() should fail for number")
	}
}

func TestTaskDef_MarshalJSON_String(t *testing.T) {
	td := TaskDef{Cmd: "echo hello"}
	data, err := json.Marshal(td)
	if err != nil {
		t.Fatalf("MarshalJSON() failed: %v", err)
	}
	if string(data) != `"echo hello"` {
		t.Errorf("MarshalJSON() = %s, want %q", data, `"echo hello"`)
	}
}

func TestTaskDef_MarshalJSON_Object(t *testing.T) {
	td := TaskDef{Cmd: "jpl run build.jpl", Deps: []string{"clean"}}
	data, err := json.Marshal(td)
	if err != nil {
		t.Fatalf("MarshalJSON() failed: %v", err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatalf("result is not valid JSON object: %v", err)
	}
	if obj["cmd"] != "jpl run build.jpl" {
		t.Errorf("cmd = %v, want %q", obj["cmd"], "jpl run build.jpl")
	}
}

// ============================================================================
// 任务清单 JSON 解析测试
// ============================================================================

func TestManifestTasks_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"tasks": {
			"test": "jpl run test.jpl",
			"build": {"cmd": "jpl run build.jpl", "deps": ["clean"]},
			"clean": "rm -rf build"
		}
	}`)

	var result struct {
		Tasks map[string]TaskDef `json:"tasks"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("UnmarshalJSON() failed: %v", err)
	}

	if len(result.Tasks) != 3 {
		t.Fatalf("Tasks count = %d, want 3", len(result.Tasks))
	}

	if result.Tasks["test"].Cmd != "jpl run test.jpl" {
		t.Errorf("test.cmd = %q, want %q", result.Tasks["test"].Cmd, "jpl run test.jpl")
	}
	if result.Tasks["build"].Cmd != "jpl run build.jpl" {
		t.Errorf("build.cmd = %q, want %q", result.Tasks["build"].Cmd, "jpl run build.jpl")
	}
	if len(result.Tasks["build"].Deps) != 1 || result.Tasks["build"].Deps[0] != "clean" {
		t.Errorf("build.deps = %v, want [clean]", result.Tasks["build"].Deps)
	}
	if result.Tasks["clean"].Cmd != "rm -rf build" {
		t.Errorf("clean.cmd = %q, want %q", result.Tasks["clean"].Cmd, "rm -rf build")
	}
}

// ============================================================================
// ResolveTaskOrder 测试
// ============================================================================

func TestResolveTaskOrder_NoDeps(t *testing.T) {
	tasks := map[string]TaskDef{
		"test": {Cmd: "jpl run test.jpl"},
	}

	plan, err := ResolveTaskOrder(tasks, "test")
	if err != nil {
		t.Fatalf("ResolveTaskOrder() failed: %v", err)
	}
	if len(plan.Order) != 1 || plan.Order[0] != "test" {
		t.Errorf("Order = %v, want [test]", plan.Order)
	}
}

func TestResolveTaskOrder_WithDeps(t *testing.T) {
	tasks := map[string]TaskDef{
		"clean": {Cmd: "rm -rf build"},
		"lint":  {Cmd: "jpl run lint.jpl"},
		"build": {Cmd: "jpl run build.jpl", Deps: []string{"clean", "lint"}},
		"test":  {Cmd: "jpl run test.jpl", Deps: []string{"build"}},
	}

	plan, err := ResolveTaskOrder(tasks, "test")
	if err != nil {
		t.Fatalf("ResolveTaskOrder() failed: %v", err)
	}

	// test → build → clean,lint
	// clean 和 lint 应在 build 之前
	// build 应在 test 之前
	if len(plan.Order) != 4 {
		t.Fatalf("Order length = %d, want 4", len(plan.Order))
	}

	// 验证顺序约束
	indexOf := func(name string) int {
		for i, n := range plan.Order {
			if n == name {
				return i
			}
		}
		return -1
	}

	if indexOf("clean") >= indexOf("build") {
		t.Errorf("clean should come before build, got %v", plan.Order)
	}
	if indexOf("lint") >= indexOf("build") {
		t.Errorf("lint should come before build, got %v", plan.Order)
	}
	if indexOf("build") >= indexOf("test") {
		t.Errorf("build should come before test, got %v", plan.Order)
	}
}

func TestResolveTaskOrder_Deduplication(t *testing.T) {
	tasks := map[string]TaskDef{
		"clean": {Cmd: "rm -rf build"},
		"lint":  {Cmd: "jpl run lint.jpl", Deps: []string{"clean"}},
		"build": {Cmd: "jpl run build.jpl", Deps: []string{"clean"}},
		"test":  {Cmd: "jpl run test.jpl", Deps: []string{"lint", "build"}},
	}

	plan, err := ResolveTaskOrder(tasks, "test")
	if err != nil {
		t.Fatalf("ResolveTaskOrder() failed: %v", err)
	}

	// clean 只应出现一次
	cleanCount := 0
	for _, name := range plan.Order {
		if name == "clean" {
			cleanCount++
		}
	}
	if cleanCount != 1 {
		t.Errorf("clean appears %d times, want 1; order = %v", cleanCount, plan.Order)
	}
}

func TestResolveTaskOrder_CircularDependency(t *testing.T) {
	tasks := map[string]TaskDef{
		"a": {Cmd: "cmd a", Deps: []string{"b"}},
		"b": {Cmd: "cmd b", Deps: []string{"c"}},
		"c": {Cmd: "cmd c", Deps: []string{"a"}},
	}

	_, err := ResolveTaskOrder(tasks, "a")
	if err == nil {
		t.Error("ResolveTaskOrder() should detect circular dependency")
	}
}

func TestResolveTaskOrder_SelfCircular(t *testing.T) {
	tasks := map[string]TaskDef{
		"a": {Cmd: "cmd a", Deps: []string{"a"}},
	}

	_, err := ResolveTaskOrder(tasks, "a")
	if err == nil {
		t.Error("ResolveTaskOrder() should detect self-circular dependency")
	}
}

func TestResolveTaskOrder_TaskNotFound(t *testing.T) {
	tasks := map[string]TaskDef{
		"build": {Cmd: "jpl run build.jpl"},
	}

	_, err := ResolveTaskOrder(tasks, "test")
	if err == nil {
		t.Error("ResolveTaskOrder() should return error for missing task")
	}
}

func TestResolveTaskOrder_DepNotFound(t *testing.T) {
	tasks := map[string]TaskDef{
		"build": {Cmd: "jpl run build.jpl", Deps: []string{"nonexistent"}},
	}

	_, err := ResolveTaskOrder(tasks, "build")
	if err == nil {
		t.Error("ResolveTaskOrder() should return error for missing dependency")
	}
}

// ============================================================================
// isJPLCommand 测试
// ============================================================================

func TestIsJPLCommand(t *testing.T) {
	runner := &TaskRunner{ProjectDir: "/tmp"}

	tests := []struct {
		cmd  string
		want bool
	}{
		{"jpl run script.jpl", true},
		{"jpl run build.jpl --watch", true},
		{"scripts/build.jpl", true},
		{"scripts/build.jpl --watch", true},
		{"rm -rf build", false},
		{"echo hello", false},
		{"ls -la | grep foo", false}, // 含管道符
		{"jpl fmt file.jpl", false},  // 不是 jpl run
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := runner.isJPLCommand(tt.cmd)
			if got != tt.want {
				t.Errorf("isJPLCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}
