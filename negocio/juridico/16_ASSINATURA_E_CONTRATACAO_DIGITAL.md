# Assinatura eletrônica e contratação digital

Versão documental: 1.1, de 11/09/2026.

**Estado: procedimento proposto; provedor e integração não ativados.** Não
substitui revisão jurídica dos documentos preenchidos. Este arquivo não é
comprovante de assinatura nem autorização para executar um caso.

## 1. Decisão de desenho

Apresentar ao cliente uma jornada única de contratação, mantendo documentos
com funções distintas. O título comercial pode ser **Contrato de avaliação e
correção de segurança**, com escopo e autorização destacados. Não chamar toda
a contratação apenas de Termos de Uso.

| Etapa | Instrumento | O que comprova ou limita |
| --- | --- | --- |
| Navegação e prévia | Termos de Uso + aviso de privacidade | Condições de uso do site; não autoriza avaliação contratada. |
| Contratação | Resumo, proposta e contrato-mestre | Oferta, preço, entregáveis, responsabilidades e manifestação de vontade. |
| Autorização técnica | SOW + autorização expressa | Ativos, ações, limites, janela, representantes e poderes. |
| Tratamento de dados | DPA, retenção e fornecedores aplicáveis | Instruções e condições para dados do projeto; não é consentimento genérico. |
| Conclusão | Revisão técnica final + termo de entrega/aceite | Versão aprovada e entrega; não é garantia de ausência de falhas. |

A assinatura contratual ocorre antes do serviço. A revisão técnica humana
permanece somente no final do fluxo de correção, sem aprovação por ticket.

## 2. O que a lei permite e o que não se deve prometer

A MP 2.200-2 admite meios de comprovação de autoria e integridade além da
ICP-Brasil, nas condições do art. 10, § 2º. Isso não cria validade
incontestável, não demonstra poderes de representação e não convalida uma
cláusula inválida. [Texto oficial](https://www.planalto.gov.br/ccivil_03/mpv/antigas_2001/2200-2.htm).

O STJ reconheceu que a ausência de credenciamento ICP-Brasil, por si só, não
invalida a assinatura no caso analisado no REsp 2.159.442/PR. É fundamento
para admitir assinatura privada com prova adequada, não certificação de todos
os contratos de uma plataforma. [Decisão divulgada pelo STJ](https://www.stj.jus.br/sites/portalp/Paginas/Comunicacao/Noticias/2024/03122024-Falta-de-credenciamento-da-entidade-certificadora-na-ICP-Brasil--por-si-so--nao-invalida-assinatura-eletronica-.aspx).

Validade do contrato, força da prova e possibilidade de execução judicial são
questões distintas. O art. 784, § 4º, do CPC admite dispensa de testemunhas sob
suas condições; não torna qualquer obrigação automaticamente líquida, certa
e exigível. [CPC](https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2015/lei/l13105.htm).

Portanto, não usar as expressões "validade indiscutível", "impossível contestar",
"blindagem jurídica" ou "o cliente renuncia a questionar".

## 3. Provedor e autenticação

**Recomendação para avaliação do proprietário:** Clicksign como candidata
inicial, por oferecer envelopes com vários documentos, métodos de
autenticação e API com sandbox. Confirmar no plano contratado a exportação do
arquivo assinado, comprovante, eventos, fatores de autenticação e retenção.
Não houve contratação, criação de envelope ou transferência de dados.
[API oficial](https://developers.clicksign.com/v3.0/) e
[autenticação](https://www.clicksign.com/autenticacoes).

ZapSign e Docusign são alternativas a avaliar com os mesmos critérios, não
integrações simultâneas necessárias. Não foi fixado preço nem escolhido plano.
[ZapSign API](https://docs.zapsign.com.br/) e
[Docusign](https://www.docusign.com/pt-br/produtos/assinatura-eletronica/assinatura-digital).

Critérios de aceitação do provedor:

- exportar íntegra assinada e comprovante independente da assinatura ativa;
- identificar método, signatários, eventos e versões, com detecção de alteração;
- permitir autenticação proporcional ao risco e recusa de assinatura;
- fornecer controles de acesso, MFA administrativo, recuperação e revogação;
- informar contratos de tratamento, suboperadores, locais e retenção;
- suportar os documentos e o vínculo entre anexos sem truncá-los;
- permitir sandbox e verificação de conclusão no servidor quando integrado.

Para autorização técnica, escolher modalidade com comprovação adequada de
autoria e integridade. Preferir assinatura avançada efetivamente caracterizada
ou qualificada quando o risco, o cliente ou a análise jurídica justificarem.
Não declarar uma assinatura "avançada" só porque há OTP, IP ou login social.

Verificar separadamente os poderes: ato societário ou procuração pertinente,
representante, validade e, se necessário, titular do ativo. Evitar coletar
documento integral quando bastar comprovação proporcional. Biometria não é
requisito padrão do MVP nem deve ser ativada sem avaliação específica.

## 4. Experiência do cliente no site

1. **Solicitar proposta:** resultado público não confirma falha nem autoriza
   correção. Não cobrar automaticamente a partir da prévia.
2. **Identificar contratantes:** dados empresariais, signatários, poderes e
   contato. Compra sem conta continua identificada para contratação/cobrança.
3. **Revisar a proposta:** mostrar preço total, prazo, entregáveis, limites,
   condições para começar, cancelamento e tratamento de itens não corrigíveis.
4. **Conferir documentos:** oferecer download do pacote integral e anexos.
   Permitir correção de dados antes do envio para assinatura.
5. **Manifestar concordância:** controles não pré-marcados e redação específica:
   - "Li e concordo com o contrato e os anexos identificados nesta proposta."
   - "Tenho poderes para representar o contratante e autorizo somente os
     ativos, ações e limites descritos no escopo."
6. **Assinar:** botão "Revisar e assinar contrato" leva à sessão do provedor.
   As duas manifestações devem integrar o texto efetivamente assinado ou sua
   trilha verificável. Checkboxes do site, sozinhos, não liberam o serviço.
7. **Receber cópia:** disponibilizar arquivo final, anexos e comprovante para
   ambos os contratantes por canal seguro; confirmar a conclusão.
8. **Pagar e concluir onboarding:** cobrança conforme proposta, conciliação,
   comprovação de poderes/ativos, janela e preflight antes de executar.
9. **Atendimento e cancelamento:** canal acessível também sem login.

Ciência da Política de Privacidade não é autorização irrestrita para tratar
dados. Não condicionar o contrato a consentimento de marketing. Se este existir
futuramente, deve ser separado e opcional.

O site deve mostrar somente funções disponíveis. Esta especificação não cria
nova página publicada, botão funcional ou integração de assinatura.

## 5. Montagem sem circularidade de hashes

1. Congelar conteúdo e IDs/versões dos documentos e anexos. No mesmo pacote,
   referências internas usam ID, versão e seção; não o hash do futuro arquivo
   assinado que contém a própria referência.
2. Incluir escopo técnico legível e políticas já congeladas que não referenciem
   o próprio pacote. Não anexar credenciais, código ou evidência bruta ao
   provedor de assinatura.
3. Gerar o arquivo a apresentar, guardar seus bytes e SHA-256. Enviar ao
   provedor somente depois de conferir completude e ausência de placeholders.
4. Após todas as assinaturas, baixar e validar o arquivo final e comprovantes.
   Guardar outro SHA-256 para o arquivo final. A assinatura pode modificar os
   bytes do PDF; hashes pré e pós-assinatura não precisam ser iguais.
5. Vincular arquivo apresentado, ID do envelope, arquivo final, anexos e
   signatários no registro de fechamento. Se documentos forem assinados
   separadamente, listar todos, suas versões e hashes finais.
6. Só então apontar `authorization.sow_path` ao documento final assinado que
   contém o SOW e calcular `authorization.sow_sha256` dos seus bytes exatos.
7. Gerar a configuração operacional e registrar seu hash em manifesto externo.
   Conferir campo a campo ativos, métodos, taxas, janela e dados contra o SOW.
   O manifesto posterior não é uma nova autorização.
8. Mudança material exige nova versão e assinatura. Não editar o PDF assinado,
   alterar limites depois ou marcar `confirmed=true` só para passar no motor.

O motor atual verifica campos de autorização e o hash do arquivo, mas **não
valida por si só a assinatura do provedor nem os poderes do representante**.
Essa constatação vem de `pipeline/internal/app/pipeline2.go`. Até existir
integração validada, a conferência de contratação é operacional e registrada;
não deve ser anunciada como automatizada.

## 6. Dossiê mínimo privado de cada contratação

Guardar fora dos repositórios de código e com acesso restrito:

- identificação do contratante e comprovação proporcional de representação;
- proposta, resumo e pacote apresentados, com versão e hash;
- pacote final assinado, anexos e comprovante/trilha do provedor;
- ID do envelope, método, signatários, data/hora UTC e apresentação de fuso;
- resultado de validação de assinatura e responsável pela conferência;
- prova de ativos, autorizações de terceiros e limites;
- confirmação de pagamento e documento fiscal, sem dados de cartão;
- manifesto de vínculo com configuração e preflight;
- registro de envio da cópia, aditivos, cancelamento e entrega;
- prazo, fundamento e responsável pela retenção de cada categoria.

Hash sem preservação do artefato não permite reconstruir a prova. Backup deve
ser recuperável e expirar conforme a política. A conta da plataforma de
assinatura não pode ser a única cópia disponível.

## 7. Integração futura: critérios obrigatórios, não implementação concluída

- criar envelope no servidor, a partir de versão aprovada e íntegra;
- validar webhook pelo mecanismo oficial do provedor escolhido; deduplicar
  eventos, consultar estado autenticado e verificar todos os signatários;
- não confiar no retorno do navegador ou apenas no evento "documento visto";
- separar contrato assinado, pagamento, autorização vigente, onboarding e
  aprovação técnica final;
- cancelamento, recusa, expiração, documento trocado ou assinatura incompleta
  não podem liberar execução;
- servir documentos somente ao contratante correto, sem links públicos em
  logs e sem anexar material técnico desnecessário;
- mudanças materiais invalidam a liberação anterior até nova conferência.

## 8. Aceite antes de ativar

- [ ] provedor/plano, conta administrativa e tratamento de dados aprovados;
- [ ] documentos preenchidos e revisados, sem referências circulares;
- [ ] representação e ativos conferidos independentemente do login;
- [ ] assinatura de teste concluída e arquivo final validado;
- [ ] teste de recusa, expiração, adulteração e ausência de signatário bloqueia;
- [ ] download e recuperação do dossiê comprovados;
- [ ] nenhum evento de assinatura libera pagamento ou execução indevidamente;
- [ ] atendimento e cancelamento funcionando;
- [ ] registros coerentes com retenção e privacidade;
- [ ] pacote real aprovado para aquele contratante e escopo.

Todos permanecem pendentes até existir evidência. Revisão técnica final de
correção não substitui este aceite da contratação.

## 9. Preparação local disponível

O roteiro [VALIDACAO_LOCAL_MVP.md](../../docs/VALIDACAO_LOCAL_MVP.md) executa
checagens documentais e ensaios sintéticos dos bloqueios técnicos. O comando
`python3 pipeline/scripts/check_docs.py`, na raiz, verifica UTF-8, título e
destinos dos links Markdown do pacote canônico, sem ler dossiês de clientes.
Não confere cláusulas, placeholders preenchidos, poderes, assinatura ou
validade jurídica. Um resultado `passed` nunca altera o gate `14`.

Os testes técnicos exercitam onboarding sem autorização, revisão final e
integridade da entrega. Não simulam assinatura aceita nem pagamento recebido
para declarar contratação pronta. Os cenários de provedor da seção 8 continuam
dependendo da integração escolhida; o teste com o fornecedor deve preservar
documento apresentado, documento assinado e trilha recuperável.
