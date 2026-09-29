# Validação do Pipeline 0.5 web

Data: 26 de agosto de 2026  
Resultado: aprovado para a oferta pública limitada descrita na documentação.

## Escopo validado

- build estático React/Vinext e geração do Worker;
- API efêmera `POST /api/scan`;
- budget de 46 subrequisições e limites de bytes;
- DNS público, HTTP/HTTPS, cabeçalhos, CSP, cookies agregados, e-mail do
  domínio, security.txt, CT limitada e fingerprint textual;
- bloqueio de rede privada e escopo estrito de candidatos CT;
- retirada pública do Pipeline 1, Check e scheduler;
- conteúdo, privacidade, contratação e página de resultados;
- deploy e leitura em `https://vexkeep.com`.

## Comandos e resultados

```text
npm test
11 testes aprovados; 0 falhas

npm run lint
aprovado

npx tsc --noEmit
aprovado

npm audit --omit=dev --audit-level=high
0 vulnerabilidades

npx wrangler deploy --dry-run --config wrangler.vexkeep.jsonc
upload validado: 296,48 KiB; gzip 122,58 KiB; sem bindings declarados no arquivo

git diff --check -- README.md docs landing-page
sem erros de whitespace; somente avisos de conversão LF/CRLF do Git no Windows
```

O `TURNSTILE_SECRET` permanece como binding secreto já existente no Worker e
foi conferido após a publicação. Nenhum segredo foi gravado no repositório.

## Testes de produção

- `GET https://vexkeep.com/`: HTTP 200;
- título: `Verificação gratuita da superfície pública | Vexkeep`;
- HSTS e CSP: presentes;
- `POST /api/check/cases`: HTTP 404;
- `POST /api/scan` com token Turnstile inválido: HTTP 403;
- interface publicada inspecionada no navegador, com o novo fluxo e sem ação
  pública de Pipeline 1/Check;
- binding `TURNSTILE_SECRET`: preservado;
- observabilidade: ativa, amostragem 1, query string redigida;
- versão Worker: `ad0601ce-2b1c-41bc-9f7e-3de176cfc6ba`;
- deployment: `7cf8d4d1-b878-4935-add9-68a6a361a4cf`, 100% da versão acima.

## Integridade local

```text
landing-page/dist/cloudflare/vexkeep-landing-worker.mjs
SHA-256 F2DAD29FD6D0A081A2202AB3859304896BE32FEB958CA1AC7A02664572437BEA

landing-page/worker/runtime-template.mjs
SHA-256 06F999C5068E3EA1D73F3C1E0692289583020DEDC1302221A130E6AC36E07901

landing-page/public/app.js
SHA-256 37F3DBDAC71621659A8E383CEBC3CB2203413CE021D029DFD80F0D0C2B68CFE7
```

## Conclusão e limite honesto

A implementação é confiável para uma leitura gratuita, pública e de baixo
impacto. Ela não é confiável como substituta do Pipeline 1, pentest ou prova de
ausência de falhas. O caminho de sucesso completo em produção depende de um
token Turnstile resolvido por uma pessoa; por política, a automação não resolveu
o desafio. A lógica de sucesso foi coberta por teste de integração do Worker
com respostas externas controladas e budget conferido.

O Pipeline 1 integral permanece preservado no motor e deve ser executado apenas
depois da contratação, autorização e configuração do engagement.
