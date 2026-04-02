package main

import (
	"fmt"
	"os"

	"github.com/gnuos/jpl/pkg/pm"
	"github.com/gnuos/jpl/pkg/task"
	"github.com/spf13/cobra"
)

// ============================================================================
// jpl task - 运行定义在 jpl.json 中的任务
// ============================================================================

var (
	taskList   bool
	taskDryRun bool
)

var taskCmd = &cobra.Command{
	Use:   "task [name]",
	Short: "运行项目任务",
	Long: `运行在 jpl.json 中定义的任务。

支持的任务格式：
  简单字符串:  "test": "jpl run tests/main.jpl"
  带依赖对象:  "build": {"cmd": "jpl run build.jpl", "deps": ["clean", "lint"]}

示例：
  jpl task test           # 运行 test 任务（自动执行其依赖）
  jpl task --list         # 列出所有可用任务
  jpl task build --dry-run  # 显示执行顺序但不执行`,
	Args: cobra.MaximumNArgs(1),
	Run:  runTask,
}

func init() {
	taskCmd.Flags().BoolVar(&taskList, "list", false, "列出所有可用任务")
	taskCmd.Flags().BoolVar(&taskDryRun, "dry-run", false, "显示执行顺序但不执行")
	rootCmd.AddCommand(taskCmd)
}

func runTask(cmd *cobra.Command, args []string) {
	projectDir, _ := os.Getwd()

	// 加载清单
	manifest, err := pm.LoadOrCreateManifest(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// --list 模式
	if taskList {
		fmt.Printf("%s@%s tasks:\n", manifest.Name, manifest.Version)
		task.FormatTaskList(manifest.Tasks)
		return
	}

	// 需要任务名
	if len(args) == 0 {
		if len(manifest.Tasks) == 0 {
			fmt.Println("No tasks defined in jpl.json.")
			fmt.Println("\nAdd tasks to jpl.json:")
			fmt.Println(`  "tasks": {`)
			fmt.Println(`    "test": "jpl run tests/main.jpl"`)
			fmt.Println(`  }`)
			return
		}
		fmt.Fprintf(os.Stderr, "错误: 请指定任务名，或使用 --list 查看可用任务\n")
		os.Exit(1)
	}

	taskName := args[0]

	// 检查任务是否存在
	if !manifest.HasTask(taskName) {
		fmt.Fprintf(os.Stderr, "错误: 任务 %q 未定义\n", taskName)
		if len(manifest.Tasks) > 0 {
			fmt.Fprintf(os.Stderr, "\n可用任务:\n")
			task.FormatTaskList(manifest.Tasks)
		}
		os.Exit(1)
	}

	// 解析执行顺序
	plan, err := task.ResolveTaskOrder(manifest.Tasks, taskName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// dry-run 模式
	if taskDryRun {
		fmt.Printf("Execution plan for %q:\n", taskName)
		runner := task.NewTaskRunner(projectDir, verbose)
		runner.DryRun(plan)
		return
	}

	// 显示执行计划（如果有依赖）
	if len(plan.Order) > 1 {
		fmt.Printf("Running %q (with %d dependencies):\n", taskName, len(plan.Order)-1)
	} else {
		fmt.Printf("Running %q:\n", taskName)
	}

	// 执行
	runner := task.NewTaskRunner(projectDir, verbose)
	if err := runner.Run(plan); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] task %q completed successfully\n", taskName)
	}
}
