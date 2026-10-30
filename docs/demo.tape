# Gera docs/demo.gif. Rode a partir da raiz do repositório:
#   make build && vhs docs/demo.tape

Output docs/demo.gif

Require ttyd
Set Shell bash
Set FontSize 15
Set Width 1100
Set Height 400
Set Padding 16
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms

Hide
Type "bash docs/demo-tree.sh /tmp/udirstats-demo/home/ana && export PATH=$PWD:$PATH && cd /tmp/udirstats-demo/home && clear"
Enter
Sleep 500ms
Show

Type "udirstats -apparent ana"
Sleep 400ms
Enter
Sleep 2s

# entra em Vídeos/gravacoes e volta
Enter
Sleep 1.5s
Down
Sleep 400ms
Enter
Sleep 1.2s
Left
Sleep 600ms
Left
Sleep 800ms

# Projetos, ordenando por nome e por nº de arquivos
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 400ms
Enter
Sleep 1.2s
Type "s"
Sleep 1.2s
Type "s"
Sleep 1.2s
Type "s"
Sleep 600ms
Down
Sleep 400ms
Enter
Sleep 1.2s
Left
Sleep 500ms
Left
Sleep 1s

Type "q"
Sleep 300ms

Hide
Type "rm -rf /tmp/udirstats-demo"
Enter
