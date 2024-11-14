package gpt

import (
	"datamin/config"
    "strings"
	"fmt"
	"net/http"
    "io"
)

func AskGpt(assistantCommand string, question string) (answer string) {
    client := &http.Client{}
    req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", nil)
    if err != nil {
        fmt.Println(err)
    }
    req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", config.GPT_AUTH_TOKEN))
    req.Header.Add("Content-Type", "application/json")
    req.Body = io.NopCloser(strings.NewReader(fmt.Sprintf(`
        {"model": "gpt-3.5-turbo",
        "messages": [
             {
              "role": "system",
              "content": "%s"
            },
            {
              "role": "user",
              "content": "%s"
            }
        ]}
    `, assistantCommand, question)))

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println(err)
    }
    bodyStr := string(body)
    fmt.Println(bodyStr)
    // return only the answer which is "content": "bee" 
    answer = strings.Split(bodyStr, "\"content\":")[1]
    answer = strings.Split(answer, "},")[0]
    answer = strings.TrimSpace(answer)
    answer = answer[1: len(answer)-1]
    fmt.Printf("answer: %v", answer)
    return answer
}

