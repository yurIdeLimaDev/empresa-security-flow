# Bases jurídicas e limites do pacote

Revisão documental e consulta a fontes oficiais: 11/09/2026. Não houve parecer
de advogado, definição fiscal individual ou certificação de conformidade.

## Fontes oficiais consideradas

- [LGPD — Lei nº 13.709/2018](https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2018/lei/l13709compilado.htm): bases legais, transparência, término, direitos, agentes, segurança e incidentes;
- [Código Civil — Lei nº 10.406/2002](https://www.planalto.gov.br/ccivil_03/leis/2002/l10406compilada.htm): liberdade contratual empresarial, alocação de riscos, boa-fé, inadimplemento e contratos de adesão;
- [Código de Defesa do Consumidor](https://www.planalto.gov.br/ccivil_03/leis/l8078compilado.htm): informação, oferta, arrependimento e nulidade de cláusulas abusivas;
- [Decreto nº 7.962/2013](https://www.planalto.gov.br/ccivil_03/_ato2011-2014/2013/decreto/d7962.htm): contratação eletrônica, resumo, correção de erros, cópia do contrato e atendimento;
- [Código Penal, art. 154-A](https://www.planalto.gov.br/ccivil_03/decreto-lei/del2848compilado.htm): relevância da autorização para acesso/testes;
- [MP nº 2.200-2/2001](https://www.planalto.gov.br/ccivil_03/mpv/antigas_2001/2200-2.htm) e [Lei nº 14.063/2020](https://www.planalto.gov.br/ccivil_03/_ato2019-2022/2020/lei/l14063.htm): documentos e assinaturas eletrônicas;
- [Lei do Software — Lei nº 9.609/1998](https://www.planalto.gov.br/ccivil_03/leis/l9609.htm): direitos sobre software e contratação;
- [Resolução CD/ANPD nº 2/2022](https://www.gov.br/anpd/pt-br/acesso-a-informacao/institucional/atos-normativos/regulamentacoes_anpd/resolucao-cd-anpd-no-2-de-27-de-janeiro-de-2022): agentes de pequeno porte;
- [Resolução CD/ANPD nº 15/2024](https://www.gov.br/anpd/pt-br/assuntos/noticias/anpd-aprova-o-regulamento-de-comunicacao-de-incidente-de-seguranca): comunicação e registro de incidentes;
- [Resolução CD/ANPD nº 19/2024](https://www.gov.br/anpd/pt-br/acesso-a-informacao/institucional/atos-normativos/regulamentacoes_anpd/resolucao-cd-anpd-no-19-de-23-de-agosto-de-2024): transferências internacionais e cláusulas-padrão;
- [Guia de agentes de tratamento da ANPD](https://www.gov.br/anpd/pt-br/centrais-de-conteudo/materiais-educativos-e-publicacoes/anonimizado___guia_de_agente_de_tratamento_e_encarregado_da_anpd_novo.pdf);
- [Guia de segurança para agentes de pequeno porte](https://www.gov.br/anpd/pt-br/centrais-de-conteudo/materiais-educativos-e-publicacoes/processo-guia-orientativo-sobre-seguranca-da-informacao-para-agentes-de-tratamento-de-pequeno-porte.pdf).

## Decisões prudenciais adotadas

1. autorização é seção/documento específico e temporário; pode integrar o mesmo
   pacote assinado do SOW, preservando vínculos sem circularidade de hashes;
2. pagamento não equivale a autorização;
3. consumidor não renuncia genericamente a direitos obrigatórios;
4. limite de responsabilidade contém exceções e precisa de revisão para o caso;
5. papéis LGPD são funcionais, por operação;
6. transferência internacional não é declarada regular sem verificar o
   provedor e o mecanismo real;
7. correção não inclui melhoria funcional e não gera garantia absoluta;
8. checkout permanece desativado sem identidade empresarial e fluxo de aceite.

## Fundamentação adicional da revisão de 11/09/2026

- MP 2.200-2, art. 10: outros meios de comprovação de autoria/integridade são
  admissíveis nas condições legais; não existe promessa de incontestabilidade.
- [CPC, art. 784, § 4º](https://www.planalto.gov.br/ccivil_03/_ato2015-2018/2015/lei/l13105.htm): dispensa de testemunhas em hipóteses previstas não equivale a execução automática de toda cláusula.
- [STJ, REsp 2.159.442/PR](https://www.stj.jus.br/sites/portalp/Paginas/Comunicacao/Noticias/2024/03122024-Falta-de-credenciamento-da-entidade-certificadora-na-ICP-Brasil--por-si-so--nao-invalida-assinatura-eletronica-.aspx): ausência de credenciamento ICP-Brasil não invalida por si só a assinatura no contexto julgado. Não é chancela geral de um fornecedor ou do contrato Vexkeep.
- CDC, arts. 46, 49 e 51: conhecimento prévio, arrependimento aplicável e
  restrições a cláusulas abusivas. Vender para CNPJ não comprova, sozinho,
  relação paritária nem afasta normas de consumo.
- Decreto 7.962/2013, arts. 2–5: informação, resumo, correção de erros, cópia,
  atendimento e meios de arrependimento no comércio eletrônico abrangido.
- [Marco Civil, arts. 5º e 15](https://www.planalto.gov.br/ccivil_03/_ato2011-2014/2014/lei/l12965.htm): classificar registros de acesso sujeitos à guarda legal separadamente dos logs/evidências dos testes. Confirmar aplicabilidade antes de adotar o prazo geral de 90 dias da minuta.
- [ANPD, incidentes](https://www.gov.br/anpd/pt-br/canais_atendimento/agente-de-tratamento/comunicado-de-incidente-de-seguranca-cis): comunicação depende da hipótese, risco e papel efetivo. Operador informa sem demora injustificada; meta contratual não substitui prazo regulatório.
- [ANPD, transferências](https://www.gov.br/anpd/pt-br/assuntos/assuntos-internacionais/transferencia-internacional-de-dados): mecanismo depende do tratamento real. Não declarar regularidade sem verificar fornecedor e condições.

As referências sustentam o desenho dos modelos; sua aplicação às partes,
responsabilidade, setor e operação precisa de revisão profissional. O
procedimento `16` separa formação, prova e autorização técnica. O `17` converte
os deveres em rotina proposta, ainda não testada com clientes.

## Decisões não tomadas em nome do proprietário

Não foram escolhidos entidade, CNPJ, município fiscal, regime, foro, limite
financeiro, prazo definitivo de suporte ou assinatura comercial de fornecedor.
A revisão substituiu cláusulas automáticas arriscadas por decisão destacada a
aprovar, em vez de inventar uma proteção supostamente absoluta.

## O que este pacote não resolve sozinho

- não constitui parecer jurídico nem cria relação advogado-cliente;
- não confirma CNPJ, poderes, município, tributos, seguro ou foro;
- não substitui DPA/cláusulas oficiais dos fornecedores;
- não determina automaticamente se o CDC incide em cada Cliente;
- não garante validade de teto de responsabilidade em toda situação;
- não autoriza testes sem assinatura e prova do titular;
- não cobre setores regulados sem adaptação, como saúde, finanças, governo,
  telecom, crianças ou infraestrutura crítica.

O primeiro uso real deve ser aprovado por advogado brasileiro independente. A
revisão deve considerar a entidade já constituída, a oferta exata, o cliente e
o SOW; revisar apenas o contrato-mestre, sem os anexos técnicos, é insuficiente.
