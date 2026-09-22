# Contrato-mestre de serviços de segurança da informação

**Modelo revisado em 11/09/2026 — não liberado para assinatura.** Preencher os
campos, confirmar a operação real e obter revisão jurídica individual. Esta
revisão não constitui parecer de advogado nem altera contrato já celebrado.

Pelo presente instrumento, de um lado, **`[CONTRATADA]`**, inscrita no CNPJ sob
`[CNPJ]`, com sede em `[ENDEREÇO]`, representada por `[REPRESENTANTE]`, doravante
**Contratada**; e, de outro, **`[CLIENTE]`**, inscrito(a) no CPF/CNPJ sob
`[NÚMERO]`, com endereço em `[ENDEREÇO]`, representado(a) por `[REPRESENTANTE]`,
doravante **Cliente**, resolvem celebrar este Contrato.

## 1. Objeto e documentos integrantes

1.1. A Contratada prestará serviços profissionais de avaliação, orientação,
correção e reteste de segurança de aplicações, conforme cada Statement of Work
(SOW) assinado.

1.2. Cada contratação é delimitada pela proposta, SOW, autorização expressa,
política de engajamento e, quando aplicável, Anexo de Proteção de Dados.

1.3. A execução é obrigação de meios. Relatórios, testes ou correções não são
certificação nem garantia de ausência de vulnerabilidades ou incidentes.

1.4. Aditivos identificam expressamente o que modificam. O SOW e a autorização
definem os limites técnicos; o Anexo de Proteção de Dados prevalece em matéria
de tratamento de dados. Este Contrato rege as condições gerais. Divergências
com a proposta ou entre anexos serão resolvidas antes da assinatura e execução,
sem ampliação tácita de autorização ou supressão de direitos obrigatórios.

## 2. Formação, assinatura e integridade

2.1. As partes reconhecem como válidas assinaturas eletrônicas admitidas entre
elas, desde que permitam identificar signatário, autoria, integridade, data e
trilha de auditoria.

2.2. Contrato, SOW, autorização e política de engajamento receberão versão e
SHA-256. Divergência, documento expirado ou assinatura ausente bloqueia a
execução.

2.3. O signatário do Cliente declara possuir poderes para contratar e autorizar
as atividades descritas. A Contratada verificará os poderes por documentação
proporcional ao risco. Controle de domínio ou acesso a uma conta de e-mail não
comprovam, isoladamente, poderes de representação da pessoa jurídica.

2.4. O contrato e seus anexos podem integrar um único pacote eletrônico,
identificado por ID, versão e relação de documentos. Todas as condições devem
estar disponíveis para leitura e download antes da assinatura, com destaque
para limites, preço, prazo e cancelamento. O ato de assinatura deve abranger
expressamente o SOW e a autorização técnica, não apenas os termos públicos.

2.5. As partes receberão o arquivo eletrônico final assinado e o comprovante
disponibilizado pelo provedor. Serão preservados a versão apresentada, a
versão assinada, os respectivos hashes e os registros de autoria, integridade
e manifestação de vontade. Login, Turnstile, hash ou imagem de assinatura
isolados não são apresentados como prova incontestável de contratação.

2.6. Assinatura eletrônica não implica renúncia ao direito de questionar fraude,
erro, autoria, integridade ou cláusula inválida. A modalidade e os fatores de
autenticação serão escolhidos conforme o risco, nos termos do procedimento
`16_ASSINATURA_E_CONTRATACAO_DIGITAL.md`.

2.7. Os hashes do arquivo final assinado serão calculados após a conclusão de
todas as assinaturas. Documentos que integram o mesmo pacote referenciam-se por
ID, versão e seção, sem dependência circular entre seus próprios hashes. A
configuração operacional gerada depois da assinatura não pode ampliar o
conteúdo autorizado.

## 3. Condições para início

3.1. Nenhuma avaliação contratada, inclusive reconhecimento integral, teste
ativo ou correção, começará antes de, cumulativamente:

a) assinatura do Contrato, SOW e autorização expressa;  
b) prova de titularidade ou autorização dos ativos e terceiros envolvidos;  
c) pagamento exigível confirmado;  
d) janela, contatos, limites e parada de emergência validados;  
e) acessos fornecidos por canal seguro; e  
f) preflight técnico e jurídico aprovado.

3.2. Pagamento, cadastro, e-mail ou conversa isolados não constituem autorização
técnica.

3.3. A prévia gratuita do site é serviço distinto, limitado pelos Termos de Uso.
Seu aceite não será reutilizado para autorizar a avaliação contratada.

## 4. Obrigações da Contratada

4.1. Executar somente o SOW autorizado, com diligência profissional, pessoal
qualificado, isolamento, limitação de taxa, registros de integridade e parada de
emergência.

4.2. Minimizar impacto, interromper atividade diante de risco material não
previsto e comunicar o contato de emergência.

4.3. Confirmar achados antes de apresentá-los como vulnerabilidade, distinguir
observação de conclusão e declarar limitações e falsos negativos possíveis.

4.4. Proteger credenciais e evidências; não inseri-las em repositórios, modelos
ou serviços não autorizados; e limitar acesso por necessidade.

4.5. Corrigir apenas questões de segurança contratadas, preservar
funcionalidades não relacionadas e manter a melhor versão aprovada quando uma
tentativa causar regressão de segurança.

4.5.1. A verificação de não regressão se limita à cobertura acordada e aos testes
executados, cujas lacunas serão declaradas. A Contratada não entregará candidato
reprovado pelos gates obrigatórios. Esgotado o limite de tentativas, comunicará
o resultado, as pendências e o tratamento comercial previamente acordado.

4.5.2. A revisão técnica humana ocorre ao final, antes da entrega. Assinatura
contratual, autorização e revisão técnica final têm funções distintas. Nenhuma
dessas etapas é substituída pela conclusão de um modelo de IA.

4.6. Entregar os artefatos acordados, comunicar incidentes nos termos do DPA e
encerrar acessos ao final.

## 5. Obrigações do Cliente

5.1. Fornecer informações verdadeiras, escopo completo, prova de autorização,
contas de teste, backups, contatos disponíveis e ambiente adequado.

5.2. Obter autorizações de provedores, titulares, parceiros e terceiros cujos
ativos possam ser afetados.

5.3. Não incluir dados reais desnecessários, credenciais compartilhadas ou
ativos de terceiros sem aviso e autorização documental.

5.4. Manter backup restaurável e plano de retorno antes de qualquer correção.

5.5. Avaliar e aprovar implantação. A Contratada não implantará em produção,
alterará dados ou realizará ação irreversível sem previsão específica no SOW.

5.6. Corrigir dependências e condições fora do escopo, responder dentro dos
prazos e não usar o relatório para afirmar certificação inexistente.

## 6. Escopo, mudanças e cooperação

6.1. Somente ativos, técnicas, identidades, taxas, horários e efeitos listados
no SOW são permitidos. Silêncio significa exclusão.

6.2. Descoberta relevante fora do escopo será registrada como limitação ou
candidato, sem exploração. Ampliação exige ordem de mudança e nova autorização.

6.3. Atrasos causados por acesso, indisponibilidade, mudança do sistema ou falta
de resposta do Cliente prorrogam os marcos na medida do impacto documentado.

## 7. Correções, propriedade intelectual e código

7.1. O Cliente preserva a titularidade de seu código, dados, marcas e materiais.

7.2. Após pagamento integral, os direitos patrimoniais sobre patches produzidos
especificamente para o Cliente e previstos no SOW são cedidos ao Cliente, sem
abranger ferramentas, métodos, bibliotecas, templates e conhecimento anterior
da Contratada.

7.3. Componentes de terceiros permanecem sujeitos às respectivas licenças. A
Contratada entregará identificação das dependências adicionadas ou alteradas.

7.4. A Contratada pode conservar conhecimento geral não confidencial e não
identificável, sem reutilizar código, dados ou evidências do Cliente.

7.5. Salvo SOW expresso, a entrega ocorre em branch, patch ou pacote separado;
deploy em produção e operação contínua não estão incluídos.

## 8. Aceite e suporte corretivo

8.1. O Cliente terá `[5]` dias úteis após a entrega para apontar, por escrito e
com evidência, desconformidade objetiva com os critérios de aceite do SOW.

8.2. Ausência de manifestação não elimina vício comprovado nem direito
obrigatório, mas permite o fechamento administrativo do projeto.

8.3. Durante `[PRAZO DE SUPORTE A APROVAR]`, a Contratada tratará sem cobrança
adicional desconformidade reproduzível imputável à sua entrega no escopo
contratado. Definirá com o Cliente a correção ou solução cabível, sem afastar
garantias e outros direitos obrigatórios. Mudança independente do sistema, novo
requisito ou falha preexistente não incluída depende de análise e proposta,
nunca de cobrança automática.

8.4. Achado não confirmado, mitigação, item sem correção viável e limitação de
cobertura serão identificados separadamente. Não serão declarados resolvidos
nem aceitos pelo Cliente por silêncio. O SOW deve prever o tratamento dessas
situações e os critérios de aceite antes do pagamento.

## 9. Preço, tributos, atraso e despesas

9.1. Preço, marcos, vencimentos, tributos e despesas constam da proposta/SOW.

9.2. A Contratada emitirá o documento fiscal aplicável. Tributos legalmente
retidos pelo Cliente devem ser comprovados.

9.3. Atraso autoriza suspensão segura de atividades ainda não executadas, após
notificação, sem retenção de credenciais ou dados como garantia.

9.4. Cancelamento sem culpa da Contratada obriga pagamento proporcional do
trabalho aceito e comprovadamente realizado e das despesas irreversíveis
previamente autorizadas, respeitada a legislação obrigatória.

9.5. Em caso de inadimplemento da Contratada, serão preservados os direitos de
reexecução, resolução, restituição e reparação aplicáveis. Valores antecipados
relativos a serviço não prestado serão apurados e restituídos conforme o
contrato e a lei, sem retenção automática integral ou multa não informada.

9.6. O procedimento `17_COBRANCA_ATENDIMENTO_E_CANCELAMENTO.md` registra os
canais, prazos aprovados, conciliação e restituição. Não se exige a criação de
conta para exercer cancelamento ou pedir atendimento. Eventual verificação de
identidade deve ser necessária e proporcional.

## 10. Direito de arrependimento e contratação eletrônica

10.1. Quando o Código de Defesa do Consumidor for aplicável, prevalecem seus
direitos obrigatórios, inclusive informação clara, atendimento e arrependimento
em contratação fora do estabelecimento.

10.2. A Contratada não exigirá renúncia genérica ao direito de arrependimento.
Início antecipado do serviço dependerá de solicitação destacada do Cliente e de
procedimento juridicamente validado, sem afastar direito obrigatório.

10.3. Em relação empresarial paritária, cancelamento seguirá este Contrato e o
SOW, com apuração proporcional e documentada.

## 11. Confidencialidade

11.1. Informação não pública recebida em razão do projeto é confidencial,
independentemente de marcação quando sua natureza indicar sigilo.

11.2. Cada parte usará a informação apenas para o Contrato, limitará acesso e
aplicará proteção compatível com sua sensibilidade.

11.3. Não são confidenciais informações comprovadamente públicas sem violação,
já conhecidas legitimamente, recebidas de terceiro autorizado ou desenvolvidas
de forma independente.

11.4. Divulgação legalmente exigida será limitada e, quando permitido,
precedida de aviso. Vulnerabilidades não podem ser divulgadas publicamente sem
autorização escrita do Cliente.

11.5. O sigilo permanece por `[5]` anos; credenciais, dados pessoais e segredos
industriais permanecem protegidos enquanto sua natureza ou a lei exigir.

## 12. Proteção de dados

12.1. Os papéis de controlador e operador são definidos pela atuação real e
pelo DPA, não apenas pelo nome dado pelas partes.

12.2. Quando atuar como operadora, a Contratada tratará dados pessoais somente
por instruções documentadas do Cliente, comunicará instrução aparentemente
ilícita e auxiliará no atendimento de direitos, incidentes e término.

12.3. Suboperadores, transferências, medidas de segurança, categorias de dados
e prazos constam do DPA e registros anexos.

12.4. Cada parte responde por suas decisões próprias como controladora.

## 13. Segurança e incidentes

13.1. As partes manterão medidas técnicas e administrativas proporcionais ao
risco. A Contratada notificará o Cliente sem demora indevida e, como meta
contratual, em até `[24]` horas após confirmar incidente que afete dados ou
ativos do projeto, com atualizações à medida que fatos forem verificados.

13.2. O controlador legalmente responsável decide e realiza comunicações à ANPD
e aos titulares, com assistência da outra parte. Nenhuma parte fará comunicação
em nome da outra sem autorização ou dever legal próprio.

13.3. Evidências de incidente serão preservadas de forma proporcional, sem
alterar os prazos legais aplicáveis.

## 14. Garantias e limitações técnicas

14.1. A Contratada garante aderência ao SOW, não resultado absoluto. Testes são
amostrais, dependem de tempo, acessos, versão e comportamento do ambiente.

14.2. A correção reduz o risco do item tratado; não garante ausência de falhas,
compatibilidade com mudanças futuras ou segurança de terceiros.

14.3. Serviços externos, indisponibilidade de provedor, mudança do Cliente e
ações de terceiros não estão sob controle da Contratada.

## 15. Responsabilidade

15.1. Cada parte responde por danos diretos comprovados decorrentes de seu
inadimplemento, na proporção de sua responsabilidade e observada a lei.

15.2. Qualquer teto de responsabilidade dependerá de avaliação jurídica da
relação concreta e estipulação destacada no SOW: `[VALOR/CRITÉRIO, JUSTIFICATIVA,
EXCEÇÕES E APROVAÇÃO]`. Este modelo não presume que limitar ao valor pago seja
adequado. Campo não preenchido não cria teto e impede a liberação da minuta
para assinatura até decisão expressa sobre adoção ou não da limitação.

15.3. Eventual limite não se aplica onde a lei impedir nem a dolo, e deverá
tratar expressamente culpa grave, sigilo, dados pessoais, perda de dados e
direitos de terceiros conforme a lei e a revisão jurídica do caso. Não vincula
titulares de dados ou terceiros que não sejam partes do Contrato.

15.4. Não há neste modelo exclusão automática de perda de dados ou de toda
espécie de dano. Nexo causal, extensão e responsabilidade serão apurados
conforme a lei e os termos validamente pactuados. Uso de ferramentas, IA,
suboperadores ou backup do Cliente não afasta dever próprio da Contratada.

15.5. Esta cláusula deve ser revista para seguro, porte, setor e classificação
do Cliente. Em relação de consumo, aplicam-se integralmente as proteções
obrigatórias e não se presume validade de limitação.

## 16. Vigência, suspensão e rescisão

16.1. O Contrato vigora por `[12]` meses e renova-se somente por escrito; cada
SOW conserva seus efeitos até encerramento.

16.2. A parte inocente pode rescindir por descumprimento material não sanado em
`[10]` dias após notificação, salvo urgência, ilicitude, risco de segurança ou
quebra de confiança que justifique suspensão imediata.

16.3. A Contratada pode suspender imediatamente atividade que exceda a
autorização, represente risco material, viole lei ou atinja terceiro não
autorizado.

16.4. Rescisão não elimina pagamento devido, confidencialidade, propriedade,
proteção de dados, responsabilidade, retenção probatória e encerramento.

## 17. Força maior

Nenhuma parte responde por atraso inevitável causado por evento fora de seu
controle razoável, desde que comunique, mitigue e retome a obrigação. Força
maior não justifica descumprir deveres de sigilo, segurança ou pagamento já
vencido.

## 18. Comunicações

Notificações contratuais serão enviadas aos endereços abaixo com confirmação de
recebimento. Incidentes e parada usam os contatos do SOW.

- Contratada: `[E-MAIL]`;
- Cliente: `[E-MAIL]`.

## 19. Disposições gerais

19.1. Tolerância não é renúncia. Invalidade parcial não invalida o restante.

19.2. Cessão do Contrato depende de consentimento escrito, salvo reorganização
sem redução de garantias e quando permitido por lei.

19.3. As partes são independentes; não há sociedade, emprego ou representação.

19.4. Alterações exigem instrumento escrito e assinado.

## 20. Lei e solução de conflitos

20.1. Aplica-se a lei brasileira.

20.2. As partes tentarão solução executiva por 15 dias antes de litigar, sem
impedir acesso à Justiça, reclamação administrativa, preservação de prazo ou
medidas urgentes. A tentativa não é condição obrigatória para exercer direitos.

20.3. Para relação empresarial paritária, fica eleito o foro de `[CIDADE/UF]`.
Quando houver relação de consumo ou competência obrigatória, prevalece o foro
assegurado por lei.

**Contratada:** `[NOME, CARGO, ASSINATURA, DATA E FUSO]`  
**Cliente:** `[NOME, CARGO, ASSINATURA, DATA E FUSO]`  
**Testemunha 1:** `[NOME, CPF, ASSINATURA]`  
**Testemunha 2:** `[NOME, CPF, ASSINATURA]`

Os campos de testemunhas são usados quando a modalidade e a avaliação jurídica
os exigirem ou recomendarem. Não são tratados como requisito universal de
validade. A possibilidade do art. 784, § 4º, do CPC depende de seus requisitos
e não torna toda obrigação contratual automaticamente executável.
