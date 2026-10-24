# Peitho AI - The Oracle of Text Safety

Peitho AI is a web application that evaluates the safety and risk levels of text across four distinct dimensions. It uses Google's **Gemini API** for intelligent scoring and structured reasoning, combined with local rule-based regex for privacy detection.

Designed with a mythological "Oracle" aesthetic, it provides a beautiful, split-panel interface for users to submit their "offerings" (text) and receive "prophecies" (risk assessments).

## Features

- **Multi-Dimensional Risk Scoring (0-100%):**
  - **Toxicity & Profanity:** Harassment, hate speech, and toxic language.
  - **PII & Privacy Leakage:** Emails, phone numbers, ID cards, and IP addresses (combines Gemini AI with local Regex).
  - **Phishing & Deception:** Scams, illicit behavior, and manipulative language.
  - **Violence & Self-Harm:** Threats of violence or self-harm.
- **Context-Aware Reasoning:** Select a context (e.g., General/Public, Professional/Work) and the AI Oracle will tailor its reasoning paragraph to fit the setting.
- **Structured Output:** Uses Gemini's JSON schema feature to guarantee reliable, parseable scores and text.
- **Embedded Static Files:** Go's `embed` package packs the frontend (HTML/JS) directly into the binary, making deployment a breeze.

## Tech Stack

- **Backend:** Go (Standard Library `net/http`)
- **AI Engine:** Google Gemini (`gemini-2.0-flash` via `google.golang.org/genai`)
- **Frontend:** Vanilla HTML, CSS, and JavaScript
- **Deployment:** Vercel

## Getting Started Locally

### Prerequisites
- [Go](https://go.dev/dl/) installed on your machine.
- A Google Gemini API Key. Get one for free at [Google AI Studio](https://aistudio.google.com/).

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/peitho.git
   cd peitho
2. **Download dependencies:**
   ```bash
   go mod tidy
   ```
3. Run the server
  ```bash
  go run .
  ```

### AI Usage
This project is vibe code project create by deepseek.
