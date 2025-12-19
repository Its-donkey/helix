// Command gqlproxy runs a proxy server to intercept and log Twitch GQL requests.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Its-donkey/helix/gql"
)

func main() {
	port := flag.Int("port", 19808, "Port to listen on")
	outputDir := flag.String("output", "./gql_captures", "Directory to save captured requests")
	verbose := flag.Bool("v", false, "Verbose output (show variables)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gqlproxy - Intercept and log Twitch GQL requests\n\n")
		fmt.Fprintf(os.Stderr, "Usage: gqlproxy [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nBrowser Setup:\n")
		fmt.Fprintf(os.Stderr, "  1. Configure browser to use HTTP proxy: localhost:19808\n")
		fmt.Fprintf(os.Stderr, "  2. Browse Twitch normally\n")
		fmt.Fprintf(os.Stderr, "  3. GQL requests will be logged and saved\n\n")
		fmt.Fprintf(os.Stderr, "Output:\n")
		fmt.Fprintf(os.Stderr, "  - Individual JSON files for each operation\n")
		fmt.Fprintf(os.Stderr, "  - operations.jsonl - All operations in JSONL format\n")
		fmt.Fprintf(os.Stderr, "  - proxy.log - Proxy log file\n")
	}

	flag.Parse()

	cfg := gql.ProxyConfig{
		ListenAddr: fmt.Sprintf(":%d", *port),
		OutputDir:  *outputDir,
		Verbose:    *verbose,
	}

	proxy, err := gql.NewProxy(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create proxy: %v\n", err)
		os.Exit(1)
	}
	defer proxy.Close()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Printf("\n\nShutting down...\n")
		fmt.Printf("Captured %d unique operations\n", proxy.CapturedCount())
		proxy.Close()
		os.Exit(0)
	}()

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Twitch GQL Proxy (HTTPS)                   ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Listening on: localhost:%-5d                                ║\n", *port)
	fmt.Printf("║  Output dir:   %-47s ║\n", *outputDir)
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  HTTPS Setup (required for intercepting Twitch):              ║")
	fmt.Println("║  1. Import CA certificate into browser:                       ║")
	fmt.Printf("║     %s\n", proxy.GetCAPath())
	fmt.Println("║  2. Set browser HTTP/HTTPS proxy to: localhost:" + fmt.Sprintf("%-5d", *port) + "          ║")
	fmt.Println("║  3. Browse Twitch normally                                    ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  macOS: Import ca.crt → Keychain → Always Trust               ║")
	fmt.Println("║  Chrome: Settings → Security → Manage certificates            ║")
	fmt.Println("║  Firefox: Settings → Privacy → View Certificates → Import    ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Legend: * = new operation                                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if err := proxy.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Proxy error: %v\n", err)
		os.Exit(1)
	}
}
