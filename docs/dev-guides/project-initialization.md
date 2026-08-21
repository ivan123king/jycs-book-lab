# 工程创建指南

本文用于从零创建佳依拆书的本地 PC 端工程骨架。当前阶段采用 Wails v2 稳定版、Vue 3 和 TypeScript，目标创建目录固定为仓库根目录下的 `init/`。

## 适用范围

当需要首次创建应用工程、复现工程初始化过程、检查初始化命令是否符合项目方向时，先阅读本文。

本文只覆盖工程骨架创建和首次运行验证，不包含业务功能实现。业务代码开发仍需继续阅读：

1. 页面、交互、组件：`docs/dev-guides/frontend-pages.md`
2. 后端、本地数据、业务服务：`docs/dev-guides/backend-local-data.md`
3. 注释、文件组织、协作规则：`docs/dev-guides/coding-rules.md`

## 初始化原则

1. 工程目录必须创建为 `init/`，不要直接把 Wails 文件散放到仓库根目录。
2. 不要在 `init/` 内再次执行 `git init`，当前仓库根目录已经由 Git 管理。
3. 第一阶段只创建本地单机应用，不引入账号、云同步、远程服务或多人协作依赖。
4. 前端使用 Vue 3 + TypeScript，后端使用 Wails 生成的 Go 后端。
5. 正式业务数据后续必须保存到用户选择的小说项目目录中，不要默认写死在应用安装目录或工程目录里。
6. SQLite 即使后续引入，也只能作为可重建缓存，不能作为正式数据唯一来源。

## 环境要求

在 Windows PowerShell 中执行下面命令，确认基础工具可用：

```powershell
go version
node --version
npm --version
git --version
```

建议版本：

1. Go：遵循 Wails v2 当前要求，至少 Go 1.21。
2. Node.js：使用当前 LTS 版本，确保自带 npm。
3. Git：使用稳定版本即可。
4. Windows WebView2 Runtime：Windows 10/11 通常已安装；如果 `wails doctor` 提示缺失，再按提示安装。

如果 `go`、`node`、`npm` 或 `git` 命令不存在，先安装对应工具，关闭并重新打开 PowerShell 后再继续。

## 国内网络配置

如果在国内网络环境下执行 `go install`、`go mod tidy` 或 `npm install` 长时间无响应，先配置国内可用的模块代理。

推荐先使用当前 PowerShell 会话级配置。这种方式只影响当前终端窗口，不会修改用户级 Go 配置，也不会影响以后其他工程：

```powershell
$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOSUMDB = "sum.golang.google.cn"
```

确认当前终端内配置是否生效：

```powershell
Write-Output $env:GOPROXY
Write-Output $env:GOSUMDB
```

随后在同一个 PowerShell 窗口里继续执行：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

如果关闭这个 PowerShell 窗口，上面的 Go 代理配置就会失效。以后其他工程默认仍使用原始配置。

`npm install` 也可以使用单次命令级镜像，不修改全局 npm 配置：

```powershell
npm install --registry=https://registry.npmmirror.com
```

如果只想为本工程前端目录持久配置 npm 镜像，可以在创建 `init/` 后进入 `init/frontend/`，再执行：

```powershell
Set-Location E:\workspace\html_workspace\jiayi-book-lab\init\frontend
npm config set registry https://registry.npmmirror.com --location=project
npm config get registry
```

这会在当前前端工程目录写入 `.npmrc`，只影响该 npm 项目。

如果你明确希望所有 Go 工程长期使用国内代理，才使用用户级持久配置：

```powershell
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

如果你明确希望所有 npm 工程长期使用国内镜像，才使用用户级持久配置：

```powershell
npm config set registry https://registry.npmmirror.com
```

## 安装 Wails CLI

本项目当前优先使用 Wails v2 稳定版。Wails v3 仍处于快速变化阶段，除非后续技术选型文档明确调整，否则不要用 `wails3` 创建正式工程。

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

安装后确认 `wails` 是否可用：

```powershell
wails version
```

如果提示找不到 `wails`，通常是 Go 的 bin 目录没有加入当前终端的 `PATH`。可以先在当前 PowerShell 会话临时加入：

```powershell
$env:Path = "$(go env GOPATH)\bin;$env:Path"
wails version
```

如果临时加入后可用，再把 `$(go env GOPATH)\bin` 加入用户级环境变量 `Path`，然后重新打开 PowerShell。

执行 Wails 环境检查：

```powershell
wails doctor
```

只有当 `wails doctor` 没有阻断性错误后，再创建工程。

## 创建 init 工程目录

从仓库根目录执行：

```powershell
Set-Location E:\workspace\html_workspace\jiayi-book-lab
```

确认当前目录正确：

```powershell
Get-Location
Get-ChildItem
```

确认 `init/` 尚不存在：

```powershell
Test-Path .\init
```

如果输出 `False`，可以继续创建。  
如果输出 `True`，说明目录已经存在，先人工确认里面是否有需要保留的内容，不要直接覆盖或删除。

创建 Wails + Vue 3 + TypeScript 工程：

```powershell
wails init -n jiayi-book-lab -d init -t vue-ts
```

参数含义：

1. `-n jiayi-book-lab`：应用工程名称。
2. `-d init`：把工程文件创建到仓库根目录下的 `init/`。
3. `-t vue-ts`：使用 Vue + TypeScript 模板。

不要添加 `-g` 参数，避免在 `init/` 里生成嵌套 Git 仓库。

## 安装前端依赖

进入新创建的工程目录：

```powershell
Set-Location .\init
```

整理 Go 依赖：

```powershell
go mod tidy
```

进入前端目录并安装 npm 依赖：

```powershell
Set-Location .\frontend
npm install
```

回到 Wails 工程根目录：

```powershell
Set-Location ..
```

## 首次运行验证

启动开发模式：

```powershell
wails dev
```

预期结果：

1. Wails 会启动 Go 后端和 Vite 前端开发服务。
2. 桌面窗口应能打开。
3. 修改前端代码后应能热更新。
4. PowerShell 中不应出现阻断性编译错误。

如果 `wails dev` 失败，优先执行：

```powershell
wails doctor
go mod tidy
Set-Location .\frontend
npm install
Set-Location ..
wails dev
```

## 首次构建验证

开发模式能运行后，再执行构建：

```powershell
wails build
```

构建成功后，检查产物目录：

```powershell
Get-ChildItem .\build\bin
```

Windows 下通常会生成可执行文件。实际文件名以 `wails.json` 和 Wails 输出为准。

## 初始化后的目录结构

创建完成后，`init/` 内大致应包含：

```text
init/
├── app.go
├── main.go
├── go.mod
├── go.sum
├── wails.json
├── build/
└── frontend/
    ├── package.json
    ├── index.html
    └── src/
```

后续业务代码建议按职责逐步整理，而不是在模板文件中不断堆叠逻辑。

建议 Go 后端后续演进方向：

```text
init/
└── internal/
    ├── bookproject/
    ├── importbook/
    ├── deconstruction/
    ├── material/
    └── appsettings/
```

建议 Vue 前端后续演进方向：

```text
init/
└── frontend/
    └── src/
        ├── pages/
        ├── components/
        ├── modules/
        ├── services/
        ├── stores/
        └── types/
```

这些目录不要求初始化时一次性全部创建。只有开始开发对应业务模块时，再按模块边界创建，避免空目录和过早抽象。

## 初始化后必须保留的产品边界

工程创建完成后，后续代码落地必须继续遵守：

1. 导入小说、阅读原文、选中文字、创建拆解、沉淀素材是第一阶段核心闭环。
2. 原文、拆解卡片、素材必须分离建模。
3. 素材必须能追溯到书籍、章节和原文位置。
4. 一个原文片段必须允许创建多个拆解。
5. 每本书拥有独立拆解方案，不把拆解字段写死为固定模板。
6. 小说正文和可直接阅读的长文本保存为 TXT。
7. 结构化业务数据保存为 JSON。
8. 单本小说正式数据保存到用户选择的小说项目目录。
9. 全局主题、窗口偏好、最近打开项目等才保存到应用系统数据目录。
10. 不主动扩张跨书比较、复杂可视化、AI 自动拆解和 AI 自动总结。

## 常见问题

### `wails` 命令不存在

先确认 Go 已安装：

```powershell
go version
go env GOPATH
```

再把 Go bin 目录加入当前 PowerShell 会话：

```powershell
$env:Path = "$(go env GOPATH)\bin;$env:Path"
wails version
```

### `go install` 长时间无响应

国内网络环境可能无法稳定访问默认 Go 模块源。先停止当前命令，再在当前 PowerShell 窗口配置临时 Go 模块代理：

```powershell
$env:GOPROXY = "https://goproxy.cn,direct"
$env:GOSUMDB = "sum.golang.google.cn"
Write-Output $env:GOPROXY
Write-Output $env:GOSUMDB
```

重新执行：

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

这种写法只影响当前 PowerShell 窗口，不会影响以后其他工程。

### `npm install` 失败

先确认 Node 和 npm：

```powershell
node --version
npm --version
```

如果版本过旧，安装当前 Node.js LTS 后重新打开 PowerShell，再回到 `init/frontend` 执行：

```powershell
npm install
```

### `wails doctor` 提示 WebView2 缺失

按 `wails doctor` 的提示安装 Microsoft Edge WebView2 Runtime。安装完成后重新打开 PowerShell，再执行：

```powershell
wails doctor
```

### 误在 `init/` 内创建了嵌套 Git 仓库

如果创建工程时误加了 `-g`，`init/.git/` 可能会出现嵌套仓库。先确认没有需要单独保留的提交历史，再人工处理。不要在未确认前批量删除。

## 参考资料

1. Wails v2 安装文档：`https://wails.io/zh-Hans/docs/gettingstarted/installation/`
2. Wails v2 创建项目文档：`https://wails.io/zh-Hans/docs/gettingstarted/firstproject/`
3. Wails v2 CLI 参数说明：`https://wails.io/zh-Hans/docs/reference/cli/`
