@echo off 
set GO_COMPILER_PATH=tools\thrift\thrift.exe


%GO_COMPILER_PATH% -r --gen go .\utils\thrift\plugin\protocol.thrift

echo complate 
pause 