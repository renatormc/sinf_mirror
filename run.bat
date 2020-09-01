@echo off
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "D:\teste_report\Nova pasta" -d "H:\deletar"
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "E:\Dados da Prefeitura de Silv�nia" -d "I:\Dados da Prefeitura de Silv�nia"
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -f inputs.txt
go build -a -o "C:\Users\Will\Desktop\git\test go\sinf_mirror.exe" && "C:\Users\Will\Desktop\git\test go\sinf_mirror.exe" -s F:\pubgLite\PUBGLite -d R:\folder -l "C:\Users\Will\Desktop\logOutput"
REM  && s-mirror -c case_test
REM s-mirror -f inputs.txt
