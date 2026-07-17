package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	workDir string
	toolDir string
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run scripts/toolgen.go <工具名称> [描述]")
		fmt.Println("示例: go run scripts/toolgen.go data-cleaner \"数据清洗工具\"")
		return
	}

	toolName := os.Args[1]
	toolDesc := toolName
	if len(os.Args) > 2 {
		toolDesc = os.Args[2]
	}

	workDir, _ = os.Getwd()

	if _, err := os.Stat(filepath.Join(workDir, "go.work")); os.IsNotExist(err) {
		fmt.Println("错误: 请在 Workspace 根目录运行")
		return
	}

	toolDir = filepath.Join(workDir, "tools", toolName)
	if _, err := os.Stat(toolDir); !os.IsNotExist(err) {
		fmt.Printf("错误: 工具 %s 已存在\n", toolName)
		return
	}

	goVersion := parseGoVersion()

	fmt.Printf("创建工具: %s\n", toolName)

	os.MkdirAll(toolDir, 0755)

	// 生成最小 go.mod，依赖由 go mod tidy 补全
	goMod := fmt.Sprintf("module github.com/yourname/go2exes/tools/%s\n\ngo %s\n", toolName, goVersion)
	os.WriteFile(filepath.Join(toolDir, "go.mod"), []byte(goMod), 0644)

	// 生成 main.go 模板（纯 GUI，无 cobra）
	mainGo := fmt.Sprintf(`package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourname/go2exes/shared/winuser"
)

func showMsg(msg string, isError bool) {
	winuser.ShowMessage("%s", msg, isError)
}

func main() {
	// TODO: 实现功能
	exePath, err := os.Executable()
	if err != nil {
		showMsg(fmt.Sprintf("获取程序路径失败: %%v", err), true)
		return
	}
	dir := filepath.Dir(exePath)

	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	showMsg(fmt.Sprintf("目录: %%s\\n\\nTODO: 实现功能", dir), false)
}
`, toolDesc)
	os.WriteFile(filepath.Join(toolDir, "main.go"), []byte(mainGo), 0644)

	copyDefaultIcon()

	winresJson := fmt.Sprintf(`{
  "RT_GROUP_ICON": {
    "APP": {
      "0000": ["icon.png"]
    }
  },
  "RT_VERSION": {
    "#1": {
      "0000": {
        "fixed": {
          "file_version": "1.0.0.0",
          "product_version": "1.0.0.0"
        },
        "info": {
          "0804": {
            "CompanyName": "Your Company",
            "FileDescription": "%s",
            "FileVersion": "1.0.0.0",
            "InternalName": "%s",
            "LegalCopyright": "Copyright (C) 2026",
            "OriginalFilename": "%s.exe",
            "ProductName": "%s",
            "ProductVersion": "1.0.0.0"
          }
        }
      }
    }
  }
}`, toolDesc, toolName, toolName, toolDesc)
	winresJsonPath := filepath.Join(toolDir, "winres.json")
	os.WriteFile(winresJsonPath, []byte(winresJson), 0644)

	sysoCmd := exec.Command("go", "run", "github.com/tc-hib/go-winres@latest", "make",
		"--in", winresJsonPath,
		"--out", filepath.Join(toolDir, "rsrc"))
	sysoCmd.Dir = toolDir
	sysoCmd.Env = append(os.Environ(), "GOPROXY=https://goproxy.cn,direct")
	sysoCmd.Stdout = os.Stdout
	sysoCmd.Stderr = os.Stderr
	if err := sysoCmd.Run(); err != nil {
		fmt.Printf("生成资源文件失败: %v\n", err)
	}

	updateGoWork(toolName)

	fmt.Println("同步依赖...")
	syncCmd := exec.Command("go", "work", "sync")
	syncCmd.Dir = workDir
	syncCmd.Env = append(os.Environ(), "GOPROXY=https://goproxy.cn,direct")
	syncCmd.Run()

	fmt.Println()
	fmt.Println("创建成功!")
	fmt.Println()
	fmt.Printf("目录: tools/%s\n", toolName)
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Printf("  1. 编辑 tools/%s/main.go 添加功能\n", toolName)
	fmt.Println("  2. 运行 go run scripts/build.go 打包 exe")
}

// parseGoVersion 从 go.work 中提取 Go 版本号
func parseGoVersion() string {
	data, err := os.ReadFile(filepath.Join(workDir, "go.work"))
	if err != nil {
		return "1.22"
	}
	re := regexp.MustCompile(`go\s+(\d+\.\d+)`)
	if m := re.FindSubmatch(data); len(m) > 1 {
		return string(m[1])
	}
	return "1.22"
}

// copyDefaultIcon 复制默认图标到工具目录
func copyDefaultIcon() {
	src := filepath.Join(workDir, "shared", "assets", "default-icon.png")
	dst := filepath.Join(toolDir, "icon.png")

	if _, err := os.Stat(src); os.IsNotExist(err) {
		fmt.Printf("警告: 未找到默认图标 %s，跳过图标复制\n", src)
		return
	}

	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Printf("读取图标失败: %v\n", err)
		return
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		fmt.Printf("复制图标失败: %v\n", err)
	}
}

func updateGoWork(toolName string) {
	goWorkPath := filepath.Join(workDir, "go.work")
	data, _ := os.ReadFile(goWorkPath)
	content := string(data)

	useBlock := fmt.Sprintf("./tools/%s", toolName)
	if strings.Contains(content, useBlock) {
		return
	}

	newLine := fmt.Sprintf("\t./tools/%s\n", toolName)
	content = strings.Replace(content, ")", newLine+")", 1)
	os.WriteFile(goWorkPath, []byte(content), 0644)
}
