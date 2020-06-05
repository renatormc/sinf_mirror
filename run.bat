@echo off
REM go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "D:\teste_report\Nova pasta" -d "H:\deletar"
go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "D:\\" -d "C:\temp"
