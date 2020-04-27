@echo off
go build -a -o sinf_mirror.exe && sinf_mirror.exe -s "D:\temp\source" -d "H:\temp\dest" -w 30 -p
