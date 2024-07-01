@echo off 

::协议文件路径, 最后不要跟"\"符号
::set SOURCE_FOLDER=.\proto

::编译器路径
set GO_COMPILER_PATH=..\..\tools\proto\bin\protoc.exe

::删除旧文件
del .\protocols\prvs\*.pb.go /f /s /q
del .\protocols\pubs\*.pb.go /f /s /q
del .\protocols\rdsstruct\*.pb.go /f /s /q

::生成client_id.proto
echo %GO_COMPILER_PATH% --go_out=.\protocols\prvs --proto_path=.\proto\prvs report_nat.proto exitbattle.proto 
%GO_COMPILER_PATH% --go_out=.\protocols\prvs --proto_path=.\proto\prvs report_nat.proto exitbattle.proto 

echo %GO_COMPILER_PATH% --go_out=.\protocols\pubs --proto_path=.\proto\pubs report_nat_client.proto signin.proto signout.proto battle.proto
%GO_COMPILER_PATH% --go_out=.\protocols\pubs --proto_path=.\proto\pubs report_nat_client.proto signin.proto signout.proto battle.proto

echo %GO_COMPILER_PATH% --go_out=.\protocols\rdsstruct --proto_path=.\proto\rdsstruct rdslogin.proto rdsbattle.proto
%GO_COMPILER_PATH% --go_out=.\protocols\rdsstruct --proto_path=.\proto\rdsstruct rdslogin.proto rdsbattle.proto

::protoc -I="../actor" --go_out=. --go_opt=paths=source_relative --proto_path=. routercontracts.proto
::pubkey.proto forward_message.proto

echo complate 

pause 