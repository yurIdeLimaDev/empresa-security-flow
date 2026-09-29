# Pipeline 0.5 web

Implementado em agosto; filtros IPv4/IPv6 e falha fechada de Turnstile
reforçados e testados localmente em 13/09/2026. [Estado atual](ESTADO_ATUAL.md).

## Função no produto

```text
URL autorizada -> Pipeline 0.5 gratuito -> resultado efêmero -> proposta
proposta aceita -> contrato e autorização -> Pipeline 1 integral
Pipeline 1 -> onboarding -> Pipeline 2 quando aplicável
```

O Pipeline 0.5 é uma leitura pública ampliada executada no Cloudflare Worker da
landing. Ele existe para entregar valor verdadeiro antes da contratação sem
expor o runner Linux ou transformar o navegador do visitante em ferramenta de
scan. Não é Pipeline 1, pentest, score ou garantia de segurança.

## Cobertura

- DNS A, AAAA, CNAME, NS, MX, TXT, CAA e DNSKEY;
- SPF e DMARC, incluindo classificação da política DMARC;
- HTTP/HTTPS por HEAD e GET limitado, sem seguir redirects;
- HSTS, CSP, X-Content-Type-Options, Referrer-Policy, Permissions-Policy,
  COOP, CORP e X-Frame-Options;
- sinais amplos de CSP e atributos agregados de cookies sem expor valores;
- `/.well-known/security.txt`, somente presença de `Contact`;
- até quatro subdomínios candidatos vindos de Certificate Transparency,
  restritos ao domínio informado, com HEAD HTTPS;
- fingerprint textual conservador em cabeçalhos, HTML e até dois scripts do
  mesmo host: Cloudflare, nginx, Apache, Express, PHP, Next.js, React, Vue,
  Angular, Svelte, WordPress e Supabase.

## Limites de execução

- no máximo 46 subrequisições externas por execução;
- HTML: 64 KiB; cada script: 96 KiB; CT: 96 KiB; tudo lido por streaming;
- no máximo quatro candidatos de CT, dois scripts e oito observações;
- apenas HTTP/S nas portas 80/443 e hostnames DNS públicos;
- resolução A/AAAA repetida antes de conexões ao alvo;
- bloqueio de IP literal e faixas privadas, reservadas e de documentação;
- `redirect: manual`, timeout por operação, Turnstile, origem same-origin,
  corpo JSON de até 4 KiB e WAF por IP.

## O que não faz

- login, credencial, formulário, mutação, exploração ou bypass;
- varredura de portas, crawling amplo ou execução do JavaScript do alvo;
- HIBP, subfinder, amass, httpx, testssl, nuclei ou qualquer container;
- persistência do domínio, resultado ou saída em D1;
- alegação de vulnerabilidade a partir de ausência de sinal público.

## Resultado

A resposta contém doze resumos de cobertura e até oito observações. Cada
observação separa evidência, significado e limite. A interface mostra também o
que foi e o que não foi coberto. Não há score. Uma oferta comercial aparece
depois do resultado e deixa explícito que o Pipeline 1 integral só ocorre após
contratação e autorização.

## Decisão sobre Pipeline 1 e Check

Pipeline 1 não é gratuito e não é executado pelo Worker. A implementação futura
do Vexkeep Check foi preservada no código e no D1 de staging, porém todas as
rotas públicas, rotas internas e o scheduler ficam desabilitados sem a variável
explícita `CHECK_PUBLIC_ENABLED=true`. Produção não define essa variável.

## Fontes técnicas consultadas

- Cloudflare Workers limits: até 50 subrequisições externas no plano Free e
  memória de 128 MB;
- Cloudflare Workers best practices: bindings para configuração, ausência de
  estado mutável global por requisição, promises aguardadas e Web Crypto;
- Cloudflare `waitUntil`: até 30 segundos após a resposta, inadequado para
  trabalho do qual a resposta depende;
- Certificate Transparency via `crt.sh` é fonte de candidatos, não prova de
  propriedade, disponibilidade ou falha.

Referências oficiais:

- <https://developers.cloudflare.com/workers/platform/limits/>
- <https://developers.cloudflare.com/workers/runtime-apis/context/>
- <https://developers.cloudflare.com/workers/runtime-apis/headers/>
