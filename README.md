# sinf_mirror

Aplicativo feito em GO cuja finalidade é espelhar uma pasta em outra. Copia arquivos novos que existem na pasta fonte mas que não existem na pasta destino, sobrescreve arquivos modificados caso o arquivo exista nas duas pastas porém com data de modificação ou tamanho diferente e deleta arquivos que existem na pasta destino caso se seja passado o parâmetro -p na linha de comando.

## Como utilizar

```bat
sinf_mirror.exe -s C:\caminho\pasta\fonte -d C:\caminho\pasta\destino -p
```
#### Parâmetros
```

-h  --help             Print help information
-s  --source           Folder to be mirrored
-d  --destination      Folder to mirror to
-w  --workers          Number of workers. Default: 10
-t  --threshold        Size in megabytes above which there will be no
                       concurrency. Default: 8
-c  --threshold-chunk  Size in megabytes above which file will be copied in
                       chunks. Default: 8388600
-b  --buffer           Buffer size in megabytes. Default: 1
-v  --verbose          Verbose. Default: false
-p  --purge            Purge. Default: false
-r  --retries          Specifies the number of retries on failed copies.
                       Default: 10
-i  --wait             Specifies the wait time between retries, in seconds..
                         Default: 1
```
