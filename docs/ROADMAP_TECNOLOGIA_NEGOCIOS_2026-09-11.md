# Vexkeep: estado atual e roadmap de tecnologia e negócios

Diagnóstico e conferência remota originais: **11/09/2026**.
Consolidação local/documental: **13/09/2026**. Resultados remotos de 11/09 são
históricos; não foram novamente conferidos nesta atualização.

**Atualização de implementação em 13/09:** os oito itens do lote local têm
estado e evidências em [ESTADO_ATUAL.md](ESTADO_ATUAL.md). Este roadmap mantém
o diagnóstico histórico e as pendências externas, não um segundo estado
concorrente. O kit sintético não é IA real; o ensaio comercial não é assinatura
ou cobrança real; host e release remota continuam pendentes.

**Lote local posterior:** checagem de credenciais antes do gateway, testes de
não transmissão/preservação de BEST, runtime Node alinhado, CI ajustada e
checagem dos documentos canônicos. Comandos e limites em
[VALIDACAO_LOCAL_MVP.md](VALIDACAO_LOCAL_MVP.md); evidência em
[fechamento local](../validacao/2026-09-11-fechamento-local/RESULTADO.md).
Isso não encerra T01 (release remota), T02 (host real), T03 (perfil de cliente)
ou a integração de IA/assinatura/pagamento.

**Atualização posterior na mesma data:** o ponto 3 recebeu revisão dos modelos
jurídicos e três procedimentos de assinatura/contratação e operação. Não houve
formalização empresarial, revisão por advogado ou ativação de assinatura e
cobrança. O estado exato está no
[fechamento do ponto 3](../negocio/juridico/18_FECHAMENTO_PONTO_3.md).

## 1. Diagnóstico direto

A Vexkeep tem uma base técnica implementada, uma demonstração pública disponível e um processo de correção exercitado em laboratório. **Ainda não há evidência suficiente para declarar pronta a operação comercial completa, com execução autônoma, correção e entrega a clientes.**

O que falta não é reconstruir os dois fluxos nem adicionar mais scanners. As lacunas principais são:

1. preparar e validar o host definitivo de execução;
2. transformar o perfil de laboratório em um perfil aplicável ao primeiro tipo de cliente;
3. conectar e validar o futuro provedor na camada de geração de patches já implementada, caso se mantenha o objetivo de intervenção humana somente na revisão final;
4. exercitar execução, recuperação, revisão e entrega na infraestrutura escolhida;
5. concluir identidade empresarial, contratos, privacidade, cobrança e atendimento;
6. ativar e testar a jornada de conta, contratação e processamento assíncrono, sem confundir isso com a prévia gratuita;
7. conseguir pilotos pagos e demonstrar custo, prazo e utilidade reais.

**Distinção importante:** um piloto operado de forma assistida pode ocorrer antes de existir checkout, portal e integração automática entre site e runner. Isso é uma alternativa comercial mais curta, não a conclusão do produto autônomo originalmente desejado. Usar preparação manual de patches nesse piloto exige assumir explicitamente essa limitação; não satisfaz a exigência de humano apenas no final.

### Como interpretar os estados

| Estado | Significado |
| --- | --- |
| Implementado | Existe código, configuração ou documento correspondente. |
| Validado em laboratório | Existe evidência de ensaio controlado; não comprova o host ou o cliente futuros. |
| Confirmado em produção | Foi possível conferir o serviço publicado nesta consulta. |
| Parcial | Parte existe, mas configuração, integração ou aceite ainda falta. |
| Não comprovado | Não encontrei comprovação nos registros consultados. Pode ter sido resolvido fora do projeto e precisa ser confirmado pelo proprietário. |
| Adiado | Foi retirado do escopo atual; não deve reaparecer como requisito sem nova decisão. |

Este levantamento consultou primeiro o GraphRAG, depois documentos, código pontual, GitHub, páginas públicas e DNS. A landing React não está no índice. Não executei scanners contra clientes, não refiz toda a suíte técnica e não alterei produção. Os testes de agosto são identificados como históricos, não como testes realizados hoje.

## 2. Decisões vigentes que o roadmap preserva

1. **Prévia pública gratuita e limitada** no site, com resultado real e parcial. Ela não é a execução integral do Pipeline 1.
2. **Contratação e autorização antes do Pipeline 1 integral.** Depois vêm confirmação técnica de escopo, Pipeline 2 quando aplicável, correção elegível e entrega.
3. Pipeline 1 de superfície pública e baixo impacto técnico, não “passivo estrito”; isso não precisa aparecer na linguagem comercial.
4. Pipeline 2 condicionado ao escopo e à autorização. Pagamento e login não autorizam testes por si sós.
5. Correção somente de segurança, preservando funcionalidades, com tentativas finitas e retenção da melhor versão validada, chamada `BEST`.
6. Uma revisão humana técnica final, vinculada ao commit e aos hashes da entrega. Não inserir aprovação humana por ticket como requisito padrão.
7. Hadrian para os casos de autorização compatíveis; cenário próprio restrito às regras de crédito/billing que exigirem validação de estado. Não retornar ao diff genérico de respostas como prova suficiente.
8. Ferramentas restritas não se tornam automáticas só porque existe um parser. dnsReaper gera candidatos; chave pública de BaaS não prova vazamento; HIBP automatizado no motor não significa liberado em qualquer contrato.
9. Projeto pontual é a oferta atual. O Check recorrente permanece pausado.
10. A landing React e a marca Vexkeep já existem. Não recriar site em HTML nem reabrir a escolha da marca/domínio.
11. Compra como convidado não equivale a anonimato financeiro. Conta própria por e-mail/senha não está implementada.

Referências: [jornada atual](LANDING_COMERCIAL_E_OPERACAO_PUBLICA.md), [fluxo técnico](PIPELINE_DE_GERACAO_DE_LEADS.md), [Check pausado](PLANO_VEXKEEP_CHECK_V1.md), [correção](../correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md).

---

## 3. Tecnologia: o que já está feito

| Frente | Entrega existente | Estado e limite real |
| --- | --- | --- |
| Motor de verificação | Runner Go, planejamento por padrão, schemas de escopo/política, módulos de verificação e onboarding | Implementado. Configurações de exemplo não autorizam execução real. |
| Normalização | `Asset`, `Finding`, `Evidence`, `ToolRun`, `RequestLog`, `Engagement`, `Scope`, `Exception`; bundle JSON, relatório, deduplicação, ranking e comparação | Implementado. Hash comprova integridade; alerta de scanner não vira verdade automaticamente. |
| Cadeia de ferramentas | 28 entradas revisadas para adoção; 20 automáticas aprovadas e oito restritas | Pendência antiga de 28 revisões e 14 containers foi fechada. As 20 automáticas têm digest, liveness e SBOM. Não são 28 ferramentas liberadas. |
| Isolamento | Rede por engajamento, firewall, bloqueio de acesso ao host, proxy, canaries, limites e encerramento | Ensaio integrado Linux root aprovado. É evidência do desenho em ambiente controlado, não do VPS definitivo. |
| Correção | Tickets, worktrees isoladas, agente pinado, gates, comparação de cobertura, limites, `BEST`, revisão final e autorização por hash | Implementado e ensaiado. A não regressão é relativa à cobertura e aos testes, não uma garantia absoluta sobre qualquer aplicação. |
| Perfil concreto | `python-3.13-stdlib`, aplicação de patch governada, testes funcionais e reteste por propriedade | Validado apenas no laboratório de referência. Recebe patch já preparado; não gera sozinho a correção. |
| Laboratório de referência | Cinco falhas intencionais, cinco correções, testes 8/8, pacote e restore com hashes conferidos | Aprovação humana simulada e assim identificada. Não é case comercial nem benchmark amplo de scanners. |
| CER-Fácil | Correção local anterior: cinco itens resolvidos, três mitigados, um preservado por escopo | Evidência histórica. ToolRuns anteriores à cadeia atual; não há autorização atual de entrega. |
| Entrega e continuidade | Relatório HTML/PDF, manifesto, pacote determinístico, criptografia age, backup/restore, runbooks e monitoramento | Implementado e exercitado em laboratório. Canal de entrega, backup off-host e resposta a alertas em produção ainda precisam de comprovação. |
| Site | React, marca, prévia pública, Turnstile novo por tentativa, loading, resultado parcial e oferta posterior | Site publicado. Página inicial, termos, privacidade e contratação responderam HTTP 200 nesta conferência. Nenhum scan foi disparado neste levantamento. |
| Proteções da prévia | Restrições de alvo, bytes e requisições, validação DNS, sem execução de JS do alvo, sem redirects e sem persistência do resultado no banco | Código e testes documentados. Revalidar o conjunto na versão de lançamento; logs de infraestrutura não equivalem a inexistência de tratamento de dados. |
| Autenticação | Better Auth, backend Google/Apple, D1, sessão, cookies seguros, OAuth state, rate limit e tokens cifrados | Parcial. Endpoint de produção retornou `google: false` e `apple: false` em 11/09. Não há login social operacional confirmado. |
| Contratação no site | Página com oferta, preço inicial, prazo, termos e solicitação por e-mail | Casca. Não recebe pagamento nem inicia projeto contratado automaticamente. |
| Check recorrente | Código de estados, prova de domínio, D1 de staging, jobs, heartbeat, cancelamento e exclusão | Preservado, mas pausado. Não tratar essas peças como portal pago já entregue. |
| GitHub e CI | Repositórios privados separados para motor, landing e laboratório; validação automatizada | Motor e landing confirmados privados, com última CI aprovada. Não confundir CI histórica com certificação permanente. |
| DNS e recebimento de e-mail | Domínio publicado; MX de Cloudflare Email Routing; SPF de encaminhamento | DNSSEC apresentou DS/DNSKEY e resposta autenticada por resolvedor. Encaminhamento não comprova caixa de envio profissional nem recebimento efetivo de uma mensagem. |
| Conhecimento e documentação | GraphRAG local, docs técnicas, runbooks, planos e pacote jurídico | Consolidada localmente em 13/09, com índice e diagramas atualizados. O MCP conectado ainda retornou trechos antigos de outra base/corpus. |

### Limite descoberto no código do laboratório

Em `pipeline/internal/app/first_case.go`, o perfil usa cinco regras `LAB-*`, caminhos e metadados próprios do laboratório. Seu `security-full` constrói um bundle a partir dessas regras; os retestes funcionais rodam em etapas separadas. **Isso demonstra o funcionamento do controle de correção, mas não a eficácia geral dos scanners nem a segurança de uma aplicação arbitrária.**

O perfil deve continuar existindo como teste de referência. O que falta é um perfil de cliente com verificadores e evidências reais, sem reaproveitar marcadores, identificadores ou hashes sintéticos como atestação comercial.

Fontes: [normalizador](MODELO_CANONICO_E_NORMALIZACAO.md), [cadeia](CADEIA_DE_FERRAMENTAS.md), [fechamento técnico](../validacao/2026-08-22-fechamento-prontidao-operacional/RESULTADO.md), [laboratório](../validacao/2026-08-24-primeiro-caso-controlado/RESULTADO.md), [perfil](../correcao/adapters/python-3.13-stdlib/README.md), [implementação do perfil](../pipeline/internal/app/first_case.go), [landing](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/README.md).

## 4. Tecnologia: o que falta fazer

Prioridades: **P0** antes de executar e entregar o primeiro caso real; **P1** antes de anunciar a jornada automática como funcional; **P2** depois da validação comercial. A prioridade não exige fazer tudo em sequência: várias frentes podem avançar em paralelo.

### T01. Consolidar versão de referência, documentação e release — P0

**Preparado localmente em 13/09:** documentação consolidada, diagramas,
manifesto local, validação da landing/motor, CI e SBOM. Esses itens não
precisam ser refeitos do zero.

**Falta:** vincular a revisão local a uma release remota, conferir sincronização
entre repositórios e validar o artefato efetivamente implantado.

- Conferir divergências entre a cópia agregadora local, o repositório privado do motor e o da landing; preservar mudanças existentes e não sincronizar tudo indiscriminadamente.
- Vincular cada release ao commit, build, lock, SBOM, configuração e CI correspondente; registrar também qual artefato foi implantado.
- Preservar na release a documentação local corrigida: P1 após contratação e
  autorização, OAuth implementado mas não homologado, sete páginas e DNSSEC
  com evidência histórica; conferir o ambiente real antes de afirmar seu estado.
- Separar claramente regras do projeto pontual e peças do Check pausado. Atualizar o índice de negócio para apontar o pacote jurídico canônico.
- Escolher local privado e backup para os documentos de negócio; não presumir que estejam protegidos pelo GitHub só porque o código está.
- Revalidar dependências e pins da release. Atualização necessária deve repetir testes; não trocar para `latest`.

**Aceite:** manifesto da release, CI correspondente e roteiro de rollback conferidos; nenhuma instrução antiga leva a executar o produto errado. Documentos de negócio recuperáveis sem publicar dados de clientes.

**Responsável:** implementação; proprietário confirma acessos e armazenamento de negócio.

### T02. Preparar o host Linux definitivo — P0

**Falta:** contratar ou identificar o host e provar que ele aplica os controles já desenvolvidos.

- Confirmar provedor, região, capacidade, política de uso para testes autorizados, custo de renovação e acesso administrativo.
- Aplicar baseline de SSH por chave, administração restrita, atualizações, armazenamento e Docker compatível com o perfil validado.
- Separar casos, evidências, temporários e backups; não compartilhar diretórios entre clientes.
- Executar build reproduzível, cadeia estrita, liveness dos runtimes necessários, drift e `deploy-preflight-linux.sh` como root em diretório novo.
- Validar firewall, `DOCKER-USER`, bloqueio do host, proxy, canaries e cleanup no host real.
- Ensaiar reboot, timeout, processo interrompido e emergência. Confirmar que falha não deixa rede ou container com permissões indevidas.
- Definir orçamento, limite de concorrência e procedimento de patching. Não congelar versões indefinidamente sem revisão de segurança.

**Aceite:** evidência saneada vinculada ao host e à release, com testes positivos e negativos aprovados. Repetir preflight com a política de cada caso, não apenas uma vez na instalação.

**Dependências:** aprovação de infraestrutura e fornecedores, N03/N05. **Responsável:** proprietário pela conta; implementação pela configuração e teste.

### T03. Criar um perfil realmente aplicável ao primeiro cliente — P0

**Falta:** transformar o suporte técnico declarado em capacidade verificável, sem prometer correção universal.

- Escolher uma stack a partir dos primeiros prospects qualificados. O laboratório Python não obriga o negócio a vender exclusivamente para aplicações Python sem dependências.
- Para a stack escolhida, definir build, testes funcionais, scanner completo autorizado, reteste de cada classe de falha e smoke test do comportamento protegido.
- Produzir evidência real: execução, requisição/estado relevante, versão, escopo, resultado e hash. Não usar os marcadores `LAB-*` como verificador de cliente.
- Proteger testes e políticas contra alteração pelo agente. Criar baseline de comportamento antes das correções.
- Declarar operações e categorias não cobertas. Stack desconhecida ou perfil incompleto deve bloquear a promessa de correção automática.
- Exercitar exemplos vulneráveis, corrigidos e sem falha, inclusive falso positivo, falha do scanner e regressão funcional.

**Aceite:** um projeto representativo fora do laboratório original passa pelo perfil; falhas conhecidas são detectadas/retestadas, regressões são rejeitadas e cobertura ausente não vira sucesso.

**Dependências:** N04, T02 e T05. **Responsável:** implementação com responsável técnico pela validação.

### T04. Integrar a geração de correções — P1; obrigatória para o objetivo de humano só no final

**Implementado nesta data, após o levantamento:** contrato de geração independente
de fornecedor, cliente HTTPS para gateway futuro, aplicação validada da proposta,
hashes, comando `remediation-run`, tentativas finitas, verificadores separados e
parada na revisão humana final. O exemplo permanece desativado por decisão do
usuário, que escolherá o fornecedor depois. Ponto 3 de negócio/jurídico pausado.

**Falta:** integração específica/credenciais do futuro fornecedor e ensaio com IA
real e gates representativos. O simulador valida a governança, não a qualidade
de correções de uma IA. O perfil manual anterior continua disponível. Ver
[contrato, limites e próximos passos](../correcao/docs/GERACAO_PATCHES_SEM_PROVEDOR.md).

- Definir provedor e configuração do agente, autenticação, custo máximo, timeout, limites de tentativas e condições de falha.
- Conectar o fornecedor à camada de geração já integrada a `remediation-agent`, mantendo o agente sem poder de aprovar ou promover `BEST`.
- Validar o gateway e seu isolamento operacional: o modelo não recebe shell/filesystem no novo contrato, mas o host e a integração do fornecedor ainda exigem restrições de rede, credenciais e recursos. Hash do executável sozinho não fornece esse isolamento.
- Tratar arquivos do cliente, comentários e saídas de ferramentas como dados não confiáveis, inclusive instruções maliciosas que tentem alterar políticas ou exfiltrar informações.
- Enviar ao provedor apenas material autorizado, com configuração de retenção e tratamento compatível com N03. Não enviar credenciais e evidência bruta indiscriminadamente.
- Repetir com o provedor real os testes de proposta inválida, dependência não autorizada, mudança não relacionada à segurança, tentativas esgotadas, indisponibilidade e orçamento. A camada local já possui testes simulados de rejeição, limites, isolamento de credencial, interrupção e preservação de BEST; custo monetário depende do gateway futuro.

**Aceite:** correção produzida pelo agente, sem patch pré-preparado, passa pelos gates reais de T03; tentativas ruins preservam `BEST`; nenhum humano precisa aprovar tickets intermediários; revisão final continua obrigatória.

**Alternativa:** operar patches preparados externamente em piloto assistido, com concordância explícita sobre a limitação. Não marcar este item como concluído nesse caso.

### T05. Validar a eficácia do conjunto de verificadores e normalizadores — P0

**Falta:** complementar liveness e ensaios de controle com evidência de detecção no escopo que será vendido.

- Montar uma matriz pequena e representativa de falhas conhecidas e controles negativos para o perfil inicial; não exigir todas as ferramentas em todo projeto.
- Confirmar as saídas reais das versões pinadas nos parsers. Testar campo ausente, saída truncada, formato desconhecido, erro e timeout.
- Conferir deduplicação sem juntar ativos ou vulnerabilidades diferentes, ranking explicável e associação correta entre achado, evidência e execução.
- Para autorização, exigir identidade e efeito observado compatíveis com o teste; para billing, conferir estado final. Um campo de saída do scanner não dispensa qualidade de evidência.
- Registrar cobertura planejada, executada e não executada. Erro ou ausência de detecção não significa ausência de falha.
- Manter baseline e candidato com políticas e versões comparáveis. Mudança de ferramenta exige nova baseline antes de comparar correção.
- Reexecutar CER-Fácil pela cadeia atual apenas se ele for usado como demonstração ou entrega. Sua revalidação não é requisito para um cliente de outra stack.

**Aceite:** resultados conhecidos são explicáveis e reproduzíveis; falso positivo permanece candidato ou é descartado; nenhum relatório apresenta “corrigido” apenas porque o scanner deixou de rodar.

### T06. Fechar backup, retenção, exclusão e resposta operacional — P0

**Falta:** sair dos scripts e testes locais para uma rotina exercitada na infraestrutura definitiva.

- Configurar backup cifrado off-host, conta restrita e expiração; manter chave privada fora do runner e testar sua recuperação.
- Definir perda máxima tolerada de trabalho e tempo de recuperação. Traduzir isso em frequência de backup e procedimento operacional.
- Restaurar em ambiente limpo, conferir hashes e recuperar um caso interrompido sem liberar execução fora de escopo.
- Testar alertas de disco, falha de backup, preflight vencido e resíduos de execução com um destinatário real, sem dados de cliente no alerta.
- Executar encerramento: interromper jobs, revogar acessos, eliminar cópias de trabalho, aplicar retenção e emitir comprovante verificável.
- Incluir dados de conta, logs e fornecedores no inventário. Exclusão no banco principal não apaga automaticamente backups ou cópias externas.
- Resolver a compatibilidade entre evidência bruta por sete dias, projeto estimado em 7–10 dias úteis, reteste e suporte. Não apagar a única prova antes de concluir a entrega nem ampliar retenção sem critério.

**Aceite:** simulação de perda do host recuperada; alerta recebido; caso encerrado com rastreabilidade; política aprovada em N03 corresponde à configuração efetiva.

### T07. Revalidar a experiência pública e seus limites — P0 para divulgação ampliada

**Falta:** um aceite da versão de lançamento, não uma nova landing.

- Testar em navegador e celular duas verificações consecutivas, token novo, expiração do Turnstile, loading, timeout e falha de rede em alvos controlados.
- Conferir que resposta da API, interface e páginas auxiliares não expõem nomes internos de pipelines, evidência sensível ou alegações excessivas.
- Testar autorização, limites de entrada, DNS privado/misto, redirects, consumo de requisições e rate limit. Conferir o controle na configuração implantada, não só no código.
- Confirmar que um resultado sem observações relevantes também recebe uma explicação honesta, sem fabricar urgência para vender.
- Concluir a identificação pública e a consistência da oferta com N02/N04. Evitar encaminhar o usuário para botões indisponíveis como se fossem contratação concluída.
- Conferir acessibilidade básica, contato, domínio principal, HTTPS e cache após deploy. DNSSEC já tem evidência atual de publicação/validação; não consta como compra ou implantação do zero.

**Aceite:** jornada pública sem erro bloqueante nos cenários testados, limitada ao que realmente executa, e canal comercial funcional. Registrar navegador, release e resultados.

### T08. Ativar e validar Google e Apple — P1

**Falta:** configuração dos provedores e teste completo. O backend já está implementado.

- Recuperar ou confirmar acesso administrativo às contas oficiais; concluir exigências de cada provedor para disponibilização ao público.
- Configurar aplicação web, domínios e callbacks exatos; guardar credenciais como secrets, nunca em Git ou bundle cliente.
- Testar login inicial e recorrente, cancelamento, state inválido, callback repetido, sessão expirada, logout e isolamento entre duas contas.
- Conferir vinculação de contas por e-mail, Apple Hide My Email e respostas sem e-mail; não tratar identificador sintético como endereço entregável.
- Validar rotação/revogação de credenciais e comportamento em falha de provedor.
- Concluir N03 antes da ativação e operacionalizar exclusão de conta e sessões.

**Aceite:** logins reais de teste aprovados no domínio de produção, testes negativos e de sessão registrados, recuperação administrativa documentada. Booleans `true` no endpoint não bastam como teste.

**Dependência do proprietário:** acesso às contas e eventuais requisitos de cadastro/assinatura. Login não é requisito técnico para um contrato assistido fora do portal.

### T09. Definir recebimento de material e entrega privada — P0; portal automático em P1

**Falta:** um percurso comprovado entre cliente, material autorizado e pacote final.

- Definir como receber escopo, repositório/cópia, contas A/B e recursos sintéticos sem usar Git público ou e-mail comum para credenciais.
- Se houver upload, bloquear traversal, symlinks indevidos, arquivos excessivos e descompactação abusiva; não executar scripts do projeto recebido fora do sandbox.
- Associar versão recebida, autorização e configuração ao caso. Confirmar controle ou poderes sobre cada ativo, não só sobre o domínio principal.
- Entregar relatório saneado, limitações, patches, commit e manifesto por canal privado; conferir destinatário e recuperação de acesso.
- Se houver área do cliente, impor autorização por caso em cada operação. Login bem-sucedido não autoriza ler outro caso.
- Não inserir resultados ou códigos de acesso sensíveis em URLs, analytics e logs. Para convidado, definir acesso, expiração e suporte sem promessa de anonimato absoluto.
- Definir quem aplica o patch no ambiente do cliente. Entregar patch não equivale a implantar em produção; implantação exige escopo próprio, backup e reversão.

**Aceite:** cliente de teste recebe e confere somente seu pacote; segundo usuário não acessa o caso; material não escapa para repositório/log indevido; alteração após aprovação invalida a entrega anterior.

### T10. Implementar cobrança online com estado confiável — P1

**Falta:** pagamento integrado. A página atual não faz cobrança.

- Após N02/N07, integrar sandbox do provedor escolhido e criar cobranças no servidor, vinculadas à proposta e ao caso.
- Validar o mecanismo oficial de autenticação do webhook, deduplicar eventos e conferir estado, valor, moeda e vínculo pela API quando necessário.
- Tratar eventos repetidos, atrasados e fora de ordem, cancelamento, vencimento, reembolso e contestação.
- Separar `pago`, `autorizado`, `onboarding concluído`, `pronto para executar` e `aprovado para entregar`. Nenhum retorno de navegador pode pular esses estados.
- Registrar versão/documentos aceitos e disponibilizar cópia conservável. Não tratar o checkbox da prévia como assinatura do contrato pago.
- Testar conciliação e recuperação quando o provedor confirmar o pagamento, mas a aplicação estiver indisponível.

**Aceite:** cobrança de teste reconciliada uma única vez; evento falso ou valor divergente não libera execução; cancelamento/reembolso tem efeito correto e trilha de auditoria.

**Caminho anterior possível:** cobrança e conciliação assistidas, documentadas, sem habilitar checkout. Isso atende um piloto, mas não fecha T10.

### T11. Conectar a jornada paga ao runner — P1

**Falta:** integração assíncrona do serviço pontual. Não basta habilitar o código do Check recorrente.

- Definir estados do projeto pago e contrato de jobs, separado da requisição gratuita do site.
- Gerar política determinística por execução a partir do escopo autorizado, com DNS/IPs, portas, limites, janela e hashes conferidos.
- Implementar agente Linux pull com credencial de serviço restrita, sem porta pública desnecessária no runner.
- Impedir job duplicado e execução simultânea indevida; implementar claim, heartbeat, timeout, retry finito, cancelamento e revogação.
- Revalidar política e DNS imediatamente antes da execução. Falta de autorização, credencial ou política válida deve bloquear, não cair em fallback permissivo.
- Devolver apenas estado e relatório saneado ao site. Saída bruta e credenciais ficam no ambiente segregado.
- Reaproveitar peças existentes somente após revisar suas premissas: periodicidade, cobrança e retenção do Check não são as do projeto pontual.

**Aceite:** um projeto pago de teste percorre site/serviço, worker e runner; duplicidade, cancelamento, reboot, credencial revogada e hash alterado não produzem execução ou entrega indevida.

**Dependências:** T02/T03/T05/T06/T09/T10. Execução pela CLI pode atender piloto assistido, mas não representa esta integração concluída.

### T12. Fazer o ensaio final de lançamento — P0; repetir em P1 com automação

**Falta:** comprovar a combinação dos componentes na release e no ambiente definitivos.

- Usar projeto controlado representativo, entradas realistas e dados fictícios; percorrer qualificação, escopo, execução, normalização, correção, revisão, entrega e encerramento.
- Exercitar caminho com falha real, sem achado, scanner indisponível, regressão, limites esgotados, cancelamento e recuperação.
- Conferir que a revisão humana é feita por pessoa capaz de avaliar diff, comportamento preservado, evidência e limitações. Não pode ser apenas marcar “ok”.
- Medir tempo de máquina, preparação, correção, revisão e suporte. O gate humano não deve ser removido para compensar uma estimativa comercial irreal.
- Registrar o que passou, o que não foi testado, risco residual e responsável pela decisão de liberação.

**Aceite:** nenhum bloqueador conhecido de segurança ou da jornada vendida; falhas controladas terminam sem entrega indevida; artefato aprovado é exatamente o entregue; N01–N07 estão resolvidos para o contrato em questão.

---

## 5. Negócios: o que já está feito

| Frente | O que existe | O que isso não comprova |
| --- | --- | --- |
| Marca e presença | Vexkeep, domínio `vexkeep.com`, identidade visual e site publicado | Não comprova registro de marca, regularidade empresarial nem aquisição de clientes. |
| Tese comercial | Revisão autorizada, evidência rastreável, correção elegível e reteste | Ainda precisa de validação por compradores reais. |
| ICP inicial | SaaS B2B brasileiro, decisor técnico acessível, aplicação web e capacidade de autorizar | Não há comprovação de uma carteira qualificada já preenchida. |
| Modelo de receita | Projeto pontual; recorrência pausada | Não existe assinatura comercial ativa comprovada. |
| Preço e prazo | Hipótese publicada de R$ 4.750 e 7–10 dias úteis após onboarding | Não são margem validada nem compromisso aplicável a qualquer escopo. |
| Jornada comercial | Prévia pública, oferta posterior e solicitação de proposta por e-mail | Checkout, contrato aceito online e entrega automática não estão concluídos. |
| Documentação jurídica | 19 arquivos originais, revisados nos pontos aplicáveis, e três procedimentos novos: assinatura digital, operação de cobrança/atendimento e fechamento do ponto 3 | São minutas/procedimentos para preenchimento e validação. Não há aprovação jurídica profissional nem assinatura real concluída nesta revisão. |
| Informação pública | Termos de Uso e Política de Privacidade publicados | Publicar texto não completa identificação empresarial, conformidade operacional ou coleta de aceite contratual. |
| Organização comercial | Plano de negócio, roteiro de descoberta e modelos de CRM/custo | Os CSVs consultados contêm exemplos; não são operação comercial preenchida. A planilha de custo ainda é orientada ao Check. |
| Provas técnicas | Ensaios privados, resultados saneados e exemplo de processo | Não são depoimentos, certificações, casos de clientes nem autorização para afirmar eficácia universal. |
| Contato | Endereço `contato@vexkeep.com` divulgado e Email Routing configurado no DNS | Falta comprovar envio autenticado, caixa monitorada e atendimento. |

**Sem comprovação suficiente nos registros consultados:** constituição/cadastro empresarial concluído, contador contratado, parecer jurídico final, conta PJ/provedor financeiro operacional, NF emitida, clientes pagantes, receita, margem, seguro contratado e casos comerciais autorizados para divulgação. Não estou afirmando que você não fez isso fora do projeto; esses itens precisam ser confirmados e documentados.

Fontes: [plano de negócio](PLANO_DE_NEGOCIOS_MVP_PRIMEIROS_CLIENTES.md), [pacote jurídico](../negocio/juridico/README.md), [pendências jurídicas](../negocio/juridico/00_PENDENCIAS_ANTES_DO_PRIMEIRO_CLIENTE.md), [CRM de exemplo](../negocio/operacao/CRM_PILOTOS.example.csv), [custo de exemplo](../negocio/operacao/CUSTO_CHECK.example.csv).

## 6. Negócios: o que falta fazer

### N01. Confirmar a identidade empresarial e a operação fiscal — antes de contratar

- Informar a entidade que prestará o serviço, identificação fiscal, endereço, representante e poderes de assinatura.
- Validar com contador atividade, enquadramento, regime tributário, município, cadastro e emissão fiscal aplicáveis. Não presumir São Paulo nem usar estimativa antiga de imposto como fato.
- Confirmar conta de recebimento, titularidade e processo de conciliação.
- Testar emissão fiscal pelo procedimento admitido para a empresa, sem emitir documento fictício de operação inexistente.
- Publicar identificação e contatos coerentes no site e nos documentos antes de habilitar venda online.

**Entrega de aceite:** cadastro empresarial confirmado, dados únicos nos documentos, orientação contábil registrada e procedimento fiscal utilizável.

**Responsável:** proprietário e contador. É dependência externa; código não a resolve.

### N02. Transformar os modelos jurídicos em documentos utilizáveis — antes do pagamento do cliente

- Preencher os modelos aplicáveis e operacionalizar os procedimentos novos, eliminando placeholders e resolvendo decisões contratuais pendentes.
- Submeter contrato, proposta, SOW, autorização, DPA e políticas a advogado que conheça a operação concreta. O pacote existente é preparação, não parecer profissional definitivo.
- Definir responsabilidade, exclusões, direito de cancelamento aplicável, reembolso, suporte, aceite, confidencialidade, propriedade/licença dos patches e tratamento de material de terceiros.
- Descrever exatamente a fronteira entre avaliação, correção e implantação. Autorizar teste não equivale a autorizar alteração em produção.
- Definir precedência dos documentos e como alterações do escopo geram aditivo e nova configuração técnica.
- Escolher assinatura com identificação, íntegra e trilha de auditoria; guardar versões, signatários e hashes.
- Entregar resumo, preço, prazo, restrições e cópia do contrato antes/depois do aceite nos momentos adequados. Não deixar a pessoa descobrir limitações somente depois de pagar.

**Entrega de aceite:** pacote preenchido, revisão profissional concluída e ensaio de assinatura/arquivamento. O caso só inicia com autorização, SOW e políticas coerentes. Checklist em Markdown não é, sozinho, bloqueio técnico de checkout.

**Responsável:** proprietário e advogado; implementação conecta registros ao fluxo.

### N03. Fechar privacidade, fornecedores e retenção — antes de receber material de cliente ou ativar novos tratamentos

- Mapear dados da prévia, contato, conta, cobrança, código, evidência, suporte e logs; identificar papéis e finalidades em cada operação.
- Completar o registro real de Cloudflare por produto, provedor de identidade, e-mail, pagamento, VPS, backup e IA. Verificar localizações e contratos efetivos, sem promessa genérica de “dados no Brasil”.
- Definir mecanismos aplicáveis às transferências internacionais e refletir isso nos documentos. A ANPD mantém regras específicas; contratar infraestrutura brasileira não elimina outros fluxos internacionais. [Referência oficial](https://www.gov.br/anpd/pt-br/assuntos/assuntos-internacionais/transferencia-internacional-de-dados).
- Aprovar prazos por categoria e seus marcos iniciais, harmonizando reteste, suporte, exclusão, preservações e backup. O padrão atual de sete dias de evidência bruta exige revisão frente ao prazo comercial.
- Estabelecer canal de direitos, verificação proporcional de identidade e registro de atendimento. Ter conta sem senha não elimina dados pessoais.
- Designar responsável por incidentes e alternativa de contato; ajustar comunicação ao cliente e obrigações aplicáveis com assessoria. Vulnerabilidade encontrada não é automaticamente incidente comunicável. [Orientação oficial da ANPD](https://www.gov.br/anpd/pt-br/canais_atendimento/agente-de-tratamento/comunicado-de-incidente-de-seguranca-cis).
- Verificar licenças e condições comerciais das ferramentas/API, especialmente módulos usados em benefício de clientes. Não reativar módulo sem condições operacionais; não apagar a capacidade técnica só porque a contratação está pendente.

**Entrega de aceite:** inventário preenchido, fornecedores aprovados para usos específicos, documentos coerentes e rotinas de T06 exercitadas. Não basta uma política pública genérica.

### N04. Fechar uma oferta que corresponda à capacidade real — antes de enviar proposta

- Escolher um problema e uma faixa de escopo iniciais: aplicação/ambiente, recursos, profundidade, duração e contatos responsáveis.
- Confirmar a stack elegível com T03. O ICP pode ser amplo; a promessa de correção precisa ser estreita e verificável.
- Definir o que o cliente compra quando a prévia ainda não confirmou falha: avaliação autorizada e entregáveis contratados, não “correção garantida” de um sinal público.
- Explicitar, na venda posterior à prévia, como serão tratados achados adicionais, falso positivo, ausência de falha confirmada, impossibilidade de corrigir e dependência do fornecedor do cliente.
- Definir se o preço inclui correções elegíveis até limite acordado ou se haverá proposta adicional. Não cobrar complemento surpresa para algo anunciado como incluído.
- Separar entrega de patch, orientação e implantação; dizer quem faz cada etapa.
- Preservar o site simples: não listar ferramentas nem score artificial. Escopo e exclusões aparecem na proposta, antes do pagamento.

**Entrega de aceite:** uma proposta que um comprador entenda sem explicação técnica extensa, e que o responsável técnico consiga executar sem ampliar autorização ou prometer cobertura inexistente.

**Responsável:** proprietário, com implementação validando viabilidade.

### N05. Calcular preço, margem e capacidade — antes de aceitar preço/prazo

- Manter R$ 4.750 como hipótese histórica até validar custo; não tratar “5% abaixo de um concorrente” como prova de sustentabilidade.
- Criar cálculo do projeto pontual, separado da planilha de recorrência: venda, onboarding, execução, IA, verificação, revisão humana, relatório, reunião, reteste e suporte.
- Incluir provedor, armazenamento, backups, licenças, câmbio quando houver, impostos confirmados, tarifas, retrabalho e reserva de risco, sem contar a mesma parcela duas vezes.
- Definir remuneração do trabalho do fundador e horas faturáveis realistas. Tempo próprio não é custo zero.
- Estabelecer piso, forma de pagamento, validade da proposta e limite de horas/escopo.
- Conferir o prazo de 7–10 dias úteis com ensaio medido e disponibilidade do cliente; definir quando o relógio começa e quais dependências suspendem o prazo.
- Limitar a quantidade de projetos simultâneos à capacidade de revisão e atendimento. Sugestão inicial: um projeto em execução por vez, até conhecer os tempos reais.

**Entrega de aceite:** planilha preenchida para o primeiro escopo, preço acima do piso aprovado e capacidade reservada. Não foi revalidado neste roadmap o preço atual dos concorrentes nem o orçamento antigo de infraestrutura.

### N06. Tornar contato, confiança e atendimento operacionais — antes da prospecção

- Testar recebimento real de `contato@vexkeep.com` e resposta com identidade profissional.
- Concluir envio autenticado, SPF coerente com os emissores, DKIM e DMARC; os registros atuais comprovam encaminhamento, não essa etapa completa. Nesta consulta não houve resposta TXT para `_dmarc.vexkeep.com`; seletores DKIM não foram inferidos.
- Definir responsável, horário, prazo de resposta, contato de emergência e substituto. Não prometer atendimento 24x7 sem estrutura.
- Publicar uma amostra saneada e claramente demonstrativa do relatório, com limitações e antes/depois verificável.
- Explicar o processo e apresentar credenciais profissionais verdadeiras. Não chamar teste de laboratório de case de cliente, certificação ou prova de proteção total.
- Conferir MFA, recuperação e titularidade das contas críticas: domínio, nuvem, GitHub, e-mail e cobrança.

**Entrega de aceite:** contato recebido/respondido de ponta a ponta, regras de suporte documentadas, demonstração verdadeira e recuperação de contas possível.

### N07. Definir pagamento, cancelamento, conciliação e nota — antes da primeira cobrança

- Confirmar provedor/conta e procedimento de cobrança para o modelo pontual. Asaas é candidato anterior, não integração concluída.
- Determinar condições comerciais, vencimento, reembolso, contestação e tratamento de cancelamento com N02.
- Separar recebimento financeiro de autorização técnica e liberação do trabalho.
- Definir responsável pela conciliação e pela nota, inclusive falha do webhook e pagamento não identificado.
- Para piloto assistido, documentar cobrança e confirmação manual confiáveis. Comprovante enviado pelo cliente não substitui conferência de recebimento na instituição.
- Para autosserviço, cumprir T10 antes de publicar checkout funcional.

**Entrega de aceite:** pedido, valor, recebimento e documento fiscal rastreáveis, canal de cancelamento e caminho de reembolso testados pelo procedimento adequado.

### N08. Executar descoberta comercial e preencher o CRM — preparação pode começar agora

- Converter o CRM de exemplo em cópia privada operacional, fora de evidências técnicas e repositórios com acesso desnecessário.
- Atualizar os estágios para a jornada vigente: contato, descoberta, qualificado, proposta, contrato/autorização/pagamento, execução, revisão, entregue e encerrado/perdido. Não deixar Pipeline 1 gratuito como estágio obrigatório antigo.
- Montar a lista inicial priorizando contatos conhecidos, indicações e decisores técnicos acessíveis. As 50 contas do plano são meta de pesquisa, não condição para atender o primeiro cliente.
- Fazer conversas individuais sobre problema, urgência, orçamento, poderes de autorização, stack e entregável útil.
- Registrar próximo passo/data e motivo de recusa. Não aumentar volume antes de saber se a oferta é compreendida.
- Não depender de encontrar falhas em terceiros para conseguir reunião. Usar demonstração controlada e solicitação voluntária.

**Entrega de aceite:** oportunidades reais e qualificadas, cada uma com responsável, dor, escopo provável e próximo passo. Escolher os pilotos por aderência, não apenas por disposição de pagar.

### N09. Realizar dois pilotos pagos e documentar o resultado — depois dos gates técnicos e comerciais

- Selecionar o primeiro caso compatível com T03/N04 e operar dentro da capacidade reservada.
- Executar contrato, autorização, pagamento e onboarding antes da avaliação contratada.
- Entregar somente o pacote aprovado na revisão humana final; declarar o que foi corrigido, mitigado, não corrigido e não testado.
- Realizar devolutiva, colher aceite/ressalvas e registrar suporte e encerramento.
- Medir apenas o necessário: horas previstas/reais, custos, margem, prazo, retrabalho e utilidade percebida. Não criar analytics complexo para validar o MVP.
- Ajustar o processo antes do segundo piloto. Dois pilotos concluídos são a primeira validação; um terceiro pode aumentar confiança antes de escalar, não precisa ser tratado como meta contraditória.
- Pedir autorização específica para case ou depoimento. Preservar confidencialidade mesmo quando a entrega foi bem-sucedida.

**Entrega de aceite:** dois projetos pagos concluídos com escopo cumprido, aceite, custo real e feedback. CI verde e zero alertas no laboratório não substituem esse resultado comercial.

### N10. Decidir expansão usando os pilotos — após a primeira validação

- Repetir a oferta se houver demanda, margem e previsibilidade; ajustar ICP/escopo se o esforço não couber no preço.
- Decidir qual lacuna reduz mais trabalho ou abre demanda comprovada: perfil de correção adicional, integração do site, portal ou acompanhamento recorrente.
- Rever responsabilidade, continuidade caso o fundador esteja indisponível e necessidade de seguro conforme contratos e risco. Registrar decisão, sem tratar seguro como universalmente obrigatório.
- Padronizar proposta, onboarding, revisão e relatório com dados dos pilotos.
- Só ampliar aquisição paga, automação comercial e simultaneidade quando atendimento e entrega suportarem o volume.

**Entrega de aceite:** decisão escrita de repetir, ajustar ou não expandir; próximo investimento associado a uma necessidade observada, não à quantidade de funcionalidades possíveis.

---

## 7. Ordem de execução e dependências

Os marcos abaixo substituem datas arbitrárias. Prazo deve ser estimado depois de confirmar acessos, fornecedores, stack e disponibilidade de revisão.

| Marco | Tecnologia | Negócios | Condição para avançar |
| --- | --- | --- | --- |
| 1. Consolidar a base | T01; inventário do host e contas; especificação do perfil real | N01, N02, N03, N04, N05 e N06; descoberta de N08 em paralelo | Empresa/representação e oferta definidas; perfil viável; custo preliminar; infraestrutura escolhida. Conversas exploratórias podem ocorrer sem vender uma entrega ainda indisponível. |
| 2. Ficar apto ao caso real | T02, T03, T05, T06, T07, T09 e T12 | Fechar N01–N07 e qualificar primeiro piloto | Ensaio operacional completo, documentos e recebimento prontos; limites da oferta compatíveis com a capacidade. |
| 3. Validar o negócio | Executar o caso pela release aprovada; corrigir falhas de processo sem ampliar escopo | N09; acompanhar custos, aceite e feedback | Primeiro e depois segundo projeto concluídos; não aumentar carga enquanto a entrega não for previsível. |
| 4. Fechar o produto autônomo | T04, T08, T10 e T11; repetir T12 na jornada automática | Ajustar suporte/contratação à jornada online | Geração de patch, conta, pagamento, processamento e entrega funcionam juntos, com revisão técnica somente no final. |
| 5. Expandir seletivamente | Somente backlog justificado | N10 | Evidência de demanda e capacidade. |

**Se o requisito for lançar já com intervenção humana apenas no final:** T04 deve ser antecipado para o marco 2. Se o requisito também incluir contratação e processamento automáticos pelo site, T08/T10/T11 entram antes da liberação. Não é correto declarar o produto concluído com a alternativa assistida sem essa distinção.

### Divisão prática de responsabilidades

| Quem | Responsabilidade principal |
| --- | --- |
| Implementação técnica | Código, perfis, integração, testes, manifests, documentação e evidências. |
| Proprietário | Acessos, contratação de fornecedores, identidade empresarial, oferta, preço, clientes e decisões de risco. |
| Advogado e contador | Decisões jurídicas/fiscais aplicáveis e revisão dos documentos preenchidos. |
| Revisor técnico final | Conferir segurança e comportamento da versão entregue. Pode ser o proprietário se tiver capacidade para essa revisão; precisa ser designado e ter tempo reservado. |

## 8. Critérios objetivos de liberação

### Para o primeiro projeto pago operado de forma assistida

- [ ] Identidade, documentos, autorização, escopo, cobrança e operação fiscal definidos.
- [ ] Oferta limitada ao perfil efetivamente validado e com custo/prazo sustentáveis.
- [ ] Host real aprovado; ferramentas e adaptadores coerentes com a release.
- [ ] Verificadores reais e retestes, não regras exclusivas do laboratório.
- [ ] Isolamento, parada, backup, recuperação, exclusão e contato de emergência testados.
- [ ] Recebimento de material e entrega privada ensaiados.
- [ ] Revisor final designado; nenhuma alteração posterior à aprovação passa sem nova validação.
- [ ] Limitações da operação assistida explicitamente aceitas; não vendê-la como produto autônomo concluído.

### Para anunciar o produto automático originalmente desejado

Além dos itens anteriores:

- [ ] Agente produz correções sem patches previamente preparados e sem aprovações por ticket.
- [ ] Conta Google/Apple testada de verdade, se anunciada como disponível.
- [ ] Contratação e pagamento online rastreáveis, com falhas e reembolsos tratados.
- [ ] Projeto pago chega ao runner por integração segura, com política por caso e cancelamento.
- [ ] Cliente recebe somente seu resultado/pacote e consegue suporte/encerramento.
- [ ] Falhas entre serviços não pulam autorização, revisão ou gates.

Nenhum destes critérios significa “sem vulnerabilidades”. Significa processo testado, limites claros, risco conhecido e operação compatível com a promessa comercial.

## 9. Backlog completo que não deve bloquear indevidamente o MVP

| Item | Estado | Quando retomar |
| --- | --- | --- |
| Check/assinatura recorrente | Pausado por decisão de produto | Demanda recorrente comprovada; definir oferta, custo, política dinâmica e operação antes de reativar. |
| Correção multistack | Não comprovada | Uma stack adicional por vez, com testes/gates próprios e procura real. |
| Conta própria por e-mail/senha e CRUD completo | Não implementada nesta versão | Se continuar requisito comercial; exige verificação, recuperação, exclusão e controles de abuso. Não fingir que login social cobre tudo isso. |
| Pagamento/entrega com anonimato forte | Pedido original não entregue; existe casca de convidado | Especificar ameaça, limites, recuperação e exigências aplicáveis. Convidado é redução de cadastro, não implementação de anonimato. |
| Portal completo do cliente | Parcial: `/conta` é sessão, `/resultado` informativo | Priorizar autorização por caso e entrega privada antes de dashboards. |
| Automação de todos os módulos restritos | Não autorizada pelo inventário atual | Apenas com revisão/necessidade específica. Não liberar as oito entradas restritas para “completar 28”. |
| HIBP em todo contrato, OAST público e revisões sob NDA | Capacidades/políticas distintas, fora do primeiro laboratório | Contrato, fornecedor, dados e escopo específicos; não bloquear os primeiros casos que não precisam deles. |
| CRM avançado, métricas extensas e aquisição paga | Adiados | Volume e conversão justificarem; controle simples basta nos pilotos. |
| Certificações formais e operação 24x7 | Não comprovadas e fora da oferta inicial | Exigência de mercado e capacidade financeira/operacional. Não anunciar antes. |
| CER-Fácil como case público | Não liberado | Nova execução governada, revisão final e autorização de divulgação. Não necessário para vender outro perfil validado. |

## 10. Divergências levantadas em 11/09 e tratamento posterior

Em 13/09 foram corrigidas as referências operacionais de páginas, banco,
DNSSEC histórico e geração/entrega, e centralizado o estado atual. Os planos
antigos de negócio e Check são referências históricas de decisões/backlog,
não parâmetros de ativação. Datas de retenção só valem no documento específico
do produto e precisam ser definidas no contrato real; não aplicar a política
do Check ao projeto pontual.

1. O plano de negócio tem atualização inicial correta, mas conserva trechos antigos que colocam Pipeline 1 no pré-venda gratuito.
2. O plano revisado de 26/08 contém blocos e ordem de deploy do Check pausado. Eles não são todos requisitos do projeto pontual atual.
3. Trechos antigos dizem que login social está fora do MVP; o código já existe, embora produção esteja desativada.
4. `landing-page/DEPLOYMENT.md` ainda menciona seis páginas, DNSSEC pendente e ausência genérica de banco, apesar da rota de termos, do D1 de autenticação e da evidência DNS atual.
5. **Corrigido na revisão posterior de 11/09:** `negocio/README.md` passou a apontar o pacote canônico e os procedimentos de contratação em `negocio/juridico/`.
6. A amostra de custo é do Check, não um orçamento preenchido da oferta pontual.
7. “Motor pronto” deve sempre distinguir controle implementado, eficácia do perfil e operação no host real. O agente manual e as regras do laboratório não comprovam correção autônoma comercial.
8. Datas e marcos de retenção precisam ser harmonizados entre documentos de Check, projeto pontual, conta e pacote jurídico.

O levantamento inicial registrou essas divergências sem alterar o produto.
A revisão posterior de 11/09 alterou os modelos jurídicos, índices e este
roadmap; não alterou código, site publicado ou runner.

## 11. Evidências desta conferência e fontes principais

### Conferido em 11/09/2026

- `https://vexkeep.com/`, `/termos`, `/privacidade` e `/contratar`: HTTP 200.
- `https://vexkeep.com/api/auth/vexkeep-configuration`: `{"google":false,"apple":false}`.
- [Landing privada](https://github.com/yurIdeLimaDev/empresa-security-landing-page): `main` em `3febc476fd001f8dd3eb465f320eab65ad25abc1`, último push em 28/08; [CI aprovada](https://github.com/yurIdeLimaDev/empresa-security-landing-page/actions/runs/33217669756).
- [Motor privado](https://github.com/yurIdeLimaDev/empresa-security-flow): `main` em `296798a0963b7d888c8867a645f89d15ac077fb1`, último push em 27/08; [CI aprovada](https://github.com/yurIdeLimaDev/empresa-security-flow/actions/runs/33034055647).
- MX aponta para Cloudflare Email Routing; SPF inclui `_spf.mx.cloudflare.net`; consulta de DMARC não retornou TXT. Não foi enviado e-mail de teste nem verificado seletor DKIM.
- DNSSEC: DS/DNSKEY publicados e consulta A em resolvedor validador com `AD=true`. Isso atualiza a observação antiga de pendência; não substitui inventário de acesso à conta Cloudflare.
- Pacote jurídico no levantamento inicial: 19 arquivos Markdown. Após a revisão do ponto 3: 22 arquivos, ainda com pendências cadastrais, profissionais e de integração explícitas. CRM e custo consultados são exemplos.

### Evidência histórica consultada, não reexecutada nesta tarefa

- [Prontidão operacional do motor](../validacao/2026-08-22-fechamento-prontidao-operacional/RESULTADO.md).
- [Isolamento Linux root](../validacao/2026-08-22-isolamento-politicas/linux-root-integration/RESULTADO.md).
- [Primeiro caso controlado](../validacao/2026-08-24-primeiro-caso-controlado/RESULTADO.md).
- Correção anterior do CER-Fácil: evidência do caso mantida somente no ambiente local.

### Documentação e código de referência

- [Arquitetura](ARQUITETURA_IMPLEMENTADA.md), [cadeia de ferramentas](CADEIA_DE_FERRAMENTAS.md) e [modelo canônico](MODELO_CANONICO_E_NORMALIZACAO.md).
- [Especificação do primeiro caso](ESPECIFICACAO_IMPLEMENTACAO_PRIMEIRO_CASO.md), [perfil concreto](../correcao/adapters/python-3.13-stdlib/README.md) e [código do perfil](../pipeline/internal/app/first_case.go).
- [Runbook de host](runbooks/OPERACAO_HOST.md), [onboarding](runbooks/ONBOARDING_TECNICO.md) e [fluxo de correção](../correcao/docs/FLUXO_DE_CORRECAO_SEGURA.md).
- [Landing e operação pública](LANDING_COMERCIAL_E_OPERACAO_PUBLICA.md), [backend do Worker](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/worker/runtime-template.mjs) e [página de contratação](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/app/contratar/page.tsx).
- [Plano de negócios anterior](PLANO_DE_NEGOCIOS_MVP_PRIMEIROS_CLIENTES.md), [pacote jurídico](../negocio/juridico/README.md), [pendências](../negocio/juridico/00_PENDENCIAS_ANTES_DO_PRIMEIRO_CLIENTE.md), [retenção](../negocio/juridico/07_POLITICA_RETENCAO_DESTRUICAO.md) e [fornecedores](../negocio/juridico/13_REGISTRO_SUBOPERADORES_TRANSFERENCIAS.md).

## 12. Conclusão

**Tecnologia:** o controle do processo está muito mais adiantado que a integração operacional e a correção autônoma para clientes. Priorizar host real, perfil real, agente e ensaio de entrega; não adicionar ferramentas para aparentar maturidade.

**Negócios:** marca, site, tese, oferta preliminar e modelos documentais existem. Formalização, revisão aplicada, preço sustentável, atendimento e pilotos ainda precisam de comprovação e execução.

O próximo resultado útil é **um serviço de escopo restrito que consiga ser contratado, executado, revisado, entregue e encerrado com evidência real**. A automação completa vem com seus próprios critérios de aceite, sem ser confundida com a demonstração pública ou com o laboratório.

## Atualização operacional consolidada em 13/09

| Item | Preparado localmente | Falta de fato |
| --- | --- | --- |
| T01 | Docs, diagramas, CI, testes e manifesto local | Sincronização, commit/release, CI remota e deploy |
| T02 | Preflight, coletor e roteiro de aceitação | Executar no host Linux definitivo |
| T04 | Protocolo, gates, sete cenários e comparação offline | Provedor/gateway e eficácia com IA real |
| T06 | Backup/restore sintético e critérios de evidência | Operação off-host e alertas reais |
| T07 | Endereços, Turnstile, dependências e 17 testes locais | Aceite em navegador/produção da release |
| Negócio | Minutas, checklists e ensaio fictício | Dados da empresa, revisão profissional, assinatura, cobrança e clientes |

O [índice](INDICE_DOCUMENTACAO.md) orienta a leitura. Os resultados locais
não encerram T03/T05: perfil e eficácia para cliente real ainda precisam ser
homologados. Não foi contratado nenhum serviço por esta atualização.
