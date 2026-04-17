package main

import (
    "encoding/json"
    "net/http"
    "strings"
)

type Question struct {
    ID          int      `json:"id"`
    Text        string   `json:"text"`
    Options     []string `json:"options"`
    Correct     string   `json:"correct"`
    Explanation string   `json:"explanation"`
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "index.html")
    })
    http.HandleFunc("/api/parse", parseHandler)
    http.HandleFunc("/api/submit", submitHandler)

    println("✅ Server at http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func parseHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    text := r.FormValue("questions_text")
    
    // DEBUG: print raw text to terminal
    println("\n=== RAW TEXT ===")
    println(text)
    println("================\n")
    
    var questions []Question
    
    parts := strings.Split(text, "question:")
    
    println("Found", len(parts)-1, "questions")
    
    for i, part := range parts {
        if i == 0 {
            continue
        }
        
        println("\n--- Processing part", i, "---")
        println("Raw part:", part[:min(len(part), 100)])
        
        part = strings.TrimSpace(part)
        if part == "" {
            println("Part is empty, skipping")
            continue
        }
        
        // Extract question (everything before "options:")
        questionText := part
        if idx := strings.Index(strings.ToLower(part), "options:"); idx != -1 {
            questionText = strings.TrimSpace(part[:idx])
        } else if idx := strings.Index(strings.ToLower(part), "answer:"); idx != -1 {
            questionText = strings.TrimSpace(part[:idx])
        }
        println("Question text:", questionText)
        
        // Default options
        options := []string{"A. ", "B. ", "C. ", "D. "}
        
        // Extract options
        lowerPart := strings.ToLower(part)
        if idx := strings.Index(lowerPart, "options:"); idx != -1 {
            optPart := part[idx+8:]
            if ansIdx := strings.Index(strings.ToLower(optPart), "answer:"); ansIdx != -1 {
                optPart = optPart[:ansIdx]
            }
            optPart = strings.TrimSpace(optPart)
            println("Options raw:", optPart)
            
            // Clean up: replace commas with spaces
            optPart = strings.ReplaceAll(optPart, ",", " ")
            // Remove multiple spaces
            for strings.Contains(optPart, "  ") {
                optPart = strings.ReplaceAll(optPart, "  ", " ")
            }
            
            // Split by space
            optList := strings.Split(optPart, " ")
            optIndex := 0
            for _, opt := range optList {
                opt = strings.TrimSpace(opt)
                if opt == "" {
                    continue
                }
                // Skip A., B., etc
                if len(opt) > 1 && opt[1] == '.' {
                    if len(opt) > 2 {
                        opt = opt[2:]
                    } else {
                        continue
                    }
                }
                if optIndex < 4 && opt != "" {
                    options[optIndex] = string(rune('A'+optIndex)) + ". " + opt
                    println("  Option", string(rune('A'+optIndex)), ":", opt)
                    optIndex++
                }
            }
        }
        
        // Extract answer
        answer := "A"
        if idx := strings.Index(strings.ToLower(part), "answer:"); idx != -1 {
            ansPart := part[idx+7:]
            if expIdx := strings.Index(strings.ToLower(ansPart), "explanation:"); expIdx != -1 {
                ansPart = ansPart[:expIdx]
            }
            ansPart = strings.TrimSpace(ansPart)
            if len(ansPart) > 0 {
                answer = strings.ToUpper(string(ansPart[0]))
            }
            println("Answer:", answer)
        }
        
        // Extract explanation
        explanation := ""
        if idx := strings.Index(strings.ToLower(part), "explanation:"); idx != -1 {
            explanation = strings.TrimSpace(part[idx+12:])
            // Cut at newline
            if nlIdx := strings.Index(explanation, "\n"); nlIdx != -1 {
                explanation = explanation[:nlIdx]
            }
            println("Explanation:", explanation[:min(len(explanation), 50)])
        }
        
        if questionText != "" {
            questions = append(questions, Question{
                ID:          len(questions),
                Text:        questionText,
                Options:     options,
                Correct:     answer,
                Explanation: explanation,
            })
            println("✅ Question added!")
        } else {
            println("❌ Question text empty, skipping")
        }
    }
    
    println("\n=== TOTAL QUESTIONS:", len(questions), "===\n")
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "questions": questions,
        "total":     len(questions),
    })
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Answers   map[int]string `json:"answers"`
        Questions []Question     `json:"questions"`
    }
    json.NewDecoder(r.Body).Decode(&req)
    
    score := 0
    for i, q := range req.Questions {
        if ans, ok := req.Answers[i]; ok && ans == q.Correct {
            score++
        }
    }
    
    pct := 0.0
    if len(req.Questions) > 0 {
        pct = float64(score) / float64(len(req.Questions)) * 100
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "score":       score,
        "total":       len(req.Questions),
        "percentage":  pct,
        "questions":   req.Questions,
        "userAnswers": req.Answers,
    })
}