package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// 获取工作区根目录
	workDir, _ := os.Getwd()

	// 检查是否在 workspace 根目录
	if _, err := os.Stat(filepath.Join(workDir, "go.work")); os.IsNotExist(err) {
		fmt.Println("错误: 请在 Workspace 根目录运行")
		return
	}

	// 获取要构建的工具名称
	toolName := ""
	if len(os.Args) > 1 {
		toolName = os.Args[1]
	}

	// 创建输出目录
	os.MkdirAll(filepath.Join(workDir, "dist"), 0755)

	if toolName != "" {
		// 构建指定工具
		buildTool(workDir, toolName)
	} else {
		// 构建所有工具
		toolsDir := filepath.Join(workDir, "tools")
		entries, _ := os.ReadDir(toolsDir)
		for _, entry := range entries {
			if entry.IsDir() {
				buildTool(workDir, entry.Name())
			}
		}
	}

	fmt.Println()
	fmt.Println("构建完成!")
	fmt.Println("输出目录: dist/")
}

func buildTool(workDir, toolName string) {
	toolDir := filepath.Join(workDir, "tools", toolName)
	if _, err := os.Stat(toolDir); os.IsNotExist(err) {
		fmt.Printf("跳过 %s: 目录不存在\n", toolName)
		return
	}

	fmt.Printf("构建 %s.exe ...\n", toolName)

	cmd := exec.Command("go", "build",
		"-trimpath",
		"-ldflags", "-s -w -H windowsgui",
		"-o", filepath.Join(workDir, "dist", toolName+".exe"),
		".")
	cmd.Dir = toolDir
	cmd.Env = append(os.Environ(), "GOPROXY=https://goproxy.cn,direct")

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("失败: %s\n", output)
	} else {
		fmt.Println("成功")
	}
}
