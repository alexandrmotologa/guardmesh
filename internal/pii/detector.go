package pii

import (
	"math/big"
	"net"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/alexandrmotologa/guardmesh/internal/config"
)

type EntityCategory string

const (
	CategoryCreditCard EntityCategory = "CREDIT_CARD"
	CategoryIBAN       EntityCategory = "IBAN"
	CategoryEmail      EntityCategory = "EMAIL"
	CategoryPhone      EntityCategory = "PHONE"
	CategorySSN        EntityCategory = "SSN"
	CategoryIPAddress  EntityCategory = "IP_ADDRESS"
	CategoryAPISecret  EntityCategory = "API_SECRET"
	CategoryPersonName EntityCategory = "PERSON_NAME"
)

type DetectedEntity struct {
	Category   EntityCategory
	Value      string
	Start      int
	End        int
	Confidence float64
}

type Detector struct {
	cfg config.DetectorsConfig

	emailRegex     *regexp.Regexp
	ssnRegex       *regexp.Regexp
	phoneRegex     *regexp.Regexp
	ibanRegex      *regexp.Regexp
	ccRegex        *regexp.Regexp
	apiKeyRegexes  []*regexp.Regexp
	titleNameRegex *regexp.Regexp
	introNameRegex *regexp.Regexp
}

func NewDetector(cfg config.DetectorsConfig) *Detector {
	return &Detector{
		cfg: cfg,
		emailRegex: regexp.MustCompile(`(?i)\b[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}\b`),
		ssnRegex:   regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
		phoneRegex: regexp.MustCompile(`(?:\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`),
		ibanRegex:  regexp.MustCompile(`\b[A-Z]{2}\d{2}[A-Z0-9]{11,30}\b`),
		ccRegex:    regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b|\b\d{13,19}\b`),
		apiKeyRegexes: []*regexp.Regexp{
			regexp.MustCompile(`\bsk-[a-zA-Z0-9]{20,}\b`),                                          // OpenAI API keys
			regexp.MustCompile(`\bghp_[a-zA-Z0-9]{36}\b`),                                          // GitHub Personal Access Token
			regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),                                             // AWS Access Key ID
			regexp.MustCompile(`\bey[A-Za-z0-9_-]{15,}\.ey[A-Za-z0-9_-]{15,}\.[A-Za-z0-9_-]{15,}\b`), // JWT Token
			regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`),                // Private Keys
		},
		titleNameRegex: regexp.MustCompile(`\b(?:Mr\.|Mrs\.|Ms\.|Dr\.|Prof\.|Herr|Frau|Monsieur)\s+([A-Z][a-z]+(?:\s+[A-Z][a-z]+)?)\b`),
		introNameRegex: regexp.MustCompile(`(?i)\b(?:name is|patient|customer|client|contact|user)\s+([A-Z][a-z]+\s+[A-Z][a-z]+)\b`),
	}
}

func isValidSSN(ssn string) bool {
	parts := strings.Split(ssn, "-")
	if len(parts) != 3 {
		return false
	}
	area, err1 := strconv.Atoi(parts[0])
	group, err2 := strconv.Atoi(parts[1])
	serial, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}
	if area == 0 || area == 666 || area >= 900 {
		return false
	}
	if group == 0 || serial == 0 {
		return false
	}
	return true
}

func (d *Detector) Detect(text string) []DetectedEntity {
	var results []DetectedEntity

	if d.cfg.Email {
		for _, match := range d.emailRegex.FindAllStringIndex(text, -1) {
			results = append(results, DetectedEntity{
				Category:   CategoryEmail,
				Value:      text[match[0]:match[1]],
				Start:      match[0],
				End:        match[1],
				Confidence: 0.98,
			})
		}
	}

	if d.cfg.SSN {
		for _, match := range d.ssnRegex.FindAllStringIndex(text, -1) {
			val := text[match[0]:match[1]]
			if isValidSSN(val) {
				results = append(results, DetectedEntity{
					Category:   CategorySSN,
					Value:      val,
					Start:      match[0],
					End:        match[1],
					Confidence: 0.95,
				})
			}
		}
	}

	if d.cfg.APISecrets {
		for _, rx := range d.apiKeyRegexes {
			for _, match := range rx.FindAllStringIndex(text, -1) {
				results = append(results, DetectedEntity{
					Category:   CategoryAPISecret,
					Value:      text[match[0]:match[1]],
					Start:      match[0],
					End:        match[1],
					Confidence: 0.99,
				})
			}
		}
	}

	if d.cfg.IBAN {
		for _, match := range d.ibanRegex.FindAllStringIndex(text, -1) {
			val := text[match[0]:match[1]]
			if isValidIBAN(val) {
				results = append(results, DetectedEntity{
					Category:   CategoryIBAN,
					Value:      val,
					Start:      match[0],
					End:        match[1],
					Confidence: 0.99,
				})
			}
		}
	}

	if d.cfg.CreditCard {
		for _, match := range d.ccRegex.FindAllStringIndex(text, -1) {
			raw := text[match[0]:match[1]]
			clean := strings.ReplaceAll(strings.ReplaceAll(raw, "-", ""), " ", "")
			if len(clean) >= 13 && len(clean) <= 19 && isValidLuhn(clean) {
				results = append(results, DetectedEntity{
					Category:   CategoryCreditCard,
					Value:      raw,
					Start:      match[0],
					End:        match[1],
					Confidence: 0.99,
				})
			}
		}
	}

	if d.cfg.Phone {
		for _, match := range d.phoneRegex.FindAllStringIndex(text, -1) {
			val := text[match[0]:match[1]]
			digitsOnly := keepDigits(val)
			if len(digitsOnly) >= 7 && len(digitsOnly) <= 15 {
				results = append(results, DetectedEntity{
					Category:   CategoryPhone,
					Value:      val,
					Start:      match[0],
					End:        match[1],
					Confidence: 0.85,
				})
			}
		}
	}

	if d.cfg.IPAddress {
		words := strings.FieldsFunc(text, func(r rune) bool {
			return unicode.IsSpace(r) || r == ',' || r == ';' || r == ':' || r == '"' || r == '\'' || r == '<' || r == '>' || r == '[' || r == ']'
		})
		for _, word := range words {
			ip := net.ParseIP(word)
			if ip != nil && !ip.IsLoopback() {
				idx := strings.Index(text, word)
				if idx >= 0 {
					results = append(results, DetectedEntity{
						Category:   CategoryIPAddress,
						Value:      word,
						Start:      idx,
						End:        idx + len(word),
						Confidence: 0.95,
					})
				}
			}
		}
	}

	if d.cfg.PersonName {
		// Title-based names: "Dr. Alice Smith"
		for _, match := range d.titleNameRegex.FindAllStringSubmatchIndex(text, -1) {
			if len(match) >= 4 && match[2] >= 0 && match[3] >= 0 {
				results = append(results, DetectedEntity{
					Category:   CategoryPersonName,
					Value:      text[match[2]:match[3]],
					Start:      match[2],
					End:        match[3],
					Confidence: 0.90,
				})
			}
		}
		// Contextual intro names: "name is Alice Smith"
		for _, match := range d.introNameRegex.FindAllStringSubmatchIndex(text, -1) {
			if len(match) >= 4 && match[2] >= 0 && match[3] >= 0 {
				results = append(results, DetectedEntity{
					Category:   CategoryPersonName,
					Value:      text[match[2]:match[3]],
					Start:      match[2],
					End:        match[3],
					Confidence: 0.88,
				})
			}
		}
	}

	return filterOverlaps(results)
}

func isValidLuhn(number string) bool {
	sum := 0
	alternate := false
	for i := len(number) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}
		if alternate {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}
		sum += n
		alternate = !alternate
	}
	return sum%10 == 0
}

func isValidIBAN(iban string) bool {
	cleaned := strings.ToUpper(strings.ReplaceAll(iban, " ", ""))
	if len(cleaned) < 15 || len(cleaned) > 34 {
		return false
	}

	// Move first 4 characters to the end
	rearranged := cleaned[4:] + cleaned[:4]

	// Convert letters to numbers (A=10, B=11, ... Z=35)
	var sb strings.Builder
	for _, char := range rearranged {
		if char >= 'A' && char <= 'Z' {
			sb.WriteString(strconv.Itoa(int(char - 'A' + 10)))
		} else if char >= '0' && char <= '9' {
			sb.WriteRune(char)
		} else {
			return false
		}
	}

	// Calculate mod 97 using math/big
	n := new(big.Int)
	n, ok := n.SetString(sb.String(), 10)
	if !ok {
		return false
	}
	remainder := new(big.Int).Mod(n, big.NewInt(97))
	return remainder.Int64() == 1
}

func keepDigits(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// filterOverlaps sorts and removes entities that collide with higher-confidence matches
func filterOverlaps(entities []DetectedEntity) []DetectedEntity {
	if len(entities) <= 1 {
		return entities
	}

	// Sort by start index ascending, then by length descending
	for i := 0; i < len(entities)-1; i++ {
		for j := i + 1; j < len(entities); j++ {
			if entities[i].Start > entities[j].Start ||
				(entities[i].Start == entities[j].Start && (entities[i].End-entities[i].Start) < (entities[j].End-entities[j].Start)) {
				entities[i], entities[j] = entities[j], entities[i]
			}
		}
	}

	var filtered []DetectedEntity
	lastEnd := -1
	for _, e := range entities {
		if e.Start >= lastEnd {
			filtered = append(filtered, e)
			lastEnd = e.End
		}
	}

	return filtered
}
