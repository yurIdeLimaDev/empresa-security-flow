# Decisão de dependência — age e perfil Python

O primeiro caso usa `filippo.io/age` 1.3.1 como biblioteca Go, pinada em
`go.mod`/`go.sum` e declarada no CycloneDX. Não baixa executável durante a
operação. O objetivo é criptografia autenticada para destinatário X25519, com a
chave privada mantida fora do host.

O perfil Python reutiliza a imagem já aprovada em `tools.lock.json`; não adiciona
uma ferramenta aleatória. O runner valida lock, SBOM, revisão e hash do artefato
antes de executar os gates.

Riscos residuais: a guarda/rotação da identidade age é responsabilidade
operacional; perder a identidade torna o backup irrecuperável. Por isso, backup
só é concluído depois de restauração verificada.
