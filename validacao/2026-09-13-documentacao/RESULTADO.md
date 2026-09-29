# Atualização da documentação e diagramas

Data: 13/09/2026. Escopo exclusivamente documental e validação local de documentos.

## Alterações

- Estado, arquitetura, decisões, jornada e roadmap sincronizados com o lote local.
- Sete diagramas centrais e três resumos alinhados: contratação/autorização,
  prévia e OAuth separados, isolamento, geração, BEST, patch canônico e restore.
- Runbooks de host, onboarding e entrega atualizados; diferença entre proposta
  manual e patch final explicitada, com migração das autorizações antigas.
- Índice por finalidade e classificação de atuais, minutas, pausados e históricos.
- Índice jurídico sincronizado sem alterar cláusulas ou declarar revisão legal.
- GraphRAG consultado primeiro: devolveu trechos de agosto. Confirmações feitas
  nos arquivos relevantes; nenhum outro modelo, subagente ou API de IA foi usado.

## Validação reproduzível

Resultados finais: 66 documentos no inventário ampliado, sem erros de UTF-8,
título ou links locais; dez blocos Mermaid passaram no parser 11.17.2.
O verificador documental da CI passou com 58 documentos e seus três testes
unitários passaram. `git diff --check` não apontou erros de whitespace.
Instalação limpa do lock de validação passou; auditoria npm dessa pasta
terminou com zero vulnerabilidades conhecidas. Não foram repetidos testes do
motor/landing, pois não houve alteração no código desses produtos.

Na raiz do projeto:

```text
python validacao/2026-09-13-documentacao/check-documentation.py
```

Nesta pasta, com Node/npm disponíveis:

```text
npm ci --ignore-scripts
npm run check
npm audit --audit-level=moderate
```

Resultados em [documents.json](documents.json) e
[mermaid-validation.json](mermaid-validation.json). Inventário/hashes refletem
os arquivos atuais, não um commit, assinatura ou atestação remota. O parser
valida sintaxe; não foi feita conferência visual por renderização de imagens.
O conteúdo dos fluxos foi comparado às regras do código e à evidência do lote.

Mermaid/jsdom são exclusivos desta pasta de validação, fixados por versão e
lockfile, com integridade dos pacotes; não entram no motor ou na landing.
A primeira versão do parser apresentou advisories na auditoria e foi
substituída por uma versão corrigida antes do fechamento da validação.

## Preservação e limites

Não foram modificados código do produto, configuração operacional, contratos
assinados, casos privados, evidências históricas ou a revisão pinada de adoção.
Datas jurídicas/de pesquisa antigas foram preservadas; esta sincronização não
recota preços nem confirma legislação, fornecedores ou configurações remotas.
Nenhum push ou deploy foi realizado.

O manifesto de `2026-09-13-lote-completo` é um snapshot anterior: mudanças
documentais posteriores explicam diferenças de hash e não exigem reescrevê-lo.
Host real, IA, perfil de cliente, assinatura/cobrança e release remota continuam
dependências externas. Não há nova declaração de prontidão comercial.
