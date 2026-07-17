# go2exes

Go Workspace 多工具开发平台 —— 每个工具编译为独立 Windows exe，双击即用。

## 特性

- **一键生成工具脚手架** — `go run ./scripts/toolgen.go <name>` 自动创建目录、go.mod、图标、Windows 资源文件
- **一键打包所有 exe** — `go run ./scripts/build.go` 批量构建，输出到 `dist/`
- **共享库复用** — `shared/` 中的公共模块被所有工具共用
- **独立发布** — 每个工具是独立 Go 模块，可单独构建、单独分发
- **Windows 原生体验** — 编译时隐藏控制台窗口，带自定义图标和版本信息

## 目录结构

```
go2exes/
├── go.work                    # 工作区定义
├── shared/                    # 共享库
│   ├── winuser/               # Windows API 封装（MessageBox 等）
│   └── assets/                # 资源文件（默认图标等）
├── tools/                     # 各工具（每个 = 一个 exe）
│   └── empty-row-remover/     # 示例：Excel空行去除工具
├── scripts/
│   ├── toolgen.go             # 工具脚手架生成器
│   └── build.go               # 批量构建脚本
└── dist/                      # 构建输出
```

## 快速开始

### 环境要求

- Go 1.22+
- Windows（目标平台）

### 创建新工具

```bash
go run ./scripts/toolgen.go my-tool "我的工具"
```

自动完成：

1. 创建 `tools/my-tool/` 目录及 `go.mod`、`main.go`
2. 复制默认图标（从 `shared/assets/default-icon.png`）
3. 生成 Windows 版本资源（.syso）
4. 更新 `go.work` 注册新模块
5. 同步依赖

### 开发

编辑 `tools/my-tool/main.go`，添加你的逻辑。共享库可直接导入：

```go
import "github.com/yourname/go2exes/shared/winuser"

winuser.ShowInfo("标题", "内容")
winuser.ShowError("标题", "错误信息")
```

### 构建

```bash
# 构建所有工具
go run ./scripts/build.go

# 构建指定工具
go run ./scripts/build.go my-tool
```

输出到 `dist/my-tool.exe`，双击即可运行。

## 已有工具

| 工具                | 说明                        |
| ------------------- | --------------------------- |
| `empty-row-remover` | 批量去除 Excel 文件中的空行 |

## 注意事项

- `shared/` 不能 import 任何 `tools/` 下的模块（避免循环依赖）
- `tools/` 之间不要互相 import（保持独立）
- 分发时只需复制 `dist/*.exe`，无需 Go 环境
- 默认图标放在 `shared/assets/default-icon.png`，新工具自动复制
