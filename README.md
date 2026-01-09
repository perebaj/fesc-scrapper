# FESC Scrapper - Monitor de Vagas de Natacao

Scraper em Go que monitora vagas de natacao no site da FESC e envia notificacoes por email (Resend) e WhatsApp (Twilio).

## Configuracao

### 1. Copie o arquivo de configuracao

```bash
cp .env.example .env
```

### 2. Configure o Resend (Email)

1. Crie uma conta em https://resend.com
2. Va em **API Keys** e crie uma nova chave
3. Copie a API Key gerada

No `.env`:
```
RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxxxxxxx
EMAIL_FROM=onboarding@resend.dev
EMAIL_TO=seu_email@gmail.com
```

**Nota**: O remetente `onboarding@resend.dev` funciona apenas para testes. Para producao, adicione e verifique seu proprio dominio no Resend.

### 3. Configure o Twilio (WhatsApp)

1. Crie uma conta em https://www.twilio.com
2. Va em **Messaging > Try it out > Send a WhatsApp message**
3. Siga as instrucoes para conectar seu WhatsApp ao sandbox do Twilio
4. Copie as credenciais do dashboard

No `.env`:
```
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_WHATSAPP_FROM=whatsapp:+14155238886  # numero sandbox Twilio
TWILIO_WHATSAPP_TO=whatsapp:+5511999999999  # seu numero com DDI
```

**Importante**: Para usar o sandbox do Twilio, voce precisa enviar uma mensagem inicial para o numero do sandbox. Siga as instrucoes no painel do Twilio.

### 4. Configure o intervalo e filtros

```
CHECK_INTERVAL_MINUTES=30  # verifica a cada 30 minutos
FILTER_COURSES=            # vazio = todas as turmas de natacao
# ou especifique: FILTER_COURSES=Natacao Avancado,Natacao Iniciante
```

## Executando

### Modo desenvolvimento
```bash
go run main.go
```

### Compilar e executar
```bash
go build -o fesc-scrapper
./fesc-scrapper
```

### Executar em background (Linux/Mac)
```bash
nohup ./fesc-scrapper > scrapper.log 2>&1 &
```

### Usar com systemd (Linux)

Crie o arquivo `/etc/systemd/system/fesc-scrapper.service`:

```ini
[Unit]
Description=FESC Vagas Scrapper
After=network.target

[Service]
Type=simple
User=seu_usuario
WorkingDirectory=/caminho/para/fesc-scrapper
ExecStart=/caminho/para/fesc-scrapper/fesc-scrapper
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Depois:
```bash
sudo systemctl daemon-reload
sudo systemctl enable fesc-scrapper
sudo systemctl start fesc-scrapper
```

## Como funciona

1. O scraper acessa https://fesc.app.br/vagas periodicamente
2. Busca todas as turmas de natacao na tabela
3. Filtra as que tem vagas disponiveis (coluna "Restam" > 0)
4. Envia notificacao apenas para turmas novas (evita spam)
5. Registra logs de todas as verificacoes

## Logs

O programa exibe logs no console com informacoes sobre cada verificacao:

```
2024/01/15 10:30:00 Iniciando monitoramento de vagas FESC...
2024/01/15 10:30:00 Intervalo de verificacao: 30 minutos
2024/01/15 10:30:00 Monitorando: Todas as turmas de natacao
2024/01/15 10:30:01 [15/01/2024 10:30:01] Verificando vagas...
2024/01/15 10:30:02 Nenhuma vaga de natacao disponivel no momento
```
