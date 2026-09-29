# Anexo de proteção de dados pessoais

Este Anexo integra o Contrato e o SOW `[ID]`. Os papéis são definidos por cada
operação real. A tabela abaixo deve ser preenchida antes do início.

Revisado em 11/09/2026. Minuta sem aprovação profissional ou fornecedores
definitivos. Nenhuma disposição autoriza tratamento que ainda não tenha sido
avaliado e habilitado para a operação concreta.

## 1. Partes e papéis

| Operação | Controlador | Operador | Finalidade | Base legal indicada pelo controlador |
| --- | --- | --- | --- | --- |
| cadastro e faturamento Vexkeep | Contratada | fornecedores listados | contratar, cobrar e cumprir obrigações | contrato/obrigação legal |
| dados presentes no ambiente do Cliente | Cliente | Contratada | executar SOW | `[PREENCHER PELO CLIENTE]` |
| segurança e logs da plataforma Vexkeep | Contratada | Cloudflare e outros ativos | proteger serviço e comprovar operação | contrato/interesse legítimo/obrigação |

A denominação contratual não substitui a avaliação funcional. Se uma parte
determinar finalidade e meios essenciais por conta própria, responderá como
controladora por essa operação.

Assinatura contratual e dados dos representantes compõem operação própria de
contratação. O provedor de assinatura, seus papéis, dados, países e retenção
devem constar do registro `13`; não são automaticamente suboperação do teste
no ambiente do Cliente. Login social e aceite contratual não são consentimento
genérico para tratamentos opcionais.

## 2. Instruções documentadas

Quando atuar como operadora, a Contratada:

1. tratará dados somente para executar o SOW, este DPA e instruções escritas;
2. informará instrução aparentemente ilícita, suspendendo-a quando necessário;
3. não venderá, enriquecerá, anunciará, treinará modelo ou reutilizará dados do
   Cliente para finalidade própria;
4. não moverá dados para serviço ou país não listado sem autorização e
   mecanismo jurídico aplicável;
5. limitará pessoas autorizadas e exigirá confidencialidade;
6. devolverá ou eliminará dados no término, ressalvadas obrigações legais e
   preservação probatória lícita.

Uso de IA exige identificação do provedor, finalidade, material permitido,
retenção, eventual uso para treinamento e medidas aprovadas. Uma cláusula
genérica de confidencialidade não basta para autorizar envio de código ou
evidência. Material não aprovado permanece fora desses serviços.

## 3. Descrição do tratamento

- objeto e duração: `[PREENCHER]`;
- natureza das operações: coleta limitada, acesso, análise, comparação,
  saneamento, armazenamento cifrado e exclusão;
- finalidades: confirmar achados, corrigir itens contratados, retestar e
  entregar evidência;
- titulares: usuários, funcionários, clientes e representantes do Cliente
  eventualmente presentes no escopo;
- dados comuns previstos: identificadores técnicos, contas de teste, logs,
  conteúdo mínimo de requisição/resposta e contato profissional;
- dados sensíveis previstos: **nenhum**;
- dados de crianças/adolescentes previstos: **nenhum**;
- volume e frequência: `[PREENCHER]`;
- locais e transferências: Anexo `13`;
- retenção: Anexo `07`.

Dados sensíveis, de crianças, financeiros, de autenticação reais ou protegidos
por sigilo encontrados inesperadamente exigem minimização, interrupção da ação
associada e contato do Cliente.

## 4. Deveres do Cliente/controlador

O Cliente garante que:

- possui base legal e transparência adequadas;
- as instruções são lícitas, necessárias e proporcionais;
- forneceu preferencialmente dados sintéticos e contas de teste;
- respondeu solicitações de titulares e decidiu comunicações regulatórias;
- não instruiu coleta massiva ou tratamento alheio ao SOW;
- comunicará restrições setoriais e sigilos especiais antes da execução.

## 5. Segurança

A Contratada manterá, proporcionalmente ao risco:

- autenticação forte, privilégio mínimo e contas individualizadas;
- cofre de segredos e proibição de credencial em Git/e-mail;
- criptografia em trânsito e repouso quando suportada;
- host e containers isolados, egress controlado e ferramentas pinadas;
- logs minimizados, integridade por hash e evidência saneada;
- backups cifrados e restauração testada;
- gestão de vulnerabilidades, atualização controlada e resposta a incidente;
- exclusão verificável e registro de acesso às evidências.

## 6. Suboperadores

6.1. Os suboperadores efetivos constam do registro `13`, com serviço,
finalidade, dados, país, DPA e mecanismo de transferência.

6.2. O Cliente autoriza apenas os suboperadores marcados como `aprovado` no SOW.
Mudança material será informada antes de afetar dados do Cliente, permitindo
objeção fundamentada e solução razoável.

6.3. A Contratada imporá obrigações de proteção compatíveis e continuará
responsável por suas próprias escolhas e instruções, sem prometer controle
absoluto sobre infraestrutura de terceiro.

## 7. Transferência internacional

Transferência internacional só ocorrerá com hipótese válida do art. 33 da LGPD
e requisitos da Resolução CD/ANPD nº 19/2024. Quando o mecanismo escolhido
forem cláusulas-padrão da ANPD, o texto oficial será incorporado integralmente,
sem alteração, em instrumento próprio entre exportador e importador.

O controlador manterá no site as informações de transparência exigidas para as
transferências efetivamente realizadas.

## 8. Direitos dos titulares

8.1. Solicitações recebidas pela operadora serão encaminhadas ao controlador
sem resposta de mérito, salvo instrução ou dever legal próprio.

8.2. A Contratada auxiliará, considerando natureza do tratamento, com busca,
exportação, correção, restrição ou exclusão tecnicamente possível.

8.3. Prazo interno para encaminhar solicitação ao Cliente: `[2]` dias úteis.
Prazo e decisão legal permanecem com o controlador.

## 9. Incidentes

9.1. A Contratada notificará o Cliente sem demora indevida e, como meta, em até
`[24]` horas da confirmação de incidente relacionado aos dados do projeto.

9.2. A notificação progressiva conterá, quando conhecido: natureza, categorias
e quantidade aproximada; sistemas e países; medidas adotadas; riscos; contato;
e informações faltantes.

9.3. O Cliente, como controlador dos dados do seu ambiente, decide as
comunicações à ANPD e aos titulares. A Contratada prestará assistência e não
atrasará fatos já confirmados para aguardar investigação completa.

9.4. Cada controlador mantém por pelo menos cinco anos o registro de incidentes
com dados pessoais sob sua responsabilidade, conforme regulamentação vigente.

## 10. Auditoria e evidências de conformidade

A Contratada fornecerá, de forma razoável e sem expor outros clientes,
informações sobre controles, suboperadores, exclusão e incidentes. Auditoria
presencial exige motivo, aviso, confidencialidade, escopo e custo acordados,
salvo dever de autoridade.

## 11. Término

Ao término, a Contratada revoga acessos, devolve o acordado, elimina dados
conforme o Anexo `07` e emite comprovante. Dados legalmente preservados ficam
isolados, com acesso restrito e sem uso operacional.

**Cliente/controlador:** `[ASSINATURA, DATA E FUSO]`  
**Contratada/operadora ou controladora própria:** `[ASSINATURA, DATA E FUSO]`
