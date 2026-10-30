#!/usr/bin/env bash
# Monta uma home fictícia para a demonstração do README. Os arquivos são
# esparsos: têm o tamanho aparente indicado, mas quase não ocupam disco
# (por isso a demo roda com -apparent).
set -euo pipefail

root=${1:?uso: demo-tree.sh <destino>}
rm -rf "$root"

f() { # f <tamanho> <caminho>
	mkdir -p "$root/$(dirname "$2")"
	truncate -s "$1" "$root/$2"
}

f 4300M Vídeos/ferias-2025.mp4
f 2100M Vídeos/aniversario.mkv
f 1800M Vídeos/gravacoes/reuniao-planejamento.mp4
f 950M  Vídeos/gravacoes/apresentacao.webm
f 620M  Vídeos/tutorial-go.mov

f 6100M Downloads/ubuntu-26.04-desktop-amd64.iso
f 1200M Downloads/jogos/instalador.exe
f 340M  Downloads/backup-fotos.zip
f 85M   Downloads/vscode.deb
f 12M   Downloads/boleto.pdf

for i in $(seq -w 1 18); do f "$((3 + 10#$i % 5))M" "Fotos/2025/IMG_20250$((10#$i % 9 + 1))_$i.jpg"; done
for i in $(seq -w 1 9); do f "$((20 + 10#$i * 3))M" "Fotos/raw/DSC_0$i.raw"; done

f 210M Projetos/site/node_modules/.cache/webpack.pack
for p in react next typescript esbuild lodash; do f "$((RANDOM % 40 + 8))M" "Projetos/site/node_modules/$p/index.js"; done
f 120K Projetos/site/src/app.tsx
f 4K   Projetos/site/package.json
f 18M  Projetos/api/bin/api
f 64K  Projetos/api/main.go
f 2K   Projetos/api/go.mod
f 900M Projetos/ml/dados/treino.csv
f 450M Projetos/ml/modelo.bin
f 30K  Projetos/ml/treinar.py

for a in rock jazz mpb; do
	for i in 1 2 3 4; do f "$((6 + RANDOM % 8))M" "Músicas/$a/faixa-0$i.flac"; done
done

f 24M  Documentos/contratos/aluguel.pdf
f 3M   Documentos/planilha-gastos.xlsx
f 1M   Documentos/curriculo.docx
f 600K Documentos/notas.md

f 1500M .cache/go-build/cache.bin
f 380M  .cache/pip/wheels.whl
f 90M   .local/share/Trash/files/antigo.tar.gz
f 8K    .bashrc
f 2K    .gitconfig
