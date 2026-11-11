package pii

import (
	"strings"
)

const maxTokenLookahead = 64

type StreamRehydrator struct {
	sessionID string
	vault     *Vault
	buffer    string
}

func NewStreamRehydrator(sessionID string, vault *Vault) *StreamRehydrator {
	return &StreamRehydrator{
		sessionID: sessionID,
		vault:     vault,
	}
}

// ProcessChunk processes a streaming delta chunk, holding incomplete token fragments in a window buffer
func (sr *StreamRehydrator) ProcessChunk(delta string) string {
	sr.buffer += delta

	var flushed strings.Builder

	for {
		openIdx := strings.Index(sr.buffer, "{{")
		if openIdx == -1 {
			// No opening delimiter found; entire buffer is safe to flush
			flushed.WriteString(sr.buffer)
			sr.buffer = ""
			break
		}

		// Text before opening delimiter is safe to flush immediately
		if openIdx > 0 {
			flushed.WriteString(sr.buffer[:openIdx])
			sr.buffer = sr.buffer[openIdx:]
			openIdx = 0
		}

		// Look for closing delimiter
		closeIdx := strings.Index(sr.buffer, "}}")
		if closeIdx != -1 {
			// Complete token candidate found
			tokenCandidate := sr.buffer[:closeIdx+2]
			rehydrated := sr.vault.RehydrateText(sr.sessionID, tokenCandidate)
			flushed.WriteString(rehydrated)
			sr.buffer = sr.buffer[closeIdx+2:]
			continue
		}

		// No closing delimiter yet
		if len(sr.buffer) > maxTokenLookahead {
			// Exceeded max token length, not a valid GuardMesh token, flush first character and continue
			flushed.WriteString(sr.buffer[:openIdx+2])
			sr.buffer = sr.buffer[openIdx+2:]
			continue
		}

		// Hold buffer for next chunk
		break
	}

	return flushed.String()
}

// FlushRemainder flushes any remaining text when stream closes
func (sr *StreamRehydrator) FlushRemainder() string {
	if sr.buffer == "" {
		return ""
	}
	remaining := sr.vault.RehydrateText(sr.sessionID, sr.buffer)
	sr.buffer = ""
	return remaining
}
