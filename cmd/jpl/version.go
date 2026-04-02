package main

import (
	"fmt"
	"strings"

	jpl "github.com/gnuos/jpl"
	"github.com/spf13/cobra"
)

// versionCmd 实现 jpl version 子命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  "显示 JPL 解释器的详细版本信息，包括版本号、发布日期、构建信息等。",
	Run: func(cmd *cobra.Command, args []string) {
		buildInfo := jpl.BuildInfo()

		fmt.Printf("JPL v%s\n", jpl.Version)
		fmt.Printf("Release Date: %s\n", jpl.ReleaseDate)

		if goVersion, ok := buildInfo["go"]; ok {
			fmt.Printf("Go Version: %s\n", goVersion)
		}

		if commit, ok := buildInfo["commit"]; ok && commit != "" {
			if len(commit) > 8 {
				commit = commit[:8]
			}
			fmt.Printf("Git Commit: %s\n", commit)
		}

		if buildTime, ok := buildInfo["build_time"]; ok && buildTime != "" {
			// 简化时间格式
			parts := strings.Split(buildTime, "T")
			if len(parts) > 0 {
				fmt.Printf("Build Time: %s\n", parts[0])
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
