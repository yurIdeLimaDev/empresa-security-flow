# Landing comercial e verificação pública

Revisão local em 26/09/2026: oferta avulsa + assinatura de correções
mensal/anual, 2 pedidos por mês, sem reativar Check e sem checkout.
Fonte comercial: [oferta](../negocio/oferta/OFERTA_MVP_UMA_PAGINA.md).
Portal: [guia](https://github.com/yurIdeLimaDev/empresa-security-landing-page/blob/main/PORTAL_CLIENTE.md). Riscos ainda abertos:
[backlog público](BACKLOG_SEGURANCA_APLICACAO_PUBLICA.md). Este lote não é deploy.

Descrição do produto de agosto, com revisão local em 13/09/2026.
[Estado consolidado](ESTADO_ATUAL.md). Não houve novo deploy nesta revisão.

## Decisão de produto

A landing de `vexkeep.com` começa pelo endereço do site. A pessoa confirma que
pode solicitar a verificação, passa pelo controle contra abuso e recebe na tela
uma observação real, parcial e datada. Somente depois desse resultado aparece a
oferta de correção.

A interface pública não expõe os nomes internos dos fluxos técnicos. Ela fala
em verificação inicial, avaliação contratada e correção, sem prometer escopo
que ainda depende da proposta aceita.

O site não mostra uma tabela genérica de itens incluídos ou excluídos antes da
observação. A oferta posterior descreve os sinais que seriam confirmados, os
limites, o prazo e o preço inicial. O escopo definitivo continua sendo a
proposta aceita, o SOW e as políticas do engagement.

## O que roda no site: Pipeline 0.5

O Worker executa uma leitura pública ampliada do domínio informado:

- aceita somente HTTP ou HTTPS nas portas 80 e 443;
- remove caminho, consulta e fragmento; rejeita entrada com credenciais;
- recusa IP literal, nomes locais e resolução DNS para endereços reservados;
- repete a validação DNS antes de cada conexão ao alvo;
- consulta DNS, HTTP/HTTPS, oito cabeçalhos, CSP, atributos agregados de
  cookies, SPF/DMARC/MX/DNSSEC/CAA e security.txt;
- usa Certificate Transparency para no máximo quatro candidatos estritamente
  subordinados ao domínio e verifica somente HEAD por HTTPS;
- lê HTML e até dois scripts do mesmo host com limite de bytes para fingerprint
  textual conservador; nenhum script é executado;
- não segue redirecionamentos, não tenta login e não envia formulário;
- limita a execução a 46 subrequisições e oito observações;
- não grava o alvo ou o resultado em banco de dados nesta versão.

A API exige origem da Vexkeep, cabeçalho próprio, JSON limitado a 4 KiB,
confirmação de autorização e token Turnstile validado no servidor. O resultado
é verdadeiro para a requisição feita, mas não é um pentest, não confirma uma
vulnerabilidade e não garante ausência de falhas.

Cada envio exige um token Turnstile novo. Depois de cada tentativa a interface
remove e recria o widget e o servidor valida o token de uso único. O modo
Managed da Cloudflare pode aprovar uma pessoa sem exibir desafio visual; isso
não reutiliza o token anterior nem reduz a validação no servidor.

Enquanto a requisição está em andamento, o botão mostra apenas um círculo de
carregamento e `Verificando...`. O HTML exige revalidação e o JavaScript cliente
é referenciado por uma versão explícita e servido com `Cache-Control: no-store`,
para que uma publicação nova não preserve textos ou comportamentos anteriores.

Essa função é o **Pipeline 0.5 web**, não a execução integral do Pipeline 1.
Subfinder, Amass, httpx, testssl, normalização e o restante do Pipeline 1 exigem
o host Linux isolado, políticas e execução governada. O Pipeline 1 não é
gratuito: ele ocorre somente após contratação e autorização e permanece
obrigatório antes do Pipeline 2.

## Vexkeep Check v1 pausado

O Check foi retirado da oferta pública nesta etapa. Seu desenho continua sendo
um acompanhamento recorrente de **um domínio** e da sua superfície
pública. Ele não é uma correção automática, não usa credenciais e não executa
o Pipeline 2. A prévia gratuita continua efêmera; o Check começa somente após
prova de controle do domínio por um TXT `_vexkeep-verify` ou pelo arquivo
`/.well-known/vexkeep-verification.txt`.

Cada caso recebe `case_id` e access code aleatórios. O código é exibido apenas
na criação e ambos são necessários na página `/resultado`; eles não são
colocados em URLs, cookies ou analytics. O estado é controlado no servidor:
`created`, `verification_pending`, `verified`, `payment_pending`, `active`,
`queued`, `running`, `report_ready`, `failed`, `cancelled`, `expired`,
`deletion_pending` e `deleted`.

O Worker guarda somente metadados mínimos e relatório saneado no D1. O runner
Linux faz pull autenticado e recebe apenas caso, origem, IDs/hashes das
políticas, janela e limites. Ele não recebe e-mail, cobrança, credenciais ou
acesso autenticado. A política dedicada
`pipeline/config/examples/tool-policy-check.json` só permite Subfinder, Amass
passivo, httpx, testssl, jsluice e dnsReaper como candidato interno; HIBP,
TruffleHog, Gitleaks, credenciais e Pipeline 2 ficam desligados.

O resultado contém resumo, mudanças, no máximo cinco observações com evidência
pública, limite explícito e próximo passo. Não há score, CVSS inventado,
saídas brutas, e-mails, URLs internas ou logs. Metadados ficam enquanto o
serviço estiver ativo; relatório saneado por 30 dias após cancelamento;
evidência bruta do host por no máximo sete dias; backups cifrados expiram no
prazo operacional documentado. Exclusão invalida o acesso, impede novos jobs,
remove dados operacionais e deixa apenas comprovante mínimo; backups não são
descritos como excluídos instantaneamente.

O Check não está habilitado em produção: D1 e migração foram provisionados
somente em staging; faltam secrets, política dinâmica, runner isolado e webhook
de pagamento autenticado. Além desses gates, o Worker agora exige
`CHECK_PUBLIC_ENABLED=true` para expor qualquer rota do Check ou executar o
scheduler. Produção não define essa variável; as rotas retornam 404.

## Oferta e preço de lançamento

A linha pública atual é o **Pipeline 0.5 gratuito**, seguido de proposta para
projeto pontual. O Pipeline 1 integral é parte do projeto contratado, não uma
assinatura gratuita ou uma ativação automática no site.

A correção focada com reteste continua sendo projeto pontual e separado. O
valor de **R$ 4.750** permanece somente como hipótese comercial anterior para
um projeto elegível, nunca como preço do Check nem como promessa automática.

A referência brasileira pública escolhida anuncia R$ 5.000 por uma aplicação
web ou API, com prazo de até 10 dias. Aplicar 5% abaixo desse valor produz:

```text
R$ 5.000 x 0,95 = R$ 4.750
```

Esse preço é uma hipótese de lançamento, não autorização para vender abaixo do
custo. Antes de aceitar o projeto, a proposta precisa confirmar escopo,
elegibilidade da correção, horas, custos e piso sustentável. Se o piso calculado
for maior, o preço público não prevalece.

O modelo comercial vigente usa projeto único: contratação, Pipeline 1
integral, confirmação de escopo e, quando aplicável, Pipeline 2/correção
autorizada. A recorrência do Check permanece futura.

## Conta e pagamento

A autenticação social foi implementada com Better Auth 1.7.2. `/entrar` inicia
Google ou Apple, `/conta` consulta a sessão e permite encerrá-la, e o Worker
atende `/api/auth/*`. Usuários, contas, sessões, verificações OAuth e rate limit
ficam no D1 isolado `vexkeep-auth-production`. Cookies são seguros, HttpOnly e
SameSite=Lax; origem e redirects são validados; OAuth state e PKCE permanecem
ativos; tokens dos provedores são cifrados antes da gravação.

O Google só é exposto quando `GOOGLE_CLIENT_ID` e `GOOGLE_CLIENT_SECRET`
existem. A Apple exige `APPLE_CLIENT_ID`, `APPLE_TEAM_ID`, `APPLE_KEY_ID` e
`APPLE_PRIVATE_KEY`; o Worker gera o client secret ES256 dinamicamente com
validade máxima de 180 dias. O endpoint público de configuração retorna apenas
booleans e mantém o provedor incompleto desabilitado. Nenhum secret ou chave
privada entra no repositório.

O banco, a migração e `BETTER_AUTH_SECRET` já foram provisionados. A ativação
real dos provedores permanece pendente porque o Google Cloud exigiu nova
autenticação da conta e o Apple Developer não tinha sessão ativa disponível.
Os callbacks definidos são `https://vexkeep.com/api/auth/callback/google` e
`https://vexkeep.com/api/auth/callback/apple`. Conta própria por e-mail continua
fora desta iteração, pois exigiria envio transacional, verificação, recuperação
e controles adicionais que não são necessários para Google/Apple.

`/contratar` ainda não recebe pagamento.

`/termos` publica as condições de uso da verificação pública e separa
explicitamente uso do site, contratação e autorização técnica. A Política de
Privacidade identifica finalidades, bases, fornecedores, transferências,
retenção e canal de direitos. O formulário registra a confirmação de que a
pessoa pode solicitar a consulta e apresenta link direto para os Termos.

O pacote contratual canônico está em `negocio/juridico/`. Ele contém resumo
pré-contratual, proposta, contrato-mestre, SOW, autorização expressa, DPA,
confidencialidade, retenção, incidente, aceite, aditivo, encerramento, ROPA,
suboperadores e gate jurídico. Pagamento isolado não autoriza teste ativo.

O caminho sem conta deve ser tratado como **compra como convidado**, não como
pagamento anônimo. Mesmo sem perfil na Vexkeep, o provedor financeiro processa
os dados necessários para cobrança, antifraude, suporte e obrigações aplicáveis.

## Pendências antes de ativar as cascas

1. concluir sandbox do provedor de pagamento e definir webhook autenticado, idempotência,
   conciliação, reembolso e emissão fiscal;
2. concluir a reautenticação no Google Cloud e no Apple Developer, criar as
   credenciais com os redirects exatos e cadastrá-las como secrets do Worker;
3. manter autenticação por e-mail fora do MVP até os pilotos validarem a necessidade;
4. preencher identidade empresarial, fornecedores, transferências, prazos e
   foro; obter revisão jurídica e contábil independente do pacote em
   `negocio/juridico/` antes de habilitar checkout;
5. preparar o host Linux do Pipeline 1 para projetos contratados, sem ligar o
   runner à requisição gratuita do site;
6. manter a venda online bloqueada até oferecer resumo, correção de erros,
   cópia conservável do contrato, cancelamento e direitos aplicáveis.

A decisão auditada e o fechamento atualizado estão em
`docs/PLANO_CORRECAO_VEXKEEP_REVISADO_2026-08-26.md`.

## Pesquisa de mercado e referências

- [PentestAI, preço e prazo no Brasil](https://pentestai.com.br/): R$ 5.000 por
  teste, publicado em 2026. É referência comercial de concorrente, não fonte
  neutra de custo do mercado.
- [Cobalt Pricing](https://www.cobalt.io/platform/pricing): US$ 3.500 por teste
  autônomo na promoção vigente em 2026.
- [Intruder Pentest Pricing](https://www.intruder.io/pentest-pricing): US$ 4.000
  por teste único e US$ 3.500 para assinantes da plataforma.
- [HostedScan Pricing](https://hostedscan.com/pricing): planos recorrentes de
  US$ 39, US$ 109 e US$ 189 por mês para cinco alvos.
- [Beagle Security Pricing](https://beaglesecurity.com/pricing): planos
  recorrentes de US$ 99 e US$ 299 por mês no faturamento anual.
- [Better Auth, e-mail e senha](https://better-auth.com/docs/authentication/email-password),
  [Google](https://better-auth.com/docs/authentication/google) e
  [Apple](https://better-auth.com/docs/authentication/apple).
- [Google, criação do client ID](https://developers.google.com/identity/oauth2/web/guides/get-google-api-clientid)
  e [Apple, configuração web](https://developer.apple.com/documentation/signinwithapple/configuring-your-environment-for-sign-in-with-apple).
- [Stripe, clientes convidados](https://docs.stripe.com/payments/checkout/guest-customers):
  uma transação sem objeto Customer ainda mantém agrupamento e dados de
  pagamento para operação e fraude; portanto, não deve ser chamada de anônima.
