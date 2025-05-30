@echo off
echo Building QRender...

:: 设置编译环境变量
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

:: 使用优化选项编译
go build -ldflags "-s -w" -o qrender.exe main.go

:: 检查编译结果
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b %ERRORLEVEL%
)

:: 显示原始文件大小
for %%A in (qrender.exe) do (
    echo Original binary size: %%~zA bytes
)

:: 使用 UPX 压缩
echo Compressing with UPX...
upx --best --lzma qrender.exe

:: 检查压缩结果
if %ERRORLEVEL% NEQ 0 (
    echo UPX compression failed! Make sure UPX is installed.
    echo You can download UPX from: https://github.com/upx/upx/releases
    exit /b %ERRORLEVEL%
)

:: 显示压缩后的文件大小
for %%A in (qrender.exe) do (
    echo Compressed binary size: %%~zA bytes
)

echo Done. 