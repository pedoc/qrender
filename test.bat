@echo off
set VERSION=1.5.0
set COUNT=7
set STATUS=running
set PATH=/usr/local/bin
set FILENAME=test.txt
set DEBUG=true
set LOG_LEVEL=debug
set MULTILINE=true
set EMPTY_VAR=
set SPECIAL_CHARS=!@#$%^&*()

echo 运行测试（不显示环境变量）...
qrender.exe -template example.txt -output result.txt

echo.
echo 运行测试（显示环境变量）...
qrender.exe -template example.txt -output result_verbose.txt -verbose

echo 测试完成，请查看 result.txt 和 result_verbose.txt 文件 