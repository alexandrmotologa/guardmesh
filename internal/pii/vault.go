package pii

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type TokenEntry struct {
	Token          string
	Category       EntityCategory
	EncryptedValue []byte
	Nonce          []byte
	CreatedAt      time.Time
}

type SessionTokens struct {
	forwardMap  map[string]string      // originalValue -> Token
	reverseMap  map[string]*TokenEntry // Token -> TokenEntry
	counts      map[EntityCategory]int
	lastUpdated time.Time
}

type Vault struct {
	mu       sync.RWMutex
	sessions map[string]*SessionTokens
	aesGCM   cipher.AEAD
	ttl      time.Duration
	stopCh   chan struct{}
}

func NewVault(ttl time.Duration) (*Vault, error) {
	// Generate random 256-bit AES key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate aes key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}

	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	v := &Vault{
		sessions: make(map[string]*SessionTokens),
		aesGCM:   gcm,
		ttl:      ttl,
		stopCh:   make(chan struct{}),
	}

	go v.cleanupLoop()

	return v, nil
}

func (v *Vault) Close() {
	select {
	case <-v.stopCh:
	default:
		close(v.stopCh)
	}
}

func (v *Vault) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			v.cleanupExpired()
		case <-v.stopCh:
			return
		}
	}
}

func (v *Vault) cleanupExpired() {
	v.mu.Lock()
	defer v.mu.Unlock()

	cutoff := time.Now().Add(-v.ttl)
	for id, s := range v.sessions {
		if s.lastUpdated.Before(cutoff) {
			delete(v.sessions, id)
		}
	}
}

func (v *Vault) encrypt(plaintext string) ([]byte, []byte, error) {
	nonce := make([]byte, v.aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext := v.aesGCM.Seal(nil, nonce, []byte(plaintext), nil)
	return ciphertext, nonce, nil
}

func (v *Vault) decrypt(ciphertext, nonce []byte) (string, error) {
	plaintext, err := v.aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// MaskText anonymizes entities inside text, storing cryptographic mappings for the given session ID
func (v *Vault) MaskText(sessionID string, text string, detector *Detector, mode string) (string, int, []string) {
	entities := detector.Detect(text)
	if len(entities) == 0 {
		return text, 0, nil
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	sess, ok := v.sessions[sessionID]
	if !ok {
		sess = &SessionTokens{
			forwardMap:  make(map[string]string),
			reverseMap:  make(map[string]*TokenEntry),
			counts:      make(map[EntityCategory]int),
			lastUpdated: time.Now(),
		}
		v.sessions[sessionID] = sess
	}
	sess.lastUpdated = time.Now()

	categoriesSet := make(map[string]struct{})
	var sb strings.Builder
	lastIdx := 0

	for _, ent := range entities {
		if ent.Start < lastIdx {
			continue // skip overlapped
		}

		sb.WriteString(text[lastIdx:ent.Start])
		categoriesSet[string(ent.Category)] = struct{}{}

		var replacement string
		switch mode {
		case "redact":
			replacement = fmt.Sprintf("[REDACTED_%s]", ent.Category)
		case "hash":
			h := sha256.Sum256([]byte(ent.Value))
			replacement = fmt.Sprintf("[HASH_%s]", hex.EncodeToString(h[:4]))
		default: // "token" mode
			token, exists := sess.forwardMap[ent.Value]
			if !exists {
				sess.counts[ent.Category]++
				token = fmt.Sprintf("{{GUARDMESH_ENTITY_%s_%d}}", ent.Category, sess.counts[ent.Category])

				encrypted, nonce, err := v.encrypt(ent.Value)
				if err == nil {
					entry := &TokenEntry{
						Token:          token,
						Category:       ent.Category,
						EncryptedValue: encrypted,
						Nonce:          nonce,
						CreatedAt:      time.Now(),
					}
					sess.forwardMap[ent.Value] = token
					sess.reverseMap[token] = entry
				}
			}
			replacement = token
		}

		sb.WriteString(replacement)
		lastIdx = ent.End
	}

	if lastIdx < len(text) {
		sb.WriteString(text[lastIdx:])
	}

	var categories []string
	for cat := range categoriesSet {
		categories = append(categories, cat)
	}

	return sb.String(), len(entities), categories
}

// RehydrateText restores original PII values for all tokens in a session
func (v *Vault) RehydrateText(sessionID string, text string) string {
	v.mu.RLock()
	sess, ok := v.sessions[sessionID]
	if !ok || len(sess.reverseMap) == 0 {
		v.mu.RUnlock()
		return text
	}

	// Copy tokens to replace under lock
	replacements := make(map[string]string, len(sess.reverseMap))
	for token, entry := range sess.reverseMap {
		if original, err := v.decrypt(entry.EncryptedValue, entry.Nonce); err == nil {
			replacements[token] = original
		}
	}
	v.mu.RUnlock()

	result := text
	for token, original := range replacements {
		result = strings.ReplaceAll(result, token, original)
	}
	return result
}

// ResolveToken returns the decrypted original string for a specific token
func (v *Vault) ResolveToken(sessionID string, token string) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	sess, ok := v.sessions[sessionID]
	if !ok {
		return "", false
	}

	entry, ok := sess.reverseMap[token]
	if !ok {
		return "", false
	}

	original, err := v.decrypt(entry.EncryptedValue, entry.Nonce)
	if err != nil {
		return "", false
	}
	return original, true
}
