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

## Cores

- A barra e o tamanho usam um degradê verde → amarelo → laranja → vermelho relativo ao maior
  item da pasta; a coluna `%` é relativa à pasta atual.
- Arquivos são coloridos por tipo (imagem, vídeo, áudio, compactado, código, documento,
  binário, config); symlinks aparecem em itálico com `→`.
- Funciona em true color, 256 e 16 cores (detectado pelo terminal, cada cor fixa tem valor
  explícito para cada nível), em fundo claro ou escuro, e respeita `NO_COLOR`.

Itens marcados com `!` tiveram erro de leitura (normalmente permissão); o erro aparece na
linha de status quando o item é selecionado. Symlinks não são seguidos e hardlinks são
contados uma única vez.
