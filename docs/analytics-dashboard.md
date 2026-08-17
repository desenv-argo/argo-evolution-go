# Argo Dashboard e conversas

O Manager inclui um dashboard operacional e um explorador de conversas em `/manager`. As consultas são administrativas e exigem `GLOBAL_API_KEY` no header `apikey`.

## Captura de mensagens

A captura estruturada pode ser ativada ou pausada diretamente no Dashboard ou na tela Conversas. A escolha é persistida na tabela `analytics_settings` e sobrevive a reinícios e novos deploys.

`DATABASE_SAVE_MESSAGES` continua existindo apenas como valor inicial na primeira execução. Depois que `analytics_settings` é criada, não é necessário alterar uma variável no Azure nem reiniciar o App Service para ativar ou pausar a funcionalidade.

A captura começa com as novas mensagens processadas depois da ativação. O sistema não promete reconstruir o histórico anterior do WhatsApp.

São persistidos apenas campos normalizados para consulta: instância, conversa, remetente, direção, tipo, texto/legenda, metadados da mídia, status, datas, mensagem citada e referral. O protobuf bruto, thumbnails e arquivos em base64 não são armazenados na tabela de mensagens.

## Endpoints administrativos

- `GET /analytics/dashboard`: métricas de instâncias, mensagens, conversas e volume diário.
- `GET /analytics/conversations`: conversas agregadas, com busca e filtros.
- `GET /analytics/messages`: timeline de uma conversa.
- `GET /analytics/settings`: estado atual da captura.
- `PUT /analytics/settings`: ativa ou pausa a captura em runtime.

Filtros compartilhados:

- `from` e `to`: RFC3339 ou `YYYY-MM-DD`, com período máximo de 366 dias;
- `instanceId`;
- `search`;
- `direction`: `inbound` ou `outbound`;
- `messageType`;
- `status`;
- `isGroup`;
- `before` e `limit` para paginação.

Para `GET /analytics/messages`, `instanceId` e `chatJid` são obrigatórios.

## Banco de dados

O `AutoMigrate` adiciona os campos estruturados e índices na tabela `messages`. Registros legados permanecem preservados, mas não aparecem nas novas consultas quando não possuem `instance_id`, `chat_jid` e `sent_at`.

O deploy não exige limpeza da tabela e a aplicação nova continua aceitando os registros antigos usados pelo endpoint de status. Como `ALTER TABLE` e criação de índices podem adquirir locks no PostgreSQL, em bases grandes a migração deve ser validada em staging e executada em uma janela de menor movimento ou separada do startup da aplicação.

## Segurança e privacidade

- Todos os endpoints de leitura usam autenticação administrativa.
- Tokens de instância não são retornados pelas APIs de analytics.
- Conteúdo vindo das mensagens é inserido no DOM com `textContent`, evitando execução de HTML da conversa.
- Arquivos permanecem no storage configurado; o banco guarda somente a referência da mídia.
- Retenção automática, auditoria de leitura e permissões mais granulares devem ser adicionadas antes de liberar acesso a perfis que não sejam administradores.

## Plano de evolução dos insights

O modelo atual guarda a base factual necessária para análises futuras sem acoplar o produto a um provedor de IA: identidade da instância e da conversa, direção, participante, tipo e conteúdo da mensagem, mídia, citação, referral, status e timestamps.

Próximos indicadores recomendados, em ordem:

1. **Atendimento:** tempo até a primeira resposta, tempo entre mensagens, conversas aguardando resposta e idade do backlog, com percentis P50/P90 por instância e faixa horária.
2. **Qualidade operacional:** taxa de entrega e leitura, falhas por instância, horários de pico, volume por atendente/canal e distribuição por tipo de mídia.
3. **Relacionamento:** contatos novos e recorrentes, duração da conversa, reabertura em 7/30 dias, origem/referral e equilíbrio entre mensagens recebidas e enviadas.
4. **Insights semânticos:** assunto, intenção, sentimento, objeções, pedidos recorrentes e resumo da conversa. Essa fase deve escrever resultados derivados em tabelas separadas, com versão do modelo e possibilidade de reprocessamento.

Antes de liberar insights semânticos em produção, devem entrar controles de retenção, mascaramento de dados sensíveis, exclusão por contato, trilha de auditoria e perfis de acesso. Assim, o histórico útil para negócio não vira uma cópia irrestrita de dados pessoais.

## Build do Manager

O projeto original fornece somente o bundle compilado do Manager. Para preservar as telas existentes sem editar manualmente o minificado, os componentes Argo ficam em `manager/src/argo-manager.js`. O script `manager/scripts/build.mjs`:

1. copia as logos e o componente para `manager/dist/assets`;
2. substitui de forma determinística os placeholders de Dashboard e Mensagens;
3. falha caso o bundle original mude e os pontos de integração não sejam encontrados.

O Dockerfile executa essa validação em todos os builds da imagem.

## Deploy automático no Azure

O workflow de produção publica três tags no ACR: a versão, `latest` e a tag imutável `sha-<commit>`. Em seguida, `azure/webapps-deploy` aponta o App Service para a tag imutável exata. Essa troca de configuração faz o App Service baixar e iniciar a imagem nova; não é necessário reiniciar o serviço manualmente e o deploy não depende de o Azure perceber uma alteração na tag mutável `latest`.

Configure uma única vez no repositório GitHub:

- variável `AZURE_WEBAPP_NAME`: nome do App Service;
- secret `AZURE_WEBAPP_PUBLISH_PROFILE`: conteúdo integral do perfil de publicação baixado no App Service;
- variável opcional `AZURE_WEBAPP_HEALTHCHECK_URL`: URL completa de health check, por exemplo `https://evogo.argo.app.br/server/ok`.

Se uma configuração obrigatória estiver ausente, o workflow falha com uma mensagem clara depois do build, em vez de indicar sucesso enquanto a imagem permanece apenas no registro.
