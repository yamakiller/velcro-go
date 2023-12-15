@echo off 

::协议文件路径, 最后不要跟"\"符号
set SOURCE_FOLDER=.\cluster\gateway\proto

::编译器路径
set GO_COMPILER_PATH=tools\proto\bin\protoc.exe
set GO_TARGET_PATH=.\cluster\gateway\protocols

::删除旧文件

del %GO_TARGET_PATH%\*.pb.go /f /s /q

::遍历所有文件
for /f "delims=" %%i in ('dir /b "%SOURCE_FOLDER%\*.proto"') do (
    ::生成 golang 代码
    echo %GO_COMPILER_PATH% --go_out=%GO_TARGET_PATH% %SOURCE_FOLDER%\%%i
    %GO_COMPILER_PATH% --go_out=%GO_TARGET_PATH% %SOURCE_FOLDER%\%%i 
)

echo complate 

pause 