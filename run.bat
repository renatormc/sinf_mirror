@echo off
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "D:\teste_report\Nova pasta" -d "H:\deletar"
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "E:\Dados da Prefeitura de Silv�nia" -d "I:\Dados da Prefeitura de Silv�nia"
go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -f inputs.txt
REM s-mirror -f inputs.txt