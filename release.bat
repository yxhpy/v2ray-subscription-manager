@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM V2Ray Subscription Manager Windows自动化发布脚本
REM 使用方法: release.bat <version> [commit_message]
REM 示例: release.bat v1.2.0 "修复Windows兼容性问题"

REM 检查参数
if "%~1"=="" (
    echo ❌ 使用方法: %0 ^<version^> [commit_message]
    echo ℹ️  示例: %0 v1.2.0 "修复Windows兼容性问题"
    exit /b 1
)

set VERSION=%~1
set COMMIT_MSG=%~2
if "%COMMIT_MSG%"=="" set COMMIT_MSG=Release %VERSION%

REM 验证版本格式
echo %VERSION% | findstr /r "^v[0-9]*\.[0-9]*\.[0-9]*$" >nul
if errorlevel 1 (
    echo ❌ 版本格式错误，应该是 vX.Y.Z 格式，如 v1.2.0
    exit /b 1
)

echo.
echo 🚀 开始自动化发布流程 - 版本: %VERSION%

REM 1. 检查工作目录状态
echo.
echo 🚀 检查Git工作目录状态
git status --porcelain >nul 2>&1
if not errorlevel 1 (
    for /f %%i in ('git status --porcelain') do (
        echo ⚠️  工作目录有未提交的更改，将自动提交
        git add .
        git status --short
        goto :continue_check
    )
)
echo ✅ 工作目录干净
:continue_check

REM 2. 编译所有平台版本
echo.
echo 🚀 编译所有平台版本
if not exist bin mkdir bin

echo ℹ️  开始编译各平台版本...

REM Windows amd64
echo ℹ️  编译 windows/amd64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o "bin\v2ray-subscription-manager-windows-amd64.exe" .
if exist "bin\v2ray-subscription-manager-windows-amd64.exe" (
    echo ✅ 编译完成: bin\v2ray-subscription-manager-windows-amd64.exe
) else (
    echo ❌ 编译失败: bin\v2ray-subscription-manager-windows-amd64.exe
    exit /b 1
)

REM Windows arm64
echo ℹ️  编译 windows/arm64...
set GOOS=windows
set GOARCH=arm64
go build -ldflags="-s -w" -o "bin\v2ray-subscription-manager-windows-arm64.exe" .
if exist "bin\v2ray-subscription-manager-windows-arm64.exe" (
    echo ✅ 编译完成: bin\v2ray-subscription-manager-windows-arm64.exe
) else (
    echo ❌ 编译失败: bin\v2ray-subscription-manager-windows-arm64.exe
    exit /b 1
)

REM Linux amd64
echo ℹ️  编译 linux/amd64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o "bin\v2ray-subscription-manager-linux-amd64" .
if exist "bin\v2ray-subscription-manager-linux-amd64" (
    echo ✅ 编译完成: bin\v2ray-subscription-manager-linux-amd64
) else (
    echo ❌ 编译失败: bin\v2ray-subscription-manager-linux-amd64
    exit /b 1
)

REM Linux arm64
echo ℹ️  编译 linux/arm64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags="-s -w" -o "bin\v2ray-subscription-manager-linux-arm64" .
if exist "bin\v2ray-subscription-manager-linux-arm64" (
    echo ✅ 编译完成: bin\v2ray-subscription-manager-linux-arm64
) else (
    echo ❌ 编译失败: bin\v2ray-subscription-manager-linux-arm64
    exit /b 1
)

REM macOS amd64
echo ℹ️  编译 darwin/amd64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o "bin\v2ray-subscription-manager-darwin-amd64" .
if exist "bin\v2ray-subscription-manager-darwin-amd64" (
    echo ✅ 编译完成: bin\v2ray-subscription-manager-darwin-amd64
) else (
    echo ❌ 编译失败: bin\v2ray-subscription-manager-darwin-amd64
    exit /b 1
)

REM macOS arm64
echo ℹ️  编译 darwin/arm64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags="-s -w" -o "bin\v2ray-subscription-manager-darwin-arm64" .
if exist "bin\v2ray-subscription-manager-darwin-arm64" (
    echo ✅ 编译完成: bin\v2ray-subscription-manager-darwin-arm64
) else (
    echo ❌ 编译失败: bin\v2ray-subscription-manager-darwin-arm64
    exit /b 1
)

REM 3. 创建压缩包
echo.
echo 🚀 创建发布压缩包
cd bin

REM 删除旧版本压缩包
del /q *%VERSION%*.zip 2>nul
del /q *%VERSION%*.tar.gz 2>nul

REM 创建各平台压缩包
echo ℹ️  创建压缩包: v2ray-subscription-manager-%VERSION%-windows-amd64.zip
powershell -command "Compress-Archive -Path 'v2ray-subscription-manager-windows-amd64.exe' -DestinationPath 'v2ray-subscription-manager-%VERSION%-windows-amd64.zip' -Force"
echo ✅ 压缩包创建完成: v2ray-subscription-manager-%VERSION%-windows-amd64.zip

echo ℹ️  创建压缩包: v2ray-subscription-manager-%VERSION%-windows-arm64.zip
powershell -command "Compress-Archive -Path 'v2ray-subscription-manager-windows-arm64.exe' -DestinationPath 'v2ray-subscription-manager-%VERSION%-windows-arm64.zip' -Force"
echo ✅ 压缩包创建完成: v2ray-subscription-manager-%VERSION%-windows-arm64.zip

echo ℹ️  创建压缩包: v2ray-subscription-manager-%VERSION%-linux-amd64.zip
powershell -command "Compress-Archive -Path 'v2ray-subscription-manager-linux-amd64' -DestinationPath 'v2ray-subscription-manager-%VERSION%-linux-amd64.zip' -Force"
echo ✅ 压缩包创建完成: v2ray-subscription-manager-%VERSION%-linux-amd64.zip

echo ℹ️  创建压缩包: v2ray-subscription-manager-%VERSION%-linux-arm64.zip
powershell -command "Compress-Archive -Path 'v2ray-subscription-manager-linux-arm64' -DestinationPath 'v2ray-subscription-manager-%VERSION%-linux-arm64.zip' -Force"
echo ✅ 压缩包创建完成: v2ray-subscription-manager-%VERSION%-linux-arm64.zip

echo ℹ️  创建压缩包: v2ray-subscription-manager-%VERSION%-darwin-amd64.zip
powershell -command "Compress-Archive -Path 'v2ray-subscription-manager-darwin-amd64' -DestinationPath 'v2ray-subscription-manager-%VERSION%-darwin-amd64.zip' -Force"
echo ✅ 压缩包创建完成: v2ray-subscription-manager-%VERSION%-darwin-amd64.zip

echo ℹ️  创建压缩包: v2ray-subscription-manager-%VERSION%-darwin-arm64.zip
powershell -command "Compress-Archive -Path 'v2ray-subscription-manager-darwin-arm64' -DestinationPath 'v2ray-subscription-manager-%VERSION%-darwin-arm64.zip' -Force"
echo ✅ 压缩包创建完成: v2ray-subscription-manager-%VERSION%-darwin-arm64.zip

REM 创建全平台压缩包
echo ℹ️  创建全平台压缩包: v2ray-subscription-manager-%VERSION%-all-platforms.zip
powershell -command "Compress-Archive -Path 'v2ray-subscription-manager-*' -DestinationPath 'v2ray-subscription-manager-%VERSION%-all-platforms.zip' -Force"
echo ✅ 全平台压缩包创建完成: v2ray-subscription-manager-%VERSION%-all-platforms.zip

cd ..

REM 4. 生成校验和
echo.
echo 🚀 生成文件校验和
cd bin
powershell -command "Get-ChildItem *%VERSION%* | Get-FileHash -Algorithm SHA256 | ForEach-Object { $_.Hash + '  ' + $_.Path.Name } | Out-File -Encoding UTF8 v2ray-subscription-manager-%VERSION%-checksums.txt"
echo ✅ 校验和文件生成完成: v2ray-subscription-manager-%VERSION%-checksums.txt
cd ..

REM 5. 提交代码
echo.
echo 🚀 提交代码到Git仓库
git add .

REM 生成详细的提交信息
echo release: %COMMIT_MSG% > commit_message.tmp
echo. >> commit_message.tmp
echo Version: %VERSION% >> commit_message.tmp
echo Build Date: %date% %time% >> commit_message.tmp
echo Platforms: Windows, Linux, macOS (amd64, arm64) >> commit_message.tmp
echo. >> commit_message.tmp
echo Changes in this release: >> commit_message.tmp
echo - 自动化构建和发布流程 >> commit_message.tmp
echo - 跨平台二进制文件 >> commit_message.tmp
echo - 完整的文件校验和 >> commit_message.tmp
echo - 优化的构建参数 (-ldflags="-s -w") >> commit_message.tmp

git commit -F commit_message.tmp
del commit_message.tmp
echo ✅ 代码提交完成

REM 6. 推送到远程仓库
echo.
echo 🚀 推送代码到远程仓库
git push origin main
echo ✅ 代码推送完成

REM 7. 创建并推送标签
echo.
echo 🚀 创建Git标签

REM 检查标签是否已存在
git tag -l | findstr /x "%VERSION%" >nul
if not errorlevel 1 (
    echo ⚠️  标签 %VERSION% 已存在，删除旧标签
    git tag -d "%VERSION%"
    git push origin ":refs/tags/%VERSION%" 2>nul
)

REM 创建新标签
git tag -a "%VERSION%" -m "Release %VERSION%

Build Date: %date% %time%
Commit: %COMMIT_MSG%"

git push origin "%VERSION%"
echo ✅ 标签 %VERSION% 创建并推送完成

REM 8. 生成Release说明
echo.
echo 🚀 生成GitHub Release说明
(
echo # %VERSION% - %COMMIT_MSG%
echo.
echo ## 📦 下载文件
echo.
echo ^| 平台 ^| 架构 ^| 文件名 ^| 说明 ^|
echo ^|------^|------^|--------^|------^|
echo ^| **Windows** ^| x64 ^| `v2ray-subscription-manager-%VERSION%-windows-amd64.zip` ^| Windows 64位版本 ^|
echo ^| **Windows** ^| ARM64 ^| `v2ray-subscription-manager-%VERSION%-windows-arm64.zip` ^| Windows ARM64版本 ^|
echo ^| **Linux** ^| x64 ^| `v2ray-subscription-manager-%VERSION%-linux-amd64.zip` ^| Linux 64位版本 ^|
echo ^| **Linux** ^| ARM64 ^| `v2ray-subscription-manager-%VERSION%-linux-arm64.zip` ^| Linux ARM64版本 ^|
echo ^| **macOS** ^| Intel ^| `v2ray-subscription-manager-%VERSION%-darwin-amd64.zip` ^| macOS Intel版本 ^|
echo ^| **macOS** ^| Apple Silicon ^| `v2ray-subscription-manager-%VERSION%-darwin-arm64.zip` ^| macOS M1/M2版本 ^|
echo ^| **All Platforms** ^| 通用 ^| `v2ray-subscription-manager-%VERSION%-all-platforms.zip` ^| 所有平台打包 ^|
echo ^| **Checksums** ^| - ^| `v2ray-subscription-manager-%VERSION%-checksums.txt` ^| SHA256校验和 ^|
echo.
echo ## 🔧 安装说明
echo.
echo ### Windows
echo ```bash
echo # 下载并解压
echo unzip v2ray-subscription-manager-%VERSION%-windows-amd64.zip
echo # 直接运行
echo v2ray-subscription-manager-windows-amd64.exe --help
echo ```
echo.
echo ### Linux/macOS
echo ```bash
echo # 下载并解压
echo unzip v2ray-subscription-manager-%VERSION%-linux-amd64.zip  # Linux
echo unzip v2ray-subscription-manager-%VERSION%-darwin-amd64.zip # macOS
echo.
echo # 添加执行权限
echo chmod +x v2ray-subscription-manager-*
echo.
echo # 运行
echo ./v2ray-subscription-manager-linux-amd64 --help
echo ```
echo.
echo ## 🔍 文件验证
echo.
echo 使用SHA256校验和验证文件完整性：
echo ```bash
echo # 下载校验和文件
echo wget https://github.com/yxhpy/v2ray-subscription-manager/releases/download/%VERSION%/v2ray-subscription-manager-%VERSION%-checksums.txt
echo.
echo # 验证文件
echo sha256sum -c v2ray-subscription-manager-%VERSION%-checksums.txt
echo ```
echo.
echo ## 📊 构建信息
echo.
echo - **构建时间**: %date% %time%
echo - **构建参数**: `-ldflags="-s -w"` ^(优化大小^)
echo.
echo ## 🚀 使用示例
echo.
echo ```bash
echo # 测试订阅链接
echo ./v2ray-subscription-manager parse https://your-subscription-url
echo.
echo # 启动代理
echo ./v2ray-subscription-manager start-proxy random https://your-subscription-url
echo.
echo # 批量测速
echo ./v2ray-subscription-manager speed-test https://your-subscription-url
echo ```
echo.
echo **完整文档**: [README.md]^(https://github.com/yxhpy/v2ray-subscription-manager/blob/main/README.md^)
echo **更新日志**: [RELEASE_NOTES.md]^(https://github.com/yxhpy/v2ray-subscription-manager/blob/main/RELEASE_NOTES.md^)
) > "release_notes_%VERSION%.md"

echo ✅ Release说明生成完成: release_notes_%VERSION%.md

REM 9. 显示发布摘要
echo.
echo 🚀 发布摘要
echo.
echo 🎉 发布完成！
echo.
echo 📊 发布信息:
echo   版本: %VERSION%
echo   时间: %date% %time%
echo.
echo 📦 生成的文件:
cd bin
for %%f in (*%VERSION%*) do echo   ✓ %%f
cd ..
echo.
echo 🔗 下一步操作:
echo   1. 访问: https://github.com/yxhpy/v2ray-subscription-manager/releases
echo   2. 点击 'Create a new release'
echo   3. 选择标签: %VERSION%
echo   4. 复制 release_notes_%VERSION%.md 的内容作为描述
echo   5. 上传 bin\ 目录下的所有 *%VERSION%* 文件
echo   6. 点击 'Publish release'
echo.
echo ✨ 自动化发布流程执行完成！

endlocal 