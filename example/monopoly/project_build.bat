@echo off 
del .\bin\ /f /s /q

::编译器路径
set GO_COMPILER_PATH=%cd%

start cmd /k "cd /d %GO_COMPILER_PATH%\logs.service &&go build  -o %GO_COMPILER_PATH%\bin\logs.service -buildvcs=false"
echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\logs.service\config.yaml  %GO_COMPILER_PATH%\bin\logs.service\config.yaml

start cmd /k "cd /d %GO_COMPILER_PATH%\gateway.service &&go build  -o %GO_COMPILER_PATH%\bin\gateway.service -buildvcs=false"
echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\gateway.service\config.yaml %GO_COMPILER_PATH%\bin\gateway.service\config.yaml
echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\gateway.service\routes.yaml %GO_COMPILER_PATH%\bin\gateway.service\routes.yaml

start cmd /k "cd /d %GO_COMPILER_PATH%\battle.service &&go build  -o %GO_COMPILER_PATH%\bin\battle.service -buildvcs=false"
echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\battle.service\config.yaml %GO_COMPILER_PATH%\bin\battle.service\config.yaml
echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\battle.service\routes.yaml %GO_COMPILER_PATH%\bin\battle.service\routes.yaml

start cmd /k "cd /d %GO_COMPILER_PATH%\login.service &&go build  -o %GO_COMPILER_PATH%\bin\login.service -buildvcs=false"
echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\login.service\config.yaml %GO_COMPILER_PATH%\bin\login.service\config.yaml
echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\login.service\routes.yaml %GO_COMPILER_PATH%\bin\login.service\routes.yaml

echo F|xcopy /f /s /q /y %GO_COMPILER_PATH%\project_run\project_run_windows_start.bat %GO_COMPILER_PATH%\bin\project_run_windows_start.bat
pause 