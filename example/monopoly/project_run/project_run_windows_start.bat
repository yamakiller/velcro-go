@echo off 

set GO_RUN_PATH=%cd%
echo %GO_RUN_PATH%
start cmd /k "cd /d %GO_RUN_PATH%\logs.service &&.\logs.service.exe"
start cmd /k "cd /d %GO_RUN_PATH%\gateway.service &&.\gateway.service.exe"
start cmd /k "cd /d %GO_RUN_PATH%\login.service &&.\login.service.exe"
start cmd /k "cd /d %GO_RUN_PATH%\battle.service &&.\battle.service.exe"

pause 