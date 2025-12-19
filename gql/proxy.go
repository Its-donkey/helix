package gql

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Proxy intercepts GQL requests and logs them.
type Proxy struct {
	listenAddr   string
	targetURL    *url.URL
	outputDir    string
	logFile      *os.File
	reverseProxy *httputil.ReverseProxy
	operations   map[string]bool
	opMu         sync.Mutex
	verbose      bool
	logger       *log.Logger

	// TLS/HTTPS support
	caCert     *x509.Certificate
	caKey      *rsa.PrivateKey
	certCache  map[string]*tls.Certificate
	certCacheMu sync.RWMutex
}

// ProxyConfig configures the proxy server.
type ProxyConfig struct {
	ListenAddr string // Address to listen on (default: ":19808")
	OutputDir  string // Directory to save captured requests (default: "./gql_captures")
	Verbose    bool   // Log verbose output
}

// NewProxy creates a new GQL proxy server.
func NewProxy(cfg ProxyConfig) (*Proxy, error) {
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":19808"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./gql_captures"
	}

	targetURL, err := url.Parse(TwitchGQLEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create log file
	logPath := filepath.Join(cfg.OutputDir, "proxy.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	p := &Proxy{
		listenAddr: cfg.ListenAddr,
		targetURL:  targetURL,
		outputDir:  cfg.OutputDir,
		logFile:    logFile,
		operations: make(map[string]bool),
		verbose:    cfg.Verbose,
		logger:     log.New(io.MultiWriter(os.Stdout, logFile), "[GQL Proxy] ", log.LstdFlags),
		certCache:  make(map[string]*tls.Certificate),
	}

	// Generate CA certificate for HTTPS interception
	if err := p.generateCA(); err != nil {
		return nil, fmt.Errorf("failed to generate CA: %w", err)
	}

	// Create reverse proxy for forwarding
	p.reverseProxy = &httputil.ReverseProxy{
		Director:       p.director,
		ModifyResponse: p.modifyResponse,
		ErrorHandler:   p.errorHandler,
	}

	return p, nil
}

// generateCA generates a self-signed CA certificate for MITM.
func (p *Proxy) generateCA() error {
	// Generate RSA key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	p.caKey = key

	// Create CA certificate
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"GQL Proxy CA"},
			CommonName:   "GQL Proxy Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return err
	}
	p.caCert = cert

	// Save CA certificate for browser import
	caPath := filepath.Join(p.outputDir, "ca.crt")
	caFile, err := os.Create(caPath)
	if err != nil {
		return err
	}
	defer caFile.Close()

	pem.Encode(caFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	p.logger.Printf("CA certificate saved to: %s", caPath)
	p.logger.Printf("Import this certificate into your browser's trusted roots")

	return nil
}

// generateCert generates a certificate for a specific host.
func (p *Proxy) generateCert(host string) (*tls.Certificate, error) {
	// Check cache first
	p.certCacheMu.RLock()
	if cert, ok := p.certCache[host]; ok {
		p.certCacheMu.RUnlock()
		return cert, nil
	}
	p.certCacheMu.RUnlock()

	// Generate new certificate
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"GQL Proxy"},
			CommonName:   host,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 1, 0),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{host},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, p.caCert, &key.PublicKey, p.caKey)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	// Cache it
	p.certCacheMu.Lock()
	p.certCache[host] = cert
	p.certCacheMu.Unlock()

	return cert, nil
}

// Start starts the proxy server.
func (p *Proxy) Start() error {
	p.logger.Printf("Starting proxy on %s", p.listenAddr)
	p.logger.Printf("Saving captures to: %s", p.outputDir)

	server := &http.Server{
		Addr:    p.listenAddr,
		Handler: http.HandlerFunc(p.handleRequest),
	}

	return server.ListenAndServe()
}

// Close closes the proxy and its resources.
func (p *Proxy) Close() error {
	if p.logFile != nil {
		return p.logFile.Close()
	}
	return nil
}

// handleRequest handles both HTTP and HTTPS CONNECT requests.
func (p *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleConnect(w, r)
	} else {
		p.handleHTTP(w, r)
	}
}

// handleConnect handles HTTPS CONNECT tunneling.
func (p *Proxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if !strings.Contains(host, ":") {
		host += ":443"
	}

	// Only intercept Twitch GQL traffic
	if !strings.Contains(host, "gql.twitch.tv") && !strings.Contains(host, "twitch.tv") {
		// For non-Twitch hosts, just tunnel through
		p.tunnelConnect(w, r, host)
		return
	}

	// Hijack the connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Send 200 Connection Established
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Get host without port for certificate
	hostOnly := strings.Split(host, ":")[0]

	// Generate certificate for this host
	cert, err := p.generateCert(hostOnly)
	if err != nil {
		p.logger.Printf("Failed to generate cert for %s: %v", hostOnly, err)
		clientConn.Close()
		return
	}

	// Wrap client connection with TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}
	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	if err := tlsConn.Handshake(); err != nil {
		if p.verbose {
			p.logger.Printf("TLS handshake failed for %s: %v", host, err)
		}
		return
	}

	// Read requests from the TLS connection
	reader := bufio.NewReader(tlsConn)
	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				if p.verbose {
					p.logger.Printf("Error reading request: %v", err)
				}
			}
			return
		}

		// Set the full URL
		req.URL.Scheme = "https"
		req.URL.Host = host

		// Read and capture body
		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Capture the request
		p.captureRequest(req, bodyBytes)

		// Forward to the real server
		resp, err := p.forwardRequest(req, bodyBytes)
		if err != nil {
			p.logger.Printf("Forward error: %v", err)
			return
		}

		// Write response back to client
		resp.Write(tlsConn)
		resp.Body.Close()
	}
}

// tunnelConnect creates a transparent tunnel for non-intercepted hosts.
func (p *Proxy) tunnelConnect(w http.ResponseWriter, r *http.Request, host string) {
	destConn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Bidirectional copy
	go io.Copy(destConn, clientConn)
	io.Copy(clientConn, destConn)
}

// forwardRequest forwards a request to the actual server.
func (p *Proxy) forwardRequest(req *http.Request, body []byte) (*http.Response, error) {
	// Create new request
	outReq, err := http.NewRequest(req.Method, req.URL.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Copy headers
	for key, values := range req.Header {
		for _, value := range values {
			outReq.Header.Add(key, value)
		}
	}

	// Use custom transport for TLS
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	return client.Do(outReq)
}

// handleHTTP handles regular HTTP requests (non-CONNECT).
func (p *Proxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		p.handleCORS(w)
		return
	}

	// Capture request body
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Log and save the request
	p.captureRequest(r, bodyBytes)

	// Add CORS headers to response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	// Forward to Twitch
	p.reverseProxy.ServeHTTP(w, r)
}

func (p *Proxy) director(r *http.Request) {
	if r.URL.Host == "" {
		r.URL.Scheme = p.targetURL.Scheme
		r.URL.Host = p.targetURL.Host
		r.URL.Path = p.targetURL.Path
		r.Host = p.targetURL.Host
	}
}

func (p *Proxy) modifyResponse(resp *http.Response) error {
	resp.Header.Set("Access-Control-Allow-Origin", "*")
	resp.Header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	resp.Header.Set("Access-Control-Allow-Headers", "*")
	return nil
}

func (p *Proxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	p.logger.Printf("Proxy error: %v", err)
	http.Error(w, err.Error(), http.StatusBadGateway)
}

func (p *Proxy) handleCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.WriteHeader(http.StatusNoContent)
}

func (p *Proxy) captureRequest(r *http.Request, body []byte) {
	if len(body) == 0 {
		return
	}

	// Only process GQL requests
	if !strings.Contains(r.URL.String(), "gql") && !strings.Contains(r.Host, "gql") {
		return
	}

	// Try to parse as JSON
	var requests []json.RawMessage
	if err := json.Unmarshal(body, &requests); err != nil {
		// Try single request
		var single json.RawMessage
		if err := json.Unmarshal(body, &single); err != nil {
			if p.verbose {
				p.logger.Printf("Failed to parse request body as JSON")
			}
			return
		}
		requests = []json.RawMessage{single}
	}

	for _, rawReq := range requests {
		p.processRequest(rawReq)
	}
}

func (p *Proxy) processRequest(raw json.RawMessage) {
	var req struct {
		OperationName string                 `json:"operationName"`
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
		Extensions    *struct {
			PersistedQuery *struct {
				Version    int    `json:"version"`
				SHA256Hash string `json:"sha256Hash"`
			} `json:"persistedQuery"`
		} `json:"extensions"`
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return
	}

	opName := req.OperationName
	if opName == "" {
		opName = "anonymous"
	}

	// Check if we've seen this operation
	p.opMu.Lock()
	isNew := !p.operations[opName]
	p.operations[opName] = true
	p.opMu.Unlock()

	// Log the operation
	timestamp := time.Now().Format("15:04:05")
	marker := " "
	if isNew {
		marker = "*"
	}

	var hashInfo string
	if req.Extensions != nil && req.Extensions.PersistedQuery != nil {
		hash := req.Extensions.PersistedQuery.SHA256Hash
		if len(hash) > 16 {
			hashInfo = fmt.Sprintf(" [hash: %s...]", hash[:16])
		}
	}

	p.logger.Printf("%s [%s] %s%s", marker, timestamp, opName, hashInfo)

	if p.verbose && len(req.Variables) > 0 {
		varsJSON, _ := json.Marshal(req.Variables)
		p.logger.Printf("  Variables: %s", string(varsJSON))
	}

	// Save to file
	p.saveOperation(opName, raw, req.Query, req.Extensions)
}

func (p *Proxy) saveOperation(opName string, raw json.RawMessage, query string, extensions interface{}) {
	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", sanitizeFilename(opName), timestamp)
	savePath := filepath.Join(p.outputDir, filename)

	// Format the JSON nicely
	var formatted bytes.Buffer
	if err := json.Indent(&formatted, raw, "", "  "); err != nil {
		formatted.Write(raw)
	}

	if err := os.WriteFile(savePath, formatted.Bytes(), 0644); err != nil {
		p.logger.Printf("Failed to save operation: %v", err)
		return
	}

	// Also save to a combined operations file
	p.appendToOperationsLog(opName, raw, query, extensions)
}

func (p *Proxy) appendToOperationsLog(opName string, raw json.RawMessage, query string, extensions interface{}) {
	opsFile := filepath.Join(p.outputDir, "operations.jsonl")
	f, err := os.OpenFile(opsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	entry := map[string]interface{}{
		"timestamp":     time.Now().Format(time.RFC3339),
		"operationName": opName,
		"request":       raw,
	}

	line, _ := json.Marshal(entry)
	f.Write(line)
	f.WriteString("\n")
}

func sanitizeFilename(s string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(s)
}

// GetCapturedOperations returns the list of captured operation names.
func (p *Proxy) GetCapturedOperations() []string {
	p.opMu.Lock()
	defer p.opMu.Unlock()

	ops := make([]string, 0, len(p.operations))
	for op := range p.operations {
		ops = append(ops, op)
	}
	return ops
}

// CapturedCount returns the number of unique operations captured.
func (p *Proxy) CapturedCount() int {
	p.opMu.Lock()
	defer p.opMu.Unlock()
	return len(p.operations)
}

// DecompressGzip decompresses gzipped data (useful for Spade events).
func DecompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// GetCAPath returns the path to the CA certificate file.
func (p *Proxy) GetCAPath() string {
	return filepath.Join(p.outputDir, "ca.crt")
}
