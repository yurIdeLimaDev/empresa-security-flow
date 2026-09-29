# Índice da documentação

Consolidado em 13/09/2026. Comece pelo estado atual; depois consulte o guia da
operação desejada. Datas históricas não são novas aprovações de produção.

## Estado, arquitetura e execução

| Documento | Uso |
| --- | --- |
| [Estado atual](ESTADO_ATUAL.md) | Implementado, validado localmente e pendências externas |
| [Estado do motor](ESTADO_FLUXO.md) | Resumo publicável do fluxo técnico e limites antes do primeiro cliente |
| [Diagramas](DIAGRAMAS_MERMAID.md) | Sete fluxos, seus gates e limites de integração |
| [Arquitetura](ARQUITETURA_IMPLEMENTADA.md) | Componentes, contratos, isolamento e correção |
| [Jornada de verificação](PIPELINE_DE_GERACAO_DE_LEADS.md) | Prévia, contratação, P1, P2 e entrega |
| [Cadeia de ferramentas](CADEIA_DE_FERRAMENTAS.md) | 20 automáticas aprovadas e oito restritas |
| [Modelo canônico](MODELO_CANONICO_E_NORMALIZACAO.md) | Findings, evidências, deduplicação, ranking e comparação |
| [Motor e comandos](../pipeline/README.md) | CLI, configuração e operação |
| [Correção segura](../correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md) | BEST, limites, gates e revisão final |
| [Geração de patches](../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md) | Protocolo de gateway, ainda sem fornecedor |
| [Kit de avaliação](../correcao/avaliacao/README.md) | Replay e custo sintéticos; não eficácia de IA real |
| [Perfil Python](../correcao/adapters/python-3.13-stdlib/README.md) | Referência manual de laboratório, não perfil universal |
| [Validação local](VALIDACAO_LOCAL_MVP.md) | Comandos reproduzíveis de testes e CI |
| [Consulta GraphRAG](../pipeline/knowledge/hipporag/README.md) | Recuperação local, corpus permitido e limite entre cópias |

## Operação e site

| Documento | Uso |
| --- | --- |
| [Onboarding](runbooks/ONBOARDING_TECNICO.md) | Entradas explícitas, sem inferir autorização |
| [Ferramentas manuais](runbooks/FERRAMENTAS_MANUAIS.md) | Exceções que não podem virar execução automática |
| [Operação do host](runbooks/OPERACAO_HOST.md) | Drift, preflight, monitoramento e recuperação |
| [Emergência](runbooks/EMERGENCIA.md) | Preservação e parada por caso |
| [Revisão e entrega](runbooks/ENTREGA_E_REVISAO.md) | Patch cumulativo, migração, renderer e aceite separado |
| [Baseline Linux](../deploy/linux/README.md) | Preparação do host dedicado |
| [Aceitação do host](../deploy/linux/ACEITACAO.md) | Evidências reais ainda a produzir e coletor saneado |
| [Landing comercial](LANDING_COMERCIAL_E_OPERACAO_PUBLICA.md) | Oferta pontual e limites públicos |
| [Prévia web](PIPELINE_0_5_WEB.md) | Nome interno; não expor pipelines na interface |
| [Landing React](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/README.md) | Sete páginas, prévia efêmera e OAuth separado |
| [Deploy da landing](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/DEPLOYMENT.md) | Publicação de 26/09, configuração e rollback |

## Negócio e contratação

- [Oferta de uma página](../negocio/oferta/OFERTA_MVP_UMA_PAGINA.md): ICP, entrega,
  avulsa e assinaturas mensal/anual, decisão de 26/09/2026.
- [Pesquisa de preços](../negocio/oferta/PESQUISA_PRECOS_2026-09-26.md): concorrentes
  que corrigem, exclusões da amostra e cálculo de 15%.
- [Backlog de segurança pública](BACKLOG_SEGURANCA_APLICACAO_PUBLICA.md): controles
  existentes, lacunas, ordem e critérios de aceite.
- [Portal de cliente](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/PORTAL_CLIENTE.md): leitura autenticada,
  publicação operada, migração e limites.

- [Roadmap tecnologia/negócios](ROADMAP_TECNOLOGIA_NEGOCIOS_2026-09-11.md):
  diagnóstico original preservado, com atualização explícita das tarefas locais.
- [Índice de negócio](../negocio/README.md) e
  [guia do titular](../negocio/GUIA_DE_EXECUCAO_DO_TITULAR_PARA_MVP.md): decisões,
  contas e homologações externas, sem repetir a preparação já entregue.
- [Pacote jurídico canônico](../negocio/juridico/README.md): índice de todos
  os modelos, termos, contratação, assinatura, cobrança, retenção e aceite.
  Continua minuta não liberada; atualização técnica não é parecer jurídico.
- `negocio/modelos/` e `negocio/operacao/` contêm ponteiros para o pacote
  canônico, não modelos concorrentes a preencher.

## Históricos preservados e decisões substituídas

| Documento | Interpretação vigente |
| --- | --- |
| [Decisões incorporadas](DECISOES_DE_ATUALIZACAO.md) | Consolidação atual no início; seções de agosto são registros datados |
| [Primeiro caso](ESPECIFICACAO_IMPLEMENTACAO_PRIMEIRO_CASO.md) | Referência original manual, com nota da evolução de entrega |
| [Pesquisa de ferramentas](FERRAMENTAS_PESQUISADAS.md) | Pesquisa datada; não recotação nem aprovação automática de nova versão |
| [Revisão de adoção](REVISAO_DE_ADOCAO_FERRAMENTAS_2026-08-22.md) | Evidência pinada: preservar bytes e hashes |
| [Decisão age/Python](DEPENDENCIA_AGE_E_PERFIL_PYTHON_2026-08-24.md) | Justificativa histórica; versões efetivas vêm de locks e código |
| [Plano de negócios original](PLANO_DE_NEGOCIOS_MVP_PRIMEIROS_CLIENTES.md) | Hipóteses e pesquisa históricas; roadmap/estado prevalecem |
| [Plano de correção de agosto](PLANO_CORRECAO_VEXKEEP_REVISADO_2026-08-26.md) | Backlog histórico; não reativar a jornada antiga |
| [Check recorrente](PLANO_VEXKEEP_CHECK_V1.md) | Pausado; não oferecer assinatura como serviço ativo |

Não reescrever relatórios em `validacao/`, casos ou documentos assinados para
parecerem recentes. O manifesto do lote anterior identifica seus bytes naquela
data; documentos alterados depois terão hashes diferentes legitimamente.
O [relatório da atualização documental](../validacao/2026-09-13-documentacao/RESULTADO.md)
registra a nova conferência, separada das evidências de testes do motor.

## Regra de manutenção

Mudança de comportamento atualiza o guia responsável, o estado, o diagrama e
os testes relacionados. Manter cronologia e distinguir implementação local de
homologação remota. Nenhum documento libera automaticamente IA, fornecedor,
pagamento, assinatura ou execução fora do escopo autorizado.
