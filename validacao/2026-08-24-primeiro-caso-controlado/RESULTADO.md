# Resultado do primeiro caso controlado

Status: aprovado tecnicamente no laboratório; não é autorização para cliente.

- laboratório privado: `yurIdeLimaDev/empresa-security-python-reference-lab`;
- baseline: `e1790108d608261e1fc7e0317b94dc5e677201d4`;
- BEST final: `16ba133b334bbdd371a26e125f0aadfe19716f2c`;
- cinco achados fictícios corrigidos em cinco tickets, uma tentativa aceita por
  ticket; uma tentativa adicional corresponde ao gate global;
- bundle final: `9660d1134a69939c5e6e6d8deebdf3641bfab87149b7ccd115bde28e72b46db8`;
- gate global: aprovado, zero achados e nenhuma regressão;
- testes do laboratório: 8/8; CI privada: sucesso;
- manifesto de entrega: `987381093fb551d39b0ee8f1b60d0e1c993cf91a259ca5e829260a1b77adc435`;
- ZIP determinístico: `54804c32f77056f7aefd4605c7eb0d9da2e605e6d4a304cbd05dc2fa0f84225d`;
- empacotamento host/contêiner produziu os mesmos hashes de manifesto e ZIP;
- backup Linux em contêiner: restaurado, 2/2 arquivos com hashes idênticos;
- runner Linux reproduzível: `78557c02059a7e7b03a2ed44fccea0cf4fbf4315f8ee5f3fd4741062a1509732`.

A aprovação humana foi simulada e está identificada como tal. Nenhuma chave
privada, patch operacional, worktree ou evidência bruta foi incluída aqui.

Limitação: o drift/preflight precisa ser repetido no host Linux dedicado que for
escolhido. Docker Linux no desktop validou empacotamento e restauração, mas não
prova SSH, updates e firewall de um VPS futuro.
