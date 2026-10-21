# udirstats

Analisador de uso de disco para o terminal, inspirado no WinDirStat/ncdu: varre um diretório
e mostra pastas e arquivos ordenados por tamanho, com navegação entre as pastas.

## Build

```bash
make build      # binário estático ./udirstats
make install    # copia para ~/.local/bin
make test
```

## Uso

```bash
udirstats [-x] [-apparent] [caminho]
```

- `-x` — não cruza para outros filesystems (no WSL, ignora `/mnt/c` etc.; também pula `/proc`, `/sys`).
- `-apparent` — começa exibindo o tamanho aparente em vez do uso em disco.

| Tecla | Ação |
|---|---|
| `↑` `↓` / `j` `k` | mover |
| `PgUp` `PgDn` / `Home` `End` | paginar / início e fim |
| `Enter` `→` `l` | entrar na pasta |
| `Backspace` `←` `h` `Esc` | voltar |
| `s` | alternar ordenação: tamanho, nome, nº de arquivos |
| `a` | alternar uso em disco / tamanho aparente |
| `r` | reescanear a pasta atual |
| `q` | sair |

Itens marcados com `!` tiveram erro de leitura (normalmente permissão); o erro aparece na
linha de status quando o item é selecionado. Symlinks não são seguidos e hardlinks são
contados uma única vez.
