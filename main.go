package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v2"
)

type Turma struct {
	Codigo    string
	Curso     string
	Programa  string
	Dias      string
	Horario   string
	Professor string
	Local     string
	Vagas     int
	Restam    int
}

type Config struct {
	ResendAPIKey    string
	EmailFrom       string
	EmailTo         string
	TwilioSID       string
	TwilioAuthToken string
	TwilioFrom      string
	TwilioTo        string
	CheckInterval   int
	FilterCourses   []string
}

var notifiedTurmas = make(map[string]bool)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env nao encontrado, usando variaveis de ambiente do sistema")
	}

	config := loadConfig()
	log.Printf("Iniciando monitoramento de vagas FESC...")
	log.Printf("Intervalo de verificacao: %d minutos", config.CheckInterval)
	if len(config.FilterCourses) > 0 {
		log.Printf("Filtro de cursos: %v", config.FilterCourses)
	} else {
		log.Printf("Monitorando: Todas as turmas de natacao")
	}

	// Executa imediatamente na primeira vez
	checkVagas(config)

	// Agenda verificacoes periodicas
	ticker := time.NewTicker(time.Duration(config.CheckInterval) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		checkVagas(config)
	}
}

func loadConfig() Config {
	interval, _ := strconv.Atoi(getEnv("CHECK_INTERVAL_MINUTES", "30"))

	var filters []string
	filterStr := os.Getenv("FILTER_COURSES")
	if filterStr != "" {
		filters = strings.Split(filterStr, ",")
		for i := range filters {
			filters[i] = strings.TrimSpace(filters[i])
		}
	}

	return Config{
		ResendAPIKey:  os.Getenv("RESEND_API_KEY"),
		EmailFrom:     getEnv("EMAIL_FROM", "onboarding@resend.dev"),
		EmailTo:       os.Getenv("EMAIL_TO"),
		CheckInterval: interval,
		FilterCourses: filters,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func checkVagas(config Config) {
	log.Printf("[%s] Verificando vagas...", time.Now().Format("02/01/2006 15:04:05"))

	turmas, err := scrapeTurmas()
	if err != nil {
		log.Printf("Erro ao buscar turmas: %v", err)
		return
	}

	// Filtra turmas de natacao com vagas disponiveis
	var turmasDisponiveis []Turma
	for _, t := range turmas {
		if !isNatacao(t.Curso) {
			continue
		}

		if len(config.FilterCourses) > 0 && !matchesFilter(t.Curso, config.FilterCourses) {
			continue
		}

		if t.Restam > 0 {
			turmasDisponiveis = append(turmasDisponiveis, t)
		}
	}

	if len(turmasDisponiveis) == 0 {
		log.Println("Nenhuma vaga de natacao disponivel no momento")
		return
	}

	// Verifica quais sao novas (nao notificadas ainda)
	var novasTurmas []Turma
	for _, t := range turmasDisponiveis {
		key := fmt.Sprintf("%s-%d", t.Codigo, t.Restam)
		if !notifiedTurmas[key] {
			novasTurmas = append(novasTurmas, t)
			notifiedTurmas[key] = true
		}
	}

	if len(novasTurmas) == 0 {
		log.Printf("Encontradas %d turmas com vagas, mas ja foram notificadas anteriormente", len(turmasDisponiveis))
		return
	}

	log.Printf("VAGAS ENCONTRADAS! %d novas turmas com vagas disponiveis", len(novasTurmas))

	// Envia notificacoes
	message := buildMessage(novasTurmas)

	if config.ResendAPIKey != "" && config.EmailTo != "" {
		if err := sendEmail(config, message); err != nil {
			log.Printf("Erro ao enviar email: %v", err)
		} else {
			log.Println("Email enviado com sucesso!")
		}
	}
}

func scrapeTurmas() ([]Turma, error) {
	resp, err := http.Get("https://fesc.app.br/vagas")
	if err != nil {
		return nil, fmt.Errorf("erro na requisicao: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear HTML: %w", err)
	}

	var turmas []Turma

	// Busca a tabela de turmas
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		cols := s.Find("td")
		if cols.Length() < 9 {
			return
		}

		vagas, _ := strconv.Atoi(strings.TrimSpace(cols.Eq(7).Text()))
		restam, _ := strconv.Atoi(strings.TrimSpace(cols.Eq(8).Text()))

		turma := Turma{
			Codigo:    strings.TrimSpace(cols.Eq(0).Text()),
			Curso:     strings.TrimSpace(cols.Eq(1).Text()),
			Programa:  strings.TrimSpace(cols.Eq(2).Text()),
			Dias:      strings.TrimSpace(cols.Eq(3).Text()),
			Horario:   strings.TrimSpace(cols.Eq(4).Text()),
			Professor: strings.TrimSpace(cols.Eq(5).Text()),
			Local:     strings.TrimSpace(cols.Eq(6).Text()),
			Vagas:     vagas,
			Restam:    restam,
		}

		turmas = append(turmas, turma)
	})

	return turmas, nil
}

func isNatacao(curso string) bool {
	cursoLower := strings.ToLower(curso)
	return strings.Contains(cursoLower, "natacao") || strings.Contains(cursoLower, "natação")
}

func matchesFilter(curso string, filters []string) bool {
	cursoLower := strings.ToLower(curso)
	for _, f := range filters {
		if strings.Contains(cursoLower, strings.ToLower(f)) {
			return true
		}
	}
	return false
}

func buildMessage(turmas []Turma) string {
	var sb strings.Builder
	sb.WriteString("VAGAS DE NATACAO DISPONIVEIS!\n\n")

	for _, t := range turmas {
		sb.WriteString(fmt.Sprintf("Turma: %s\n", t.Codigo))
		sb.WriteString(fmt.Sprintf("Curso: %s\n", t.Curso))
		sb.WriteString(fmt.Sprintf("Dias: %s\n", t.Dias))
		sb.WriteString(fmt.Sprintf("Horario: %s\n", t.Horario))
		sb.WriteString(fmt.Sprintf("Local: %s\n", t.Local))
		sb.WriteString(fmt.Sprintf("Vagas restantes: %d\n", t.Restam))
		sb.WriteString("-------------------\n\n")
	}

	sb.WriteString("Acesse: https://fesc.app.br/vagas")
	return sb.String()
}

func sendEmail(config Config, message string) error {
	client := resend.NewClient(config.ResendAPIKey)

	// Converte quebras de linha para HTML
	htmlMessage := strings.ReplaceAll(message, "\n", "<br>")

	params := &resend.SendEmailRequest{
		From:    config.EmailFrom,
		To:      []string{config.EmailTo},
		Subject: "ALERTA: Vagas de Natacao FESC Disponiveis!",
		Html:    fmt.Sprintf("<p>%s</p>", htmlMessage),
	}

	_, err := client.Emails.Send(params)
	return err
}
