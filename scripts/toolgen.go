package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var workDir string

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

	// 获取工作区根目录
	workDir, _ = os.Getwd()

	// 检查是否在 workspace 根目录
	if _, err := os.Stat(filepath.Join(workDir, "go.work")); os.IsNotExist(err) {
		fmt.Println("错误: 请在 Workspace 根目录运行")
		return
	}

	// 检查工具是否已存在
	toolDir := filepath.Join(workDir, "tools", toolName)
	if _, err := os.Stat(toolDir); !os.IsNotExist(err) {
		fmt.Printf("错误: 工具 %s 已存在\n", toolName)
		return
	}

	fmt.Printf("创建工具: %s\n", toolName)

	// 创建目录
	os.MkdirAll(toolDir, 0755)
	os.MkdirAll(filepath.Join(workDir, "winres"), 0755)

	// 生成 go.mod
	goMod := fmt.Sprintf(`module github.com/yourname/go2exes/tools/%s

go 1.22

require (
	github.com/spf13/cobra v1.8.1
	github.com/xuri/excelize/v2 v2.9.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.4 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xuri/efp v0.0.0-20240408161823-9ad904a10d6d // indirect
	github.com/xuri/nfp v0.0.0-20240318013403-ab9948c2c4a7 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/text v0.19.0 // indirect
)
`, toolName)
	os.WriteFile(filepath.Join(toolDir, "go.mod"), []byte(goMod), 0644)

	// 生成 main.go
	mainGo := fmt.Sprintf(`package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/spf13/cobra"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

const (
	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
	MB_TOPMOST         = 0x00040000
)

var rootCmd = &cobra.Command{
	Use:   "%s",
	Short: "%s",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TODO: 实现功能")
		return nil
	},
}

func showMsg(msg string) {
	title, _ := syscall.UTF16PtrFromString("%s")
	text, _ := syscall.UTF16PtrFromString(msg)
	procMessageBox.Call(
		0,
		uintptr(unsafe.Pointer(text)),
		uintptr(unsafe.Pointer(title)),
		MB_OK|MB_TOPMOST|MB_ICONINFORMATION,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
`, toolName, toolDesc, toolDesc)
	os.WriteFile(filepath.Join(toolDir, "main.go"), []byte(mainGo), 0644)

	// 生成图标
	genIcon()

	// 生成资源配置
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
            "CompanyName": "Excel Tools",
            "FileDescription": "%s",
            "FileVersion": "1.0.0.0",
            "InternalName": "%s",
            "LegalCopyright": "Copyright (C) 2024",
            "OriginalFilename": "%s.exe",
            "ProductName": "%s",
            "ProductVersion": "1.0.0.0"
          }
        }
      }
    }
  }
}`, toolDesc, toolName, toolName, toolDesc)
	winresJsonPath := filepath.Join(workDir, "winres", "winres.json")
	os.WriteFile(winresJsonPath, []byte(winresJson), 0644)

	// 生成 syso 文件
	sysoCmd := exec.Command("go", "run", "github.com/tc-hib/go-winres@latest", "make",
		"--in", winresJsonPath,
		"--out", filepath.Join(toolDir, "rsrc"))
	sysoCmd.Dir = workDir
	sysoCmd.Env = append(os.Environ(), "GOPROXY=https://goproxy.cn,direct")
	sysoCmd.Stdout = os.Stdout
	sysoCmd.Stderr = os.Stderr
	if err := sysoCmd.Run(); err != nil {
		fmt.Printf("生成资源文件失败: %v\n", err)
	}

	// 更新 go.work
	updateGoWork(toolName)

	// 同步依赖
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

func genIcon() {
	iconPath := filepath.Join(workDir, "winres", "icon.png")

	iconCode := `package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			if x >= 4 && x < 28 && y >= 4 && y < 28 {
				img.Set(x, y, color.RGBA{G: 0xA6, R: 0x3B, B: 0x4C, A: 255})
			} else {
				img.Set(x, y, color.RGBA{G: 0x7D, R: 0x1A, B: 0x2D, A: 255})
			}
		}
	}
	f, _ := os.Create("` + iconPath + `")
	defer f.Close()
	png.Encode(f, img)
}
`

	tmpFile := filepath.Join(workDir, "gen_icon_temp.go")
	os.WriteFile(tmpFile, []byte(iconCode), 0644)
	defer os.Remove(tmpFile)

	cmd := exec.Command("go", "run", tmpFile)
	cmd.Dir = workDir
	cmd.Run()
}

func updateGoWork(toolName string) {
	goWorkPath := filepath.Join(workDir, "go.work")
	data, _ := os.ReadFile(goWorkPath)
	content := string(data)

	// 检查是否已存在
	useBlock := fmt.Sprintf("./tools/%s", toolName)
	if len(content) > 0 && contains(content, useBlock) {
		return
	}

	// 在 use 块中添加新工具
	newLine := fmt.Sprintf("\t./tools/%s\n", toolName)
	content = replace(content, ")", newLine+")")
	os.WriteFile(goWorkPath, []byte(content), 0644)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func replace(s, old, new string) string {
	if idx := indexOf(s, old); idx >= 0 {
		return s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
