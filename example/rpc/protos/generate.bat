@echo off
for /r %%i in (*.proto) do (          
    echo %%~ni.proto
    protoc  --go_out=../protos ./%%~ni.proto
)

echo "make complate!!!"
pause