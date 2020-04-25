@echo off
go build -a -o "%SINFTOOLS%\tools\sinf_mirror\sinf_mirror.exe" && s-mirror -s "D:\teste_report\Nova pasta" -d "H:\deletar" -w 30 -p
REM go build -a && sinf_mirror.exe -s "H:\30.2020" -d "H:\deletar" -w 10 -p
REM go build -a && sinf_mirror.exe -s "D:\teste_report\Nova pasta" -d "D:\teste_report\c1_copia_deletar" -w 10 -v -p
