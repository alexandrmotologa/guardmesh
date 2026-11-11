package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexandrmotologa/guardmesh/internal/classifier"
	"github.com/alexandrmotologa/guardmesh/internal/config"
	"github.com/alexandrmotologa/guardmesh/internal/dossier"
	"github.com/alexandrmotologa/guardmesh/internal/pii"
	"github.com/alexandrmotologa/guardmesh/internal/proxy"
	"github.com/alexandrmotologa/guardmesh/internal/store"
	"github.com/alexandrmotologa/guardmesh/web"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "start":
		runStart(os.Args[2:])
	case "classify":
		runClassify(os.Args[2:])
	case "scan":
		runScan(os.Args[2:])
	case "dossier":
		runDossier(os.Args[2:])
	case "apps":
		runApps(os.Args[2:])
	case "benchmark":
		runBenchmark(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("guardmesh v%s (EU AI Act & PII Governance Gateway)\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`guardmesh v%s - Enterprise AI Governance Gateway & PII Anonymizer

USAGE:
  guardmesh <command> [flags]

COMMANDS:
  start       Start reverse proxy and embedded compliance dashboard
  classify    Audit a prompt and tool schema for EU AI Act risk classification
  scan        Inspect text or files for PII and developer secrets
  dossier     Generate an EU AI Act Annex IV compliance dossier in PDF / Markdown
  apps        Manage discovered applications routing traffic through the gateway
  benchmark   Run performance benchmark measuring PII masking latency overhead
  version     Show version information

Run 'guardmesh <command> --help' for details on a specific command.
`, version)
}

func runStart(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	configPath := fs.String("config", "guardmesh.yaml", "Path to guardmesh.yaml config file")
	port := fs.Int("port", 0, "Port to listen on (overrides config)")
	host := fs.String("host", "", "Host to bind to (overrides config)")
	dbPath := fs.String("db", "", "SQLite database path (overrides config)")
	_ = fs.Parse(args)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	if *port > 0 {
		cfg.Server.Port = *port
	}
	if *host != "" {
		cfg.Server.Host = *host
	}
	if *dbPath != "" {
		cfg.Server.DBPath = *dbPath
	}

	// Initialize SQLite Store
	s, err := store.Open(cfg.Server.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize store: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	// Get embedded frontend assets
	webFS, err := web.GetFS()
	if err != nil {
		fmt.Printf("Notice: Embedded web UI not accessible: %v\n", err)
	}

	srv, err := proxy.NewServer(cfg, s, webFS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize proxy server: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv,
	}

	fmt.Println("==================================================================")
	fmt.Printf(" guardmesh v%s - Enterprise AI Governance Gateway\n", version)
	fmt.Println("==================================================================")
	fmt.Printf("  • Proxy Gateway:     http://%s/v1/chat/completions\n", addr)
	fmt.Printf("  • Anthropic Adapter: http://%s/v1/messages\n", addr)
	fmt.Printf("  • Compliance Cockpit:http://%s/\n", addr)
	fmt.Printf("  • SQLite WAL Store:  %s\n", cfg.Server.DBPath)
	fmt.Printf("  • Upstream OpenAI:   %s\n", cfg.Server.UpstreamOpenAI)
	fmt.Printf("  • Upstream Anthropic:%s\n", cfg.Server.UpstreamAnthropic)
	fmt.Printf("  • PII Anonymization: %t (Mode: %s)\n", cfg.PII.Enabled, cfg.PII.MaskMode)
	fmt.Printf("  • EU AI Act Guard:   Active (Article 5 Hard-Kill: %t)\n", cfg.Governance.EnforceArticle5Block)
	fmt.Println("==================================================================")
	fmt.Println("Press Ctrl+C to terminate.")

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down guardmesh daemon...")
}

func runClassify(args []string) {
	fs := flag.NewFlagSet("classify", flag.ExitOnError)
	promptText := fs.String("prompt", "", "Prompt text to evaluate")
	promptFile := fs.String("prompt-file", "", "Path to text file containing prompt")
	appID := fs.String("app-id", "auditor-cli", "Application ID for policy evaluation")
	_ = fs.Parse(args)

	text := *promptText
	if text == "" && *promptFile != "" {
		data, err := os.ReadFile(*promptFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read prompt file: %v\n", err)
			os.Exit(1)
		}
		text = string(data)
	}

	if text == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			bytes, _ := io.ReadAll(os.Stdin)
			text = string(bytes)
		}
	}

	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(os.Stderr, "Error: No prompt provided. Use --prompt, --prompt-file, or pipe via stdin.")
		os.Exit(1)
	}

	cfg := config.DefaultConfig()
	policy := classifier.NewPolicyEngine(cfg)

	eval, injection := policy.CheckCompliance(*appID, text)

	fmt.Println("------------------------------------------------------------------")
	fmt.Println(" EU AI Act Compliance Evaluation (Regulation EU 2024/1689)")
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf(" Risk Classification Tier: %s\n", eval.Tier)
	fmt.Printf(" Risk Severity Score:     %.2f / 1.00\n", eval.Score)
	fmt.Printf(" Automated Hard Block:    %t\n", eval.IsBlocked)
	if eval.IsBlocked {
		fmt.Printf(" Block Reason:            %s\n", eval.BlockedReason)
	}

	fmt.Println("\n Applicable Legal Articles:")
	for _, art := range eval.ApplicableArticles {
		fmt.Printf("  • %s\n", art)
	}

	fmt.Println("\n Classification Reasons & Findings:")
	for _, reason := range eval.Reasons {
		fmt.Printf("  • %s\n", reason)
	}

	fmt.Println("\n Required Compliance Controls:")
	for _, ctrl := range eval.RequiredControls {
		fmt.Printf("  [X] %s\n", ctrl)
	}

	if injection.Detected {
		fmt.Printf("\n ⚠ Security Alert: Prompt injection pattern detected ('%s')\n", injection.Pattern)
	} else {
		fmt.Println("\n ✓ Security Check: No adversarial prompt injection detected")
	}
	fmt.Println("------------------------------------------------------------------")
}

func runScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	scanText := fs.String("text", "", "Text string to scan for PII")
	scanFile := fs.String("file", "", "Path to file to scan")
	mode := fs.String("mode", "token", "Mask mode: token, redact, hash")
	_ = fs.Parse(args)

	text := *scanText
	if text == "" && *scanFile != "" {
		data, err := os.ReadFile(*scanFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
			os.Exit(1)
		}
		text = string(data)
	}

	if text == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			bytes, _ := io.ReadAll(os.Stdin)
			text = string(bytes)
		}
	}

	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(os.Stderr, "Error: No text provided to scan. Use --text, --file, or pipe via stdin.")
		os.Exit(1)
	}

	cfg := config.DefaultConfig()
	detector := pii.NewDetector(cfg.PII.Detectors)
	vault, _ := pii.NewVault(time.Minute)
	defer vault.Close()

	entities := detector.Detect(text)
	masked, count, categories := vault.MaskText("cli-scan", text, detector, *mode)

	fmt.Println("------------------------------------------------------------------")
	fmt.Printf(" guardmesh PII & Secret Detection: %d Entities Found\n", count)
	fmt.Println("------------------------------------------------------------------")
	for _, ent := range entities {
		fmt.Printf("  • [%-12s] %-30s (Confidence: %.0f%%)\n", ent.Category, ent.Value, ent.Confidence*100)
	}
	fmt.Printf("\n Detected Categories: %s\n", strings.Join(categories, ", "))
	fmt.Println("\n Sanitized Outbound Payload (Sent to Cloud LLMs):")
	fmt.Println("------------------------------------------------------------------")
	fmt.Println(masked)
	fmt.Println("------------------------------------------------------------------")
}

func runDossier(args []string) {
	fs := flag.NewFlagSet("dossier", flag.ExitOnError)
	appID := fs.String("app-id", "enterprise-app", "Application identifier to generate dossier for")
	outputPath := fs.String("output", "", "Output PDF file path (e.g. ./dossier.pdf)")
	dbPath := fs.String("db", "./guardmesh.db", "SQLite database path")
	format := fs.String("format", "pdf", "Format: pdf, md, or both")
	_ = fs.Parse(args)

	cfg := config.DefaultConfig()
	cfg.Server.DBPath = *dbPath

	s, err := store.Open(cfg.Server.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not open database (%v), generating dossier from static profile\n", err)
	} else {
		defer s.Close()
	}

	var app *store.Application
	var events []store.AuditEvent
	if s != nil {
		app, _ = s.GetApplication(*appID)
		events, _ = s.QueryAuditEvents(store.AuditFilter{AppID: *appID, Limit: 10})
	}

	if app == nil {
		app = &store.Application{
			ID:               *appID,
			Name:             *appID,
			HighestRiskTier:  "HIGH_RISK",
			ComplianceStatus: "COMPLIANT",
			TotalRequests:    1,
			TotalPIIMasked:   1,
		}
	}

	policy := classifier.NewPolicyEngine(cfg)
	eval, _ := policy.CheckCompliance(*appID, app.Name)

	meta := dossier.DossierMetadata{
		AppID:              app.ID,
		SystemName:         app.Name,
		Organization:       "Enterprise Organization",
		Author:             "Compliance & AI Safety Lead",
		IntendedPurpose:    "Enterprise AI Production Application",
		ModelUsed:          "gpt-4o / claude-3-5-sonnet",
		GeneratedAt:        time.Now(),
		RiskEvaluation:     eval,
		TotalRequests:      app.TotalRequests,
		TotalPIIMasked:     app.TotalPIIMasked,
		ComplianceStatus:   app.ComplianceStatus,
		SampleAuditRecords: events,
	}

	if *outputPath == "" {
		*outputPath = fmt.Sprintf("EU_AI_Act_Annex_IV_%s.%s", *appID, *format)
	}

	if *format == "md" || strings.HasSuffix(*outputPath, ".md") {
		mdContent := dossier.GenerateAnnexIVMarkdown(meta)
		if err := os.WriteFile(*outputPath, []byte(mdContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write markdown dossier: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ EU AI Act Markdown dossier written to: %s\n", *outputPath)
		return
	}

	// PDF generation
	if err := dossier.GeneratePDF(meta, *outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate PDF dossier: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ EU AI Act Annex IV PDF technical dossier generated: %s\n", *outputPath)
}

func runApps(args []string) {
	if len(args) == 0 || args[0] != "list" {
		fmt.Println("Usage: guardmesh apps list [--db ./guardmesh.db]")
		return
	}

	fs := flag.NewFlagSet("apps list", flag.ExitOnError)
	dbPath := fs.String("db", "./guardmesh.db", "SQLite database path")
	_ = fs.Parse(args[1:])

	s, err := store.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	apps, err := s.ListApplications()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list apps: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("--------------------------------------------------------------------------------------------")
	fmt.Printf("%-20s %-16s %-12s %-12s %-14s %-16s\n", "APP ID", "RISK TIER", "REQUESTS", "PII MASKED", "STATUS", "LAST SEEN")
	fmt.Println("--------------------------------------------------------------------------------------------")
	if len(apps) == 0 {
		fmt.Println("No applications recorded yet. Start the proxy and route traffic to populate.")
	} else {
		for _, a := range apps {
			lastSeen := a.LastSeen.Format("2006-01-02 15:04")
			fmt.Printf("%-20s %-16s %-12d %-12d %-14s %-16s\n",
				a.ID, a.HighestRiskTier, a.TotalRequests, a.TotalPIIMasked, a.ComplianceStatus, lastSeen)
		}
	}
	fmt.Println("--------------------------------------------------------------------------------------------")
}

func runBenchmark(args []string) {
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)
	requests := fs.Int("requests", 1000, "Number of de-identification benchmark runs")
	_ = fs.Parse(args)

	cfg := config.DefaultConfig()
	detector := pii.NewDetector(cfg.PII.Detectors)
	vault, err := pii.NewVault(time.Minute)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init vault: %v\n", err)
		os.Exit(1)
	}
	defer vault.Close()

	samplePayload := "Please process transaction for John Doe with email john.doe@enterprise.com, IBAN DE89370400440532013000, and SSN 123-45-6789. Also note secret sk-1234567890abcdef1234567890."

	fmt.Printf("Running guardmesh de-identification benchmark (%d iterations)...\n", *requests)
	start := time.Now()

	totalEntities := 0
	for i := 0; i < *requests; i++ {
		sessID := fmt.Sprintf("bench-sess-%d", i)
		masked, count, _ := vault.MaskText(sessID, samplePayload, detector, "token")
		totalEntities += count
		_ = vault.RehydrateText(sessID, masked)
	}

	duration := time.Since(start)
	avgMicro := float64(duration.Microseconds()) / float64(*requests)

	fmt.Println("------------------------------------------------------------------")
	fmt.Printf(" Total Iterations:           %d\n", *requests)
	fmt.Printf(" Total Entities Intercepted: %d\n", totalEntities)
	fmt.Printf(" Total Execution Time:       %v\n", duration)
	fmt.Printf(" Average Mask + Rehydrate:   %.2f µs (%.3f ms)\n", avgMicro, avgMicro/1000.0)
	fmt.Printf(" Throughput:                 %.0f operations/sec\n", float64(*requests)/duration.Seconds())
	fmt.Println("------------------------------------------------------------------")
}
