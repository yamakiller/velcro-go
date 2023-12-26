@echo off 

::协议文件路径, 最后不要跟"\"符号
set SOURCE_FOLDER=.\proto

::编译器路径
set GO_COMPILER_PATH=..\..\..\tools\proto\bin\protoc.exe

::删除旧文件
del .\protocols\*.pb.go /f /s /q

::生成client_id.proto
echo %GO_COMPILER_PATH% --go_out=.\protocols --proto_path=%SOURCE_FOLDER% report_nat.proto require_nat.proto signin.proto signout.proto register_account.proto
%GO_COMPILER_PATH% --go_out=.\protocols --proto_path=%SOURCE_FOLDER% report_nat.proto require_nat.proto signin.proto signout.proto register_account.proto

::protoc -I="../actor" --go_out=. --go_opt=paths=source_relative --proto_path=. routercontracts.proto
::pubkey.proto forward_message.proto

echo complate 

pause 