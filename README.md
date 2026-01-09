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
