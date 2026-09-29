# Login Google e navegação

Data: 16/09/2026. Histórico das etapas locais e da publicação posterior.
Estado mais recente: seção Publicação via plugin Cloudflare, ao final.
Não houve push para GitHub nesta tarefa.

- Apple removido da tela de acesso e das mensagens de disponibilidade dessa
  tela; backend, credenciais e contas existentes não foram excluídos.
- Botão Entrar no cabeçalho, separado do CTA principal e mantido no mobile.
- Texto de login não promete acompanhamento de verificações ainda não ativo.
- Client ID fornecido pelo titular configurado no Worker como variável não
  secreta. Callback esperado: `https://vexkeep.com/api/auth/callback/google`.
- Revisão do identificador de cache de app.js para publicar a nova mensagem.
- Guia de configuração segura e documentação atualizados.

Lint, TypeScript, build e 18 testes passaram. O teste novo confirma que Google
não depende de credenciais Apple e exige banco/segredo de sessão/configuração.
Um teste inicial confundiu a fonte Apple Color Emoji do framework com o botão;
a asserção foi restringida ao conteúdo da página e retestada, sem relaxar a
proibição do botão Apple. Não houve conferência visual em navegador nem login
real com o Google. As credenciais do teste são fictícias.

Faltam cadastro seguro de GOOGLE_CLIENT_SECRET, conferência do cliente Web,
URI de retorno e público do app no Google, conferência do D1/migração e segredo
de sessão remoto, publicação controlada e teste de login/logout com o titular.
Não enviar segredo ou JSON de credenciais no chat. Client ID não é senha.

## Complemento: cabeçalho fixo e tentativa de ativação

- Cabeçalho agora fixo no topo, com compensação de espaço no conteúdo e nas
  âncoras. Entrar fica à direita, em destaque vermelho; Verificar site fica
  secundário. O destaque visual não ativa assinatura nem altera a oferta.
- Conferência visual local em desktop e mobile realizada nesta continuação.
  Em viewport de 320 px, o documento tem 305 px úteis e nenhuma rolagem
  horizontal; o botão termina em 290 px. Corrigido o pequeno corte inicial.
- Depois de rolar 740 px, o cabeçalho permanece em posição fixed e top 0.
  Entrar abre /entrar, que mostra somente Google. Sem secrets locais, o botão
  permanece desabilitado, como esperado, sem simular autenticação.
- Lint e TypeScript passaram. Build e 18 testes passaram novamente após o
  ajuste final para telas estreitas. Isso não comprova OAuth em produção.
- Wrangler sem autenticação; painel Cloudflare exige login. Console Google
  exige confirmação de identidade. Abas preservadas para o titular entrar.
- Nenhum deploy, alteração de DNS, encaminhamento de e-mail, cadastro de
  secret ou teste de login real foi realizado nesta etapa.
- Encaminhamento solicitado para contato@vexkeep.com continua pendente de
  acesso à conta, conferência dos registros de e-mail existentes e eventual
  verificação do destino indicado pelo titular. Não substituir MX existentes
  sem avaliar possível interrupção do serviço atual. Email Routing encaminha
  recebimento; envio como contato@vexkeep.com é configuração separada.

## Publicação via plugin Cloudflare

Após conexão do plugin, a conta com a zona vexkeep.com e os domínios do Worker
foram confirmados. O bloqueio de autenticação Cloudflare da etapa anterior foi
resolvido pelo conector, sem copiar credenciais para o terminal ou repositório.

- Publicação em 16/09/2026 às 12:35:54 UTC (09:35:54 em Brasília).
- Worker: vexkeep-landing; versão: `659d1352-7104-4828-82ba-180c3ce12e2a`.
- Deployment: `e3e6f0d2-8d7b-4549-a681-340852d710af`, 100% da versão nova.
- Versão anterior para rollback: `606a48b7-e4c5-4490-b331-bdd6c02aab4f`.
- Bundle: 2.266.611 bytes. SHA-256:
  `13ad682818e2d1b2f478e45551efde796febd87e81d60a9fd347dcd0854ecbd6`.
- Wrangler 4.131.1 fez o dry-run local. O bundle foi transportado comprimido,
  descomprimido e conferido por tamanho e SHA-256 antes do upload pelo plugin.
  Não foi criado endpoint de transferência nem exposto código em serviço público.
- Banco AUTH_DB preservado; cinco tabelas de autenticação compatíveis com a
  migração local consultadas somente no esquema, sem dados pessoais.
- BETTER_AUTH_SECRET e TURNSTILE_SECRET preservados por keep_bindings;
  GOOGLE_CLIENT_ID publicado. Nenhum segredo foi exposto ou rotacionado.
- Ativada redação de query strings nos logs, também registrada no wrangler.
- Home e /entrar responderam 200 com a versão nova; cabeçalho e ausência do
  botão Apple confirmados por HTTP e navegador. A primeira leitura logo após
  upload ainda mostrou a versão antiga; a nova foi confirmada depois, inclusive
  por navegação normal sem parâmetro de atualização.
- /api/auth/vexkeep-configuration permanece google=false e apple=false;
  /api/check/config permanece 404. Não houve login real nem scan de terceiro.

### E-mail e Google: pendências reais

Email Routing já estava ativo e pronto. A regra contato@vexkeep.com encaminha
para o destino antigo verificado. O novo Gmail indicado pelo titular foi
cadastrado (identificador `d882508e02104b02a7669567806e20df`), mas a API retorna
unverified. O titular deve confirmar a mensagem da Cloudflare no Gmail; depois
disso a regra existente deve ser atualizada, preservando endereço, prioridade
e demais configurações. Nenhum MX foi removido ou substituído.

Google Cloud ainda apresenta Confirme que é você. Falta acesso ao cliente Web
para conferir callback/público e cadastrar o segredo correspondente com segurança.
Não basta o Client ID. A interface permanece indisponível até essa configuração
e ainda exigirá homologação de login/logout com uma conta controlada.

Referências usadas: skills Cloudflare e Wrangler para preservar bindings,
validar a publicação e respeitar a verificação do destino; consulta à API atual
da Cloudflare para schemas de upload, configurações e Email Routing.

## Continuação: console Google autenticado

- Titular confirmou usar jozemarfilho08@gmail.com (não a conta gbaigbag).
- Projeto login-508803 contém o cliente Web Vexkeep com o Client ID já publicado.
- Origem https://vexkeep.com existente; nenhum callback cadastrado. Endereço
  https://vexkeep.com/api/auth/callback/google preenchido em formulário, sem
  salvar enquanto a confirmação de alteração de acesso está pendente.
- Google informa que o segredo existente está ativo, mas não pode mais ser
  visualizado ou baixado. Perguntado se o titular guardou o segredo; não foi
  criada, excluída nem substituída uma chave.
- Salvos no branding os links https://vexkeep.com,
  https://vexkeep.com/privacidade e https://vexkeep.com/termos. Console mostrou
  As mudanças de marca foram salvas; HTTP 200 nos três endereços.
- Nome, domínio e contatos existentes preservados. Sem upload de nova logo.
- Público observado: externo, Testando, zero usuários de teste. Console não
  lista escopos não confidenciais, confidenciais ou restritos. Sem alteração
  de permissões nem publicação do app nesta continuação.
- Login real continua não homologado. As abas de trabalho foram preservadas;
  é preciso retomar do estado observado, não presumir que um rascunho foi salvo.

### Tentativa de salvar o callback após autorização

O titular confirmou que guardou o Client Secret e autorizou salvar o callback.
Ao clicar em Salvar, o Google retornou Falha ao salvar, rastreamento
c2081063588358354. Após fechar a mensagem, recarregar e usar Tentar novamente,
a página também apresentou Falha ao carregar (c7540332245891141 e
c5293157703069956). A persistência não foi confirmada; não reportar sucesso.
Não houve criação ou rotação de chave. Falta o titular indicar o arquivo local
do segredo ou cadastrá-lo diretamente na Cloudflare, fora do chat e do Git.

### Continuação: segredo recebido e callback salvo

- Titular forneceu arquivo local com o segredo existente. Transferência direta
  para GOOGLE_CLIENT_SECRET no Worker via plugin Cloudflare, resposta HTTP 201.
  Valor não registrado neste relatório nem copiado para o repositório.
- Segredos anteriores preservados; nenhuma chave Google criada ou rotacionada.
- Nova gravação do callback concluída; console confirmou Cliente OAuth salvo.
- API pública de configuração retornou google=true e apple=false.
- Botão Google iniciou o fluxo com callback correto, state e PKCE S256;
  escopos limitados a email, profile e openid. Seletor de contas abriu.
- Selecionada a conta autorizada pelo titular. Google pediu confirmação de
  identidade em seu dispositivo. Nenhum código de autenticação registrado aqui.
- Ainda pendentes: confirmação pessoal, eventual consentimento, retorno à
  conta, persistência de sessão, logout e avaliação de publicação do app Google.
  Configuração presente não equivale a login completo homologado.

### Retentativa de login em 22/09/2026

Após a primeira solicitação expirar, o login foi reiniciado a partir de
`/entrar` com a conta confirmada. O titular aprovou a confirmação de identidade
no próprio dispositivo. O Google apresentou a tela de consentimento para nome,
foto do perfil e e-mail. Nenhum consentimento foi aceito ainda; retorno à
Vexkeep, sessão e logout não foram testados.
