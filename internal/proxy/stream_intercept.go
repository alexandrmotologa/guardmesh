package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alexandrmotologa/guardmesh/internal/pii"
)

// InterceptOpenAISSEStream reads SSE chunks from upstream, rehydrates PII across packet boundaries, and writes to client
func InterceptOpenAISSEStream(upstream io.Reader, w http.ResponseWriter, flusher http.Flusher, rehydrator *pii.StreamRehydrator, streamSuffix string) (int, error) {
	scanner := bufio.NewScanner(upstream)
	totalTokens := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			dataPayload := strings.TrimPrefix(line, "data: ")

			if strings.TrimSpace(dataPayload) == "[DONE]" {
				// Flush any remaining buffered token fragments before sending [DONE]
				remaining := rehydrator.FlushRemainder()
				if streamSuffix != "" {
					remaining += streamSuffix
				}
				if remaining != "" {
					syntheticDoneChunk := fmt.Sprintf(`{"choices":[{"index":0,"delta":{"content":%s}}]}`, jsonString(remaining))
					fmt.Fprintf(w, "data: %s\n\n", syntheticDoneChunk)
				}
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				break
			}

			// Parse chunk JSON
			var chunk map[string]interface{}
			if err := json.Unmarshal([]byte(dataPayload), &chunk); err == nil {
				choices, ok := chunk["choices"].([]interface{})
				if ok && len(choices) > 0 {
					choiceMap, ok := choices[0].(map[string]interface{})
					if ok {
						delta, ok := choiceMap["delta"].(map[string]interface{})
						if ok {
							content, ok := delta["content"].(string)
							if ok && content != "" {
								totalTokens++
								rehydratedDelta := rehydrator.ProcessChunk(content)
								delta["content"] = rehydratedDelta
								choiceMap["delta"] = delta
								choices[0] = choiceMap
								chunk["choices"] = choices

								newJSON, err := json.Marshal(chunk)
								if err == nil {
									line = "data: " + string(newJSON)
								}
							}
						}
					}
				}
			}
		}

		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()
	}

	return totalTokens, scanner.Err()
}

// InterceptAnthropicSSEStream reads Anthropic SSE chunks, rehydrates PII, and streams to client
func InterceptAnthropicSSEStream(upstream io.Reader, w http.ResponseWriter, flusher http.Flusher, rehydrator *pii.StreamRehydrator, streamSuffix string) (int, error) {
	scanner := bufio.NewScanner(upstream)
	totalTokens := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			dataPayload := strings.TrimPrefix(line, "data: ")

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(dataPayload), &event); err == nil {
				eventType, _ := event["type"].(string)

				if eventType == "content_block_delta" {
					delta, ok := event["delta"].(map[string]interface{})
					if ok {
						text, ok := delta["text"].(string)
						if ok && text != "" {
							totalTokens++
							rehydratedText := rehydrator.ProcessChunk(text)
							delta["text"] = rehydratedText
							event["delta"] = delta

							newJSON, err := json.Marshal(event)
							if err == nil {
								line = "data: " + string(newJSON)
							}
						}
					}
				} else if eventType == "message_stop" {
					remaining := rehydrator.FlushRemainder()
					if streamSuffix != "" {
						remaining += streamSuffix
					}
					if remaining != "" {
						syntheticDelta := fmt.Sprintf(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":%s}}`, jsonString(remaining))
						fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", syntheticDelta)
					}
				}
			}
		}

		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()
	}

	return totalTokens, scanner.Err()
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
