# Statement of Work e regras de engajamento

**Modelo revisado em 11/09/2026 — não liberado para assinatura.** Integra o
pacote eletrônico identificado abaixo. Campos devem ser preenchidos e
validados antes do envio ao signatário.

## 1. Identificação

- pacote eletrônico: `[ID E VERSÃO]`;
- contrato-mestre: `[ID, VERSÃO E SEÇÃO OU HASH DE DOCUMENTO ANTERIOR ASSINADO]`;
- proposta: `[ID, VERSÃO E SEÇÃO]`;
- SOW: `[ID E VERSÃO]`;
- engagement ID: `[ID]`;
- case ID: `[ID OPERACIONAL]`;
- Cliente/titular dos ativos: `[PARTE]`;
- responsável técnico do Cliente: `[NOME E CANAL]`;
- contato de emergência disponível durante toda a janela: `[NOME E CANAIS]`;
- responsável da Contratada: `[NOME E CANAL]`.

## 2. Objetivo e entregáveis

**Objetivo de negócio:** `[PREENCHER]`  
**Hipóteses de segurança:** `[PREENCHER]`

Entregáveis selecionados:

- [ ] relatório executivo;
- [ ] relatório técnico saneado;
- [ ] evidências confirmadas;
- [ ] perfil de superfície/stack;
- [ ] patches de correção;
- [ ] reteste dos itens corrigidos;
- [ ] reunião de devolutiva;
- [ ] comprovante de encerramento.

Critérios objetivos de aceite: `[PREENCHER]`.

Resultado sem falha confirmada: `[ENTREGA DE COBERTURA/LIMITAÇÕES E CONDIÇÃO
COMERCIAL ACORDADA]`.

Correção inviável ou tentativas esgotadas: `[MITIGAÇÃO/RECOMENDAÇÃO, SALDO NÃO
EXECUTADO E TRATAMENTO COMERCIAL]`. Não presumir sucesso nem cobrar ampliação
sem proposta aceita.

## 3. Escopo fechado

| Ativo/origem | Titular | Ambiente | IP/porta | Incluído? | Prova de autorização |
| --- | --- | --- | --- | --- | --- |
| `[HTTPS://EXEMPLO]` | `[PARTE]` | `[PROD/HML]` | `[LISTA]` | sim | `[REFERÊNCIA]` |

Qualquer subdomínio, IP, API, tenant, aplicativo móvel, provedor ou integração
não listado está excluído. Wildcards são proibidos sem inventário anexado.

O Cliente apresenta a cadeia de poderes; a Contratada confere compatibilidade
com o escopo. Prova de DNS não autoriza contas, dados ou infraestrutura de
terceiros. Mudança de endereço IP exige revalidação de titularidade e política;
se alterar o escopo autorizado, exige nova autorização antes da execução.

## 4. Janela e orçamento operacional

- início/fim/fuso: `[PREENCHER]`;
- dias e horários permitidos: `[PREENCHER]`;
- máximo de requisições: `[PREENCHER]`;
- requisições por segundo: `[PREENCHER]`;
- concorrência: `[PREENCHER]`;
- timeout e tentativas: `[PREENCHER]`;
- IPs de origem da Contratada: `[PREENCHER]`;
- manutenção ou freeze do Cliente: `[PREENCHER]`.

## 5. Identidades e dados de teste

| Identidade | Papel | Dados permitidos | Ações permitidas | Referência no cofre |
| --- | --- | --- | --- | --- |
| A | `[PAPEL]` | `[SINTÉTICOS]` | `[LISTA]` | `[ID, NUNCA SEGREDO]` |
| B | `[PAPEL]` | `[SINTÉTICOS]` | `[LISTA]` | `[ID, NUNCA SEGREDO]` |

É proibido inserir credenciais no SOW, Git, tickets abertos ou e-mail.

## 6. Ações autorizadas

Marcar e detalhar somente o necessário:

- [ ] reconhecimento de superfície pública;
- [ ] autenticação com contas de teste;
- [ ] autorização entre identidades A/B;
- [ ] validação de API/esquema;
- [ ] injeção controlada sem exfiltração;
- [ ] XSS refletido/armazenado com payload inerte;
- [ ] análise de dependências/código fornecido;
- [ ] teste de concorrência/regra de negócio;
- [ ] correção em branch isolada;
- [ ] reteste da correção;
- [ ] outra: `[DESCREVER E LIMITAR EFEITO]`.

Ferramentas/módulos autorizados e versões: `[ANEXO OU MATRIZ]`.

## 7. Proibições

No produto atual são proibidos e não podem ser liberados apenas por marcar
"outra ação" ou alterar uma configuração:

- negação de serviço ou teste de carga;
- engenharia social, phishing ou contato com funcionários;
- persistência, backdoor ou malware;
- destruição, alteração ou exfiltração de dados reais;
- acesso a terceiros, tenants ou contas não listados;
- quebra de senha, credential stuffing ou uso de credenciais vazadas;
- publicação de vulnerabilidade;
- alteração direta em produção;
- compra, fraude, transação financeira real ou envio de comunicação externa;
- contorno de controles para objetivo não descrito neste SOW.

Implantação autorizada em produção, se futuramente contratada, exige documento
específico com ações, responsáveis, backup e reversão; não decorre da autorização
para testar. As demais exclusões não são ofertas opcionais deste MVP.

## 8. Parada de emergência

Gatilhos: indisponibilidade, degradação não prevista, dado sensível inesperado,
terceiro atingido, orçamento excedido, divergência de escopo, contato do Cliente
ou qualquer risco material.

Procedimento:

1. interromper novas ações;
2. preservar apenas logs mínimos;
3. notificar `[CONTATO]` por `[CANAL]`;
4. registrar horário, ação anterior e estado;
5. retomar somente após autorização escrita e, quando necessário, aditivo.

Frase/código de parada acordado: `[PREENCHER]`.

## 9. Correção e implantação

- repositório/commit base: `[REFERÊNCIA]`;
- branch ou worktree de entrega: `[REFERÊNCIA]`;
- stack e versões suportadas: `[PREENCHER]`;
- testes funcionais do Cliente: `[PREENCHER]`;
- controles de não regressão de segurança: `[PREENCHER]`;
- responsável pelo deploy: `[CLIENTE/CONTRATADA]`;
- plano de rollback: `[PREENCHER]`.

A Contratada altera apenas segurança dentro dos achados aceitos. Alteração de
produto, UX, performance ou arquitetura não necessária à correção está excluída.

O reteste e a comparação de segurança se limitam à cobertura identificada. A
revisão técnica humana final confere o pacote exato antes de entregar; aceite
comercial ou pagamento não a substituem.

Uso de IA para código ou evidência: `[NÃO / SIM, PROVEDOR, FINALIDADE, MATERIAL,
LOCALIZAÇÃO E CONDIÇÕES APROVADAS NO DPA]`. Campo em branco não autoriza envio.

## 10. Dados, evidências e retenção

- categorias previstas: `[PREENCHER]`;
- evidência bruta: acesso restrito e prazo do documento `07`;
- relatório saneado: `[PRAZO]`;
- backups cifrados: `[PRAZO]`;
- canal de direitos/exclusão: `[PREENCHER]`;
- local/suboperadores: `[ANEXO 13]`.

Achado inesperado de dados sensíveis suspende a ação e aciona o contato.

## 11. Preço, marcos e dependências

- preço total: `[PREENCHER]`;
- pagamento: `[PREENCHER]`;
- início após gates: `[PREENCHER]`;
- entrega estimada: `[PREENCHER]`;
- prazo de aceite: `[PREENCHER]`;
- dependências do Cliente: `[PREENCHER]`.

- suporte pós-entrega e canais: `[PRAZO E PROCEDIMENTO 17 APROVADOS]`;
- limitação de responsabilidade: `[NÃO ADOTADA / CONDIÇÃO DESTACADA VALIDADA]`;
- início durante eventual prazo legal de arrependimento: `[PROCEDIMENTO
APROVADO / AGUARDAR O PRAZO APLICÁVEL]`, sem renúncia a direito obrigatório.

## 12. Anexos e integridade

- anexo técnico legível com ativos, ações, limites e políticas: `[ID/VERSÃO]`;
- arquivos de política já congelados sem referência ao próprio pacote: `[LISTA,
VERSÕES E SHA-256]`;
- autorização incluída no mesmo pacote: `[ID, VERSÃO E SEÇÃO]`;
- DPA: `[ID, VERSÃO E SEÇÃO OU JUSTIFICATIVA DE INAPLICABILIDADE]`;
- demais anexos assinados: `[LISTA COMPLETA]`.

O arquivo final assinado e o `scope.json` operacional recebem hashes no
manifesto externo de fechamento, conforme o procedimento `16`. O SOW não
incorpora antecipadamente o hash de um `scope.json` que, por sua vez, contenha
o hash deste SOW assinado. Identificadores e conteúdo autorizado permanecem
imutáveis; somente as referências administrativas necessárias à execução são
acrescentadas depois, com conferência registrada. Se a política não puder ser
fechada sem alterar autorização, não executar: corrigir e obter nova assinatura.

Se documento humano e JSON divergirem, vale o limite mais restritivo e a
execução permanece bloqueada até correção e novo aceite.

**Cliente:** `[NOME, CARGO, ASSINATURA, DATA E FUSO]`  
**Contratada:** `[NOME, CARGO, ASSINATURA, DATA E FUSO]`
