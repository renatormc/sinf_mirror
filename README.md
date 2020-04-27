# sinf_mirror

Aplicativo feito em GO cuja finalidade é espelhar uma pasta em outra. Copia arquivos novos que existem na pasta fonte mas que não existem na pasta destino, sobrescreve arquivos modificados caso o arquivo exista nas duas pastas porém com data de modificação ou tamanho diferente e deleta arquivos que existem na pasta destino caso se seja passado o parâmetro -p na linha de comando.

## Como utilizar

```bat
sinf_mirror.exe -s C:\caminho\pasta\fonte -d C:\caminho\pasta\destino -p
```
