# Windows 打包指南

本文用于在 Windows 环境下打包佳依拆书 Wails 桌面应用。目标产物有两类：

1. 免安装 zip 包：用户下载后解压，双击 exe 即可使用软件。
2. Windows 安装包：用户双击 installer exe，按安装向导安装软件。

## 常用打包命令

完成首次配置后，后续打包优先只看本节。

### 打免安装 zip 包

```powershell
Set-Location E:\workspace\html_workspace\jiayi-book-lab\init

$Version = "0.1.0"
$ExeName = "jiayi-book-lab.exe"
$PortableName = "jiayi-book-lab-$Version-windows-x64-portable"
$PortableDir = ".\dist\$PortableName"

wails build -clean -platform windows/amd64 -webview2 browser

下面是把EXE拷贝到 dist 目录下的Portable目录，压缩成zip文件，也可以使用其他软件
New-Item -ItemType Directory -Force $PortableDir | Out-Null
Copy-Item ".\build\bin\$ExeName" "$PortableDir\$ExeName" -Force
Compress-Archive -Path "$PortableDir\*" -DestinationPath ".\dist\$PortableName.zip" -Force
Get-ChildItem ".\dist\$PortableName.zip"
```

产物位置：

```text
init/dist/jiayi-book-lab-0.1.0-windows-x64-portable.zip
```

### 打 Windows 安装包

```powershell
Set-Location E:\workspace\html_workspace\jiayi-book-lab\init

wails build -clean -platform windows/amd64 -nsis -webview2 download
Get-ChildItem .\build\bin -Filter "*installer.exe"
```

默认产物位置：

```text
init/build/bin/jiayi-book-lab-amd64-installer.exe
```

### 同时打 zip 包和安装包

```powershell
Set-Location E:\workspace\html_workspace\jiayi-book-lab\init

$Version = "0.1.0"
$ExeName = "jiayi-book-lab.exe"
$PortableName = "jiayi-book-lab-$Version-windows-x64-portable"
$PortableDir = ".\dist\$PortableName"

wails build -clean -platform windows/amd64 -webview2 browser

New-Item -ItemType Directory -Force $PortableDir | Out-Null
Copy-Item ".\build\bin\$ExeName" "$PortableDir\$ExeName" -Force
Compress-Archive -Path "$PortableDir\*" -DestinationPath ".\dist\$PortableName.zip" -Force

wails build -clean -platform windows/amd64 -nsis -webview2 download

Get-ChildItem ".\dist\$PortableName.zip"
Get-ChildItem .\build\bin -Filter "*installer.exe"
```

## 首次打包前配置

首次正式打包前，先完成本节配置。后续版本只需要回到顶部执行常用命令。

### 检查基础环境

在 Windows PowerShell 中执行：

```powershell
go version
node --version
npm --version
wails version
wails doctor
```

如果 `wails` 命令不存在，按 `docs/dev-guides/project-initialization.md` 中的 Wails CLI 安装步骤处理。

### 安装 NSIS

只有打 Windows 安装包时需要 NSIS。免安装 zip 包不需要 NSIS。

推荐用 Winget 安装：

```powershell
winget install NSIS.NSIS --silent
```

安装后重新打开 PowerShell，再确认：

```powershell
makensis /VERSION
```

如果提示 `makensis` 不存在，说明 NSIS 的 `Bin` 目录没有加入 `Path`。手动安装 NSIS 时，需要把类似下面的目录加入用户级或系统级 `Path`：

```text
C:\Program Files (x86)\NSIS\Bin
```

### 配置应用名称和版本信息

打包前确认 `init/wails.json` 不再使用模板默认的 `init` 作为输出文件名。建议配置为：

```json
{
  "name": "jiayi-book-lab",
  "outputfilename": "jiayi-book-lab",
  "info": {
    "companyName": "Jiayi Book Lab",
    "productName": "佳依拆书",
    "productVersion": "0.1.0",
    "copyright": "Copyright 2026 Jiayi Book Lab",
    "comments": "本地优先的小说拆解与写作学习工具"
  }
}
```

说明：

1. `outputfilename` 决定普通构建生成的 exe 文件名。
2. `info.productName` 会影响安装向导、桌面快捷方式和 Windows 程序列表中的显示名。
3. `info.productVersion` 每次正式发版都应递增，建议使用 `0.1.0` 这种三段式版本号。
4. 修改 `wails.json` 后，顶部命令中的 `$Version` 也要同步修改。

### 配置应用图标

Wails Windows 打包会使用 `init/build/windows/icon.ico`。如果只替换 `init/build/appicon.png`，建议删除旧的 `icon.ico` 后重新打包，让 Wails 重新生成：

```powershell
Set-Location E:\workspace\html_workspace\jiayi-book-lab\init

Remove-Item -LiteralPath .\build\windows\icon.ico -Force
wails build -clean -platform windows/amd64
```

删除前只删除明确的 `init/build/windows/icon.ico`，不要批量删除 `build/windows/installer/`，因为里面包含安装包模板。

## 产物说明

### 免安装 zip 包

免安装 zip 包只放最终 exe 和必要说明文件，不放源码、不放 `node_modules`、不放前端 `dist`。

当前 Wails Windows 构建会把前端资源嵌入 exe，普通发布不需要把 `frontend/dist` 一起发给用户。

免安装包适合：

1. 内测分发。
2. 用户不希望安装软件。
3. 放在 U 盘或任意目录中运行。

注意：Wails Windows 应用依赖 Microsoft Edge WebView2 Runtime。Windows 10/11 通常已经具备该运行时。免安装 zip 包不负责安装系统运行时，所以顶部命令使用 `-webview2 browser`，避免把安装行为混进便携包。

### Windows 安装包

安装包由 Wails 的 `-nsis` 参数生成，输出到 `init/build/bin/`。

安装包适合：

1. 给普通用户正式分发。
2. 创建桌面快捷方式和开始菜单快捷方式。
3. 在卸载列表中显示应用。
4. 在 WebView2 Runtime 缺失时通过安装流程补齐运行时。

如果希望安装到当前用户目录，可以尝试：

```powershell
wails build -clean -platform windows/amd64 -nsis -installscope user -webview2 download
```

如果希望安装到系统 Program Files 目录，使用默认 machine scope：

```powershell
wails build -clean -platform windows/amd64 -nsis -installscope machine -webview2 download
```

machine scope 可能触发管理员权限确认，这是 Windows 安装到系统目录的正常行为。

## 发布前检查

每次正式发包前至少检查：

1. `init/wails.json` 的 `info.productVersion` 与 zip 文件名中的 `$Version` 一致。
2. `wails build` 没有前端类型错误、Go 编译错误或资源嵌入错误。
3. 在 `init/dist/*portable/` 中双击 exe 可以启动。
4. 解压 zip 到一个新目录后，双击 exe 可以启动。
5. 双击 installer exe 可以完成安装，桌面快捷方式可以启动。
6. 卸载安装包后，不应删除用户自己选择的小说项目目录。
7. 新建、导入、阅读、选文、拆解、素材沉淀相关正式数据仍保存到用户选择的小说项目目录，不写进安装目录。

## 常见问题

### 打包后还是生成 `init.exe`

检查 `init/wails.json`：

```json
{
  "outputfilename": "jiayi-book-lab"
}
```

修改后重新执行：

```powershell
wails build -clean -platform windows/amd64
```

### `makensis` 命令不存在

说明 NSIS 未安装，或 NSIS 的 `Bin` 目录没有加入 `Path`。重新安装 NSIS 或把下面目录加入 `Path` 后，重新打开 PowerShell：

```text
C:\Program Files (x86)\NSIS\Bin
```

### zip 包用户打不开

先确认用户系统是 Windows 10/11 x64，并确认 Microsoft Edge WebView2 Runtime 是否存在。免安装包不做系统组件安装；如果用户机器缺少 WebView2 Runtime，优先让用户使用 installer 版本。

### 安装包名称或快捷方式名称不对

检查 `init/wails.json` 的 `name` 和 `info.productName`。NSIS 安装包模板会使用这些信息生成安装包文件名、安装向导名称和快捷方式名称。

### 不要提交打包产物

`init/build/bin/`、`init/frontend/dist/` 和 `init/dist/` 都是可重新生成的打包产物，不应作为源码提交。发布时把最终 zip 或 installer exe 放到发布渠道即可。

## 参考资料

1. Wails 构建文档：`https://wails.io/docs/gettingstarted/building/`
2. Wails Windows NSIS 安装包文档：`https://wails.io/docs/guides/windows-installer/`
3. Wails 项目配置文档：`https://wails.io/docs/reference/project-config/`
