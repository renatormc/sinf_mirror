@echo off
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "D:\teste_report\Nova pasta" -d "H:\deletar"
go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "E:\Dados da Prefeitura de Silvƒnia" -d "I:\Dados da Prefeitura de Silv„nia"
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "C:\temp\wtforms" -d "C:\temp\wtforms2"
s-mirror -s "I:\Dados da Prefeitura de Silv„nia" -d "E:\Dados da Prefeitura de Silvƒnia"