# Argo Messaging Hub — roadmap de evolução

## Objetivo

Evoluir o fork do Evolution Go para uma plataforma de mensageria observável, capaz de atender ERP, Athlas e outros projetos sem acoplar as customizações Argo ao núcleo do upstream.

O Evolution Go permanece responsável pelo transporte WhatsApp. Os módulos sob `pkg/argo` formam o plano de controle responsável por identidade das aplicações, rastreabilidade, métricas, saúde, confiabilidade de webhooks e insights.

## Princípios de compatibilidade

1. Funcionalidades Argo devem ser aditivas e residir em `pkg/argo` sempre que possível.
2. APIs próprias usam o prefixo versionado `/argo/v1`.
3. Tabelas próprias usam nomes e modelos Argo, sem alterar o significado dos modelos upstream.
4. Headers Argo são opcionais durante a migração; consumidores antigos continuam funcionando.
5. Segredos são armazenados somente como hash e nunca retornados após criação ou rotação.
6. Atualizações upstream devem ocorrer por versão fixada e PR de sincronização, nunca pela tag `latest` diretamente em produção.
7. Toda integração nova deve possuir testes de contrato e correlation ID ponta a ponta.

## Fase 1 — identidade e rastreabilidade

Status: fundação implementada no PR #12.

- catálogo administrativo de aplicações;
- credencial própria por aplicação, retornada apenas na criação ou rotação;
- identificação declarada por `X-Argo-Application-Id`;
- autenticação da identidade por `X-Argo-Application-Key`;
- propagação ou criação de `X-Correlation-Id`;
- captura de `Idempotency-Key`;
- persistência de todas as tentativas em `/send/*`, inclusive 4xx e 5xx;
- taxonomia inicial de erros;
- resumo operacional e consulta de tentativas por aplicação, instância e período;
- tela enterprise de Integrações no Manager para catálogo, rotação de credencial,
  atividade recente, métricas, filtros e drill-down de erros;
- compatibilidade com consumidores legados, identificados como `legacy/unknown`.

### Contrato do consumidor

```http
apikey: <token-da-instancia>
X-Argo-Application-Id: argo-erp
X-Argo-Application-Key: <credencial-da-aplicacao>
X-Correlation-Id: <uuid-da-operacao>
Idempotency-Key: <uuid-do-comando>
```

O token da instância continua autorizando a operação no Evolution. A credencial Argo autentica qual aplicação está originando a chamada. Nesta fase, uma identidade ausente ou inválida é registrada como não verificada, mas não bloqueia a operação.

## Fase 2 — saúde das integrações

Status: heartbeat assinado e visão operacional implementados; polling ativo permanece pendente.

- [ ] health checks controlados pelo plano de controle, após definição de allowlist;
- [x] heartbeat assinado para aplicações que não expõem endpoint consultável;
- estados separados para aplicação, autenticação, instância e fluxo de eventos;
- [x] histórico de sinais e latência;
- alertas por falhas consecutivas, ausência de atividade e degradação;
- proteção contra SSRF por allowlist de hosts e redes permitidas.

### Contrato de heartbeat

A aplicação envia o sinal no intervalo configurado em
`expected_heartbeat_seconds`. A ausência por duas janelas muda o estado para
`offline`. O endpoint exige a credencial própria da aplicação e não aceita a
chave administrativa do Manager.

```http
POST /argo/v1/heartbeat
Content-Type: application/json
X-Argo-Application-Id: argo-erp
X-Argo-Application-Key: <credencial-da-aplicacao>

{
  "status": "healthy",
  "latency_ms": 42,
  "version": "2026.08.18",
  "component": "whatsapp-outbox",
  "message": "optional diagnostic without sensitive data"
}
```

Estados aceitos: `healthy`, `degraded` e `unhealthy`. A tela de Integrações
mantém atividade autenticada da API e saúde declarada como sinais separados,
evitando classificar uma aplicação como saudável apenas porque realizou um
envio.

## Fase 3 — entrega confiável de eventos

- outbox persistente para webhooks;
- worker com timeout, backoff exponencial e jitter;
- fila de mensagens mortas;
- replay manual e automático;
- idempotência por evento;
- métricas de backlog, atraso, sucesso e falha por destino;
- retenção e mascaramento de payloads sensíveis.

## Fase 4 — ciclo completo da mensagem

Status: primeiro vertical slice operacional implementado.

- [x] transições `received`, `validated`, `accepted`, `sent`, `delivered`, `read`, `failed` e `pending_aged`;
- [x] estados de falha conhecidos e categorizados;
- [x] histórico imutável em `argo_message_lifecycle_events`, sem sobrescrever evidências anteriores;
- [x] correlação por tentativa, `provider_message_id`, instância, aplicação, correlation ID e idempotency key;
- [x] reconciliação de `pending_aged` para ausência de receipt após janela configurável;
- [x] taxas de aceitação, envio, entrega e leitura;
- [x] tempos P50, P90, P95 e P99 para envio, entrega e leitura;
- [x] filtros por aplicação, instância, período, tipo, estado e identificadores de correlação;
- [x] API administrativa em `/argo/v1/messages/lifecycle` e `/argo/v1/messages/lifecycle/summary`;
- [x] tela enterprise com funil, falhas, latências e drill-down da linha do tempo;
- [x] worker periódico dedicado para reconciliação fora do ciclo de consultas;
  - executa imediatamente após o startup e depois por intervalo, sem sobreposição;
  - `ARGO_MESSAGE_PENDING_RECONCILE_SECONDS` define o intervalo (padrão: 60 segundos);
  - `ARGO_MESSAGE_PENDING_RECONCILE_TIMEOUT_SECONDS` define o timeout de cada execução (padrão: 10 segundos);
  - consultas de eventos e métricas permanecem somente leitura;
- [x] backfill administrativo opcional para mensagens anteriores à criação da tabela Argo;
  - `POST /argo/v1/messages/lifecycle/backfill` exige período explícito de no máximo 31 dias;
  - simula por padrão e somente persiste quando `execute: true`;
  - é limitado, idempotente e reconstrói apenas tentativas e receipts já persistidos.

Exemplo de simulação segura:

```json
{
  "from": "2026-08-01T00:00:00Z",
  "to": "2026-08-02T00:00:00Z",
  "limit": 1000,
  "execute": false
}
```

A janela padrão de envelhecimento é 15 minutos e pode ser alterada com
`ARGO_MESSAGE_PENDING_AGE_MINUTES` (1 a 10.080 minutos). Envios sem headers Argo
continuam no funil como `legacy/unknown`.

## Fase 5 — operação e decisão

### Captura de conversas e anexos

- mensagens recuperadas por `HistorySync` também alimentam a visão de conversas, cobrindo envios feitos por outros dispositivos e períodos offline;
- documentos e demais mídias recebidas são armazenados em `argo_message_media` e servidos somente pela API administrativa;
- o Manager oferece prévia autenticada sob demanda para imagens, áudio e vídeo, além de visualização e download de documentos;
- `ARGO_MESSAGE_MEDIA_MAX_BYTES` limita cada anexo capturado (padrão: 25 MiB; máximo: 100 MiB).
- a retenção automática remove anexos antigos em lotes; `ARGO_MESSAGE_MEDIA_RETENTION_DAYS` define a janela (padrão: 30 dias), `ARGO_MESSAGE_MEDIA_CLEANUP_SECONDS` o intervalo (padrão: 6 horas) e `ARGO_MESSAGE_MEDIA_CLEANUP_BATCH` o lote (padrão: 500).

- dashboard por aplicação e ambiente;
- consumo por aplicação e instância;
- concentração de erros;
- disponibilidade e reconexões por instância;
- SLA de primeira resposta e conversas aguardando atendimento;
- tendências, anomalias e capacidade;
- auditoria de credenciais, alterações e replay de eventos.

## Fase 6 — insights

- classificação de assunto e intenção;
- sentimento e sinais de insatisfação;
- objeções e pedidos recorrentes;
- resumo de conversa;
- resultados derivados versionados e reprocessáveis;
- retenção, mascaramento, exclusão por contato e controle de acesso antes da liberação ampla.

## Estratégia de atualização upstream

- adicionar `evolution-foundation/evolution-go` como remote `upstream` no processo de manutenção;
- executar verificação semanal de novas releases;
- abrir PR automático de sincronização com relatório de conflitos;
- manter uma lista explícita dos poucos arquivos upstream tocados pelas extensões;
- migrar o painel Argo para uma rota ou SPA própria, evitando depender do bundle compilado do Manager;
- executar E2E de conexão, envio, receipt, webhook e analytics antes de qualquer merge de upstream;
- contribuir correções genéricas de QR, reconexão e pool de banco de volta ao upstream para reduzir o patch local.

## Métricas prioritárias

- taxa de aceitação: aceitas / tentativas;
- taxa de envio: enviadas / aceitas;
- taxa de entrega: entregues / enviadas elegíveis;
- taxa de leitura: lidas / entregues;
- pendentes envelhecidas após a janela configurada;
- erros por aplicação, instância, endpoint e categoria;
- latência média e percentis;
- identidades não verificadas;
- disponibilidade de aplicação, integração e instância;
- sucesso, atraso e backlog de webhooks.
