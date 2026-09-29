# Fechamento do ponto 3: contratação segura do MVP

Data: 11/09/2026. **Estado: parcialmente preparado, não concluído.**

Complemento local da mesma data: o procedimento `16` agora referencia a
verificação mecânica dos modelos, e o checklist `14` explicita o controle de
envio de código ao futuro gerador. O pacote permanece em revisão; não houve
assinatura, checkout ou aprovação profissional. Ver
[resultado local](../../validacao/2026-09-11-fechamento-local/RESULTADO.md).

## O que foi feito nesta revisão

- Reutilizados os modelos existentes, sem criar outro contrato concorrente.
- Aprimorados formação eletrônica, representação, escopo, tratamento de falha
  de correção, suporte, restituição e responsabilidade de ambas as partes.
- Retiradas da minuta as regras automáticas de teto igual ao valor pago e de
  exclusão geral de perda de dados; decisão final depende do risco e revisão.
- Separados termos do site, contrato pago, autorização técnica e revisão final.
- Corrigido o desenho de referências circulares de hash entre SOW, autorização
  e configuração operacional.
- Preparados procedimentos de assinatura, montagem do dossiê, cobrança,
  atendimento, cancelamento e reembolso.
- Registradas a avaliação preliminar de provedor de assinatura e as condições
  que uma integração real precisa cumprir.
- Atualizados os índices e checklists para não representar rascunho como gate
  automático ou conformidade concluída.

## Informações que dependem do proprietário

Responder apenas o que já existir; "ainda não" é uma resposta válida.

| Grupo | Informação necessária | Por que é necessária |
| --- | --- | --- |
| Prestadora | existência de CNPJ; razão social, endereço empresarial, município/UF e representante | Identificar quem contrata; não presumir entidade ou município. |
| Clientes | somente empresas ou também pessoas físicas; segmentos e países | Adaptar oferta, documentação, consumo e risco; CNPJ não exclui CDC automaticamente. |
| Profissionais | contador e advogado já contratados ou não; orientação já obtida | Confirmar fiscal, responsabilidade, direitos e adequação do pacote. |
| Cobrança | conta de recebimento em nome de quem presta; provedor já escolhido ou não | Definir conciliação, reembolso e nota; sem pedir senha ou dados sensíveis. |
| Atendimento | e-mail funcional, responsável, dias/horários e emergência | Não prometer capacidade ou SLA que não existe. |
| Assinatura | provedor/conta existente; concordância com avaliação da Clicksign ou preferência | Fechar plano, autenticação, tratamento de dados e sandbox. Não houve compra. |

Não publicar endereço pessoal, documento de identidade, CPF ou dados bancários
no repositório de modelos. Documentos preenchidos e evidências ficam em área
privada separada, com acesso e retenção definidos.

## Decisões profissionais ainda necessárias

1. enquadramento e emissão fiscal da prestadora;
2. público efetivo e incidência das normas de consumo;
3. eventual teto de responsabilidade, exceções, seguro e foro;
4. prazos de aceite, suporte, atendimento, restituição e início do trabalho;
5. tratamento de dados, fornecedores e transferências efetivas;
6. retenção por categoria, incluindo prova contratual e registros de acesso;
7. modalidade e prova de assinatura proporcionais ao serviço e ao risco;
8. autorização/recuperação para alterações em produção, se vierem a ser vendidas.

## O que não foi feito nem pode ser considerado pronto

- não foi constituída empresa, aberta conta ou contratado contador/advogado;
- não houve revisão por profissional habilitado ou assinatura de partes;
- não foi contratado provedor de assinatura nem criado envelope real;
- não existe nova integração funcional de assinatura/checkout nesta revisão;
- não foi emitida cobrança, nota fiscal, mensagem a cliente ou reembolso;
- não foi habilitado serviço pago, feito deploy ou alterado o runner;
- não foram testados com cliente os canais, assinatura e rotinas operacionais;
- fornecedores, prazos e dados empresariais não foram inventados.

## Ordem para concluir depois das respostas

1. confirmar prestadora, público, profissionais e canais;
2. preencher o pacote e obter decisões jurídicas/contábeis reais;
3. selecionar e aprovar conta/plano do provedor de assinatura;
4. montar pacote de teste com identidade fictícia no sandbox, sem apresentação
   como contrato real, e testar assinatura, recusa, cópia e adulteração;
5. validar cobrança, restituição, fiscal e atendimento;
6. implementar e testar integração no site se houver conta/API e autorização
   de publicação, mantendo checkout bloqueado até o aceite completo;
7. liberar somente o escopo e a operação comprovados no checklist `14`.

Produzir uma minuta melhor reduz lacunas, mas não substitui prestação real de
serviço jurídico nem garante que litígios não ocorrerão. A conclusão exige
documentos válidos para as partes, prova da contratação e controles praticados.
