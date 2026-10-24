package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"
	"os"
	"regexp"
	"crypto/tls"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

//go:embed static/*
var staticFiles embed.FS

type ModerateRequest struct {
	Text    string `json:"text"`
	Context string `json:"context"`
}

type Scores struct {
	Profanity        float64 `json:"profanity"`
	Privacy          float64 `json:"privacy"`
	Phishing         float64 `json:"phishing"`
	ViolenceSelfHarm float64 `json:"violence_selfharm"`
}

type ModerateResponse struct {
	Scores    Scores `json:"scores"`
	Reasoning string `json:"reasoning"`
}

// Define the JSON structure Gemini should return
type GeminiOutput struct {
	Profanity        float64 `json:"profanity"`
	Privacy          float64 `json:"privacy"`
	Phishing         float64 `json:"phishing"`
	ViolenceSelfHarm float64 `json:"violence_selfharm"`
	Reasoning        string  `json:"reasoning"`
}

var privacyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`),
	regexp.MustCompile(`\b\d{15,18}[\dXx]\b`),
	regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),
}

func detectPrivacy(text string) float64 {
	for _, p := range privacyPatterns {
		if p.MatchString(text) {
			return 0.9
		}
	}
	return 0.0
}

func moderateHandler(client *genai.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ModerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Use Gemini for scoring and reasoning in one call
		scores, reasoning, err := analyzeWithGemini(r.Context(), client, req)
		if err != nil {
			log.Printf("Gemini error: %v", err)
			http.Error(w, "Oracle service failed", http.StatusInternalServerError)
			return
		}

		// Merge regex privacy score (take the higher value)
		regexPrivacy := detectPrivacy(req.Text)
		if regexPrivacy > scores.Privacy {
			scores.Privacy = regexPrivacy
		}

		resp := ModerateResponse{Scores: scores, Reasoning: reasoning}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func analyzeWithGemini(ctx context.Context, client *genai.Client, req ModerateRequest) (Scores, string, error) {
	systemPrompt := `You are "Peitho", an AI Oracle that evaluates text safety.
Analyze the text and provide risk scores from 0.0 to 1.0 for four dimensions:
- profanity (Toxicity & Profanity)
- privacy (PII & Privacy Leakage, detecting emails, phones, IDs, etc.)
- phishing (Phishing & Deception)
- violence_selfharm (Violence & Self-Harm)

Also write a short, objective, and slightly poetic reasoning under 60 words. Do not say "I".
You MUST return strictly valid JSON matching the schema.`

	userPrompt := fmt.Sprintf("Context: %s\nText to evaluate: \"%s\"", req.Context, req.Text)

	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"profanity":         {Type: genai.TypeNumber},
			"privacy":           {Type: genai.TypeNumber},
			"phishing":          {Type: genai.TypeNumber},
			"violence_selfharm": {Type: genai.TypeNumber},
			"reasoning":         {Type: genai.TypeString},
		},
		Required: []string{"profanity", "privacy", "phishing", "violence_selfharm", "reasoning"},
	}

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
	}

	// --- Tracing setup ---
	var dnsStart, connStart, tlsStart time.Time
	start := time.Now()

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			log.Printf("DNS lookup: %v", time.Since(dnsStart))
		},
		ConnectStart: func(_, _ string) { connStart = time.Now() },
		ConnectDone: func(_, _ string, err error) {
			log.Printf("TCP connect: %v", time.Since(connStart))
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			log.Printf("TLS handshake: %v", time.Since(tlsStart))
		},
		GotFirstResponseByte: func() {
			log.Printf("Gemini first byte received after: %v", time.Since(start))
		},
	}

	ctx = httptrace.WithClientTrace(ctx, trace)

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3.8-flash",
		genai.Text(userPrompt),
		config,
	)
	log.Printf("Total request time: %v", time.Since(start))

	if err != nil {
		return Scores{}, "", err
	}

	// Parse JSON returned by Gemini
	var output GeminiOutput
	if err := json.Unmarshal([]byte(result.Text()), &output); err != nil {
		return Scores{}, "", fmt.Errorf("failed to parse Gemini JSON: %w", err)
	}

	scores := Scores{
		Profanity:        output.Profanity,
		Privacy:          output.Privacy,
		Phishing:         output.Phishing,
		ViolenceSelfHarm: output.ViolenceSelfHarm,
	}

	return scores, output.Reasoning, nil
}

func main() {
	_ = godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/moderate", moderateHandler(client))

	// Serve embedded static files
	// staticFS, err := fs.Sub(staticFiles, "static")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// mux.Handle("/", http.FileServer(http.FS(staticFS)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Peitho AI (Gemini) running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
