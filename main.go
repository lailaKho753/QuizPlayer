package main

import (
    "encoding/json"
    "html/template"
    "net/http"
    "strings"
    "regexp"
)

type Question struct {
    ID          int      `json:"id"`
    Text        string   `json:"text"`
    Options     []string `json:"options"`
    Correct     string   `json:"correct"`
    Explanation string   `json:"explanation"`
}

func main() {
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/api/parse", parseHandler)
    http.HandleFunc("/api/submit", submitHandler)
    
    println("Server running at http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.ParseFiles("index.html"))
    tmpl.Execute(w, nil)
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    r.ParseForm()
    questionsText := r.FormValue("questions_text")
    
    questions := parseQuestions(questionsText)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "questions": questions,
        "total":     len(questions),
    })
}

func parseQuestions(text string) []Question {
    var questions []Question
    
    // Regex for format: question: ... answer: ... explanation: ...
    re := regexp.MustCompile(`(?s)question:\s*(.+?)\s*answer:\s*([A-D])\s*explanation:\s*(.+?)(?=\n\s*question:|$)`)
    matches := re.FindAllStringSubmatch(text, -1)
    
    for i, match := range matches {
        if len(match) >= 4 {
            questionText := strings.TrimSpace(match[1])
            answer := strings.ToUpper(strings.TrimSpace(match[2]))
            explanation := strings.TrimSpace(match[3])
            
            // Generate options A, B, C, D
            options := []string{
                "A. " + answer,
                "B. ",
                "C. ",
                "D. ",
            }
            
            questions = append(questions, Question{
                ID:          i,
                Text:        questionText,
                Options:     options,
                Correct:     answer,
                Explanation: explanation,
            })
        }
    }
    
    return questions
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var submission struct {
        Answers   map[int]string `json:"answers"`
        Questions []Question     `json:"questions"`
    }
    
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&submission); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    score := 0
    for i, q := range submission.Questions {
        if userAnswer, exists := submission.Answers[i]; exists {
            if userAnswer == q.Correct {
                score++
            }
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "score":       score,
        "total":       len(submission.Questions),
        "questions":   submission.Questions,
        "userAnswers": submission.Answers,
        "percentage":  float64(score) / float64(len(submission.Questions)) * 100,
    })
}