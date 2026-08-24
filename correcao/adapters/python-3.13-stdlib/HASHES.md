# Pinos e hashes auditados

- imagem Python: `sha256:00faa2debb87529f9f0764e9491d8ba400a3678976616c3bd7cb193745ac20d1`
- imagem do empacotador: `sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab`
- `first_case.go`: `2f7af1346997af1081193ce75dcef51c4564cb125ba570d7fbf1a490dfbd6f8e`
- `remediation.go`: `e8652b300f22c90d025458dedcbd823e9790b9ee7694f312271fc50420a8be77`
- runner Linux reproduzível ensaiado: `78557c02059a7e7b03a2ed44fccea0cf4fbf4315f8ee5f3fd4741062a1509732`

O hash do runner é recalculado em cada build e gravado fora do Git em
`pipeline/bin/pipeline-linux-amd64.sha256`. A configuração concreta incorpora
esse valor; divergência bloqueia o preflight.
