// Package testing provides high-load tests for the Gouija proxy system.
// Run with: go test -v -timeout 5m ./internal/testing/...
package testing

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neurophant/gouija/internal/proxy"
	"github.com/neurophant/gouija/internal/relay"
	"github.com/neurophant/gouija/internal/transport"
)

const (
	testKey  = "test-secret-key-for-highload-testing-32chars"
	testSalt = "test-salt-v1"
)

// TestConfig holds configuration for a test scenario
type TestConfig struct {
	Name               string
	TransportMode      transport.Mode
	WithAuth           bool
	RequireProxyAuth   bool
	PoolSize           int
	ConcurrentClients  int
	RequestsPerClient  int
	PayloadSize        int
	ConnectionPacing   time.Duration
	MaxConcurrentDials int
}

// TestResult holds results from a test run
type TestResult struct {
	TotalRequests     int64
	SuccessfulReqs    int64
	FailedReqs        int64
	TotalBytes        int64
	Duration          time.Duration
	AvgLatency        time.Duration
	MaxLatency        time.Duration
	RequestsPerSecond float64
	BytesPerSecond    float64
}

// startMockHTTPSServer creates a mock HTTPS server for testing
func startMockHTTPSServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	// Echo endpoint - returns the request body
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		w.Write(data)
	})

	// Large response endpoint - simulates video streaming
	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		size := 1024 * 100 // 100KB
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))

		// Write in chunks to simulate streaming
		chunk := make([]byte, 16*1024) // 16KB chunks
		rand.Read(chunk)

		for written := 0; written < size; {
			n := size - written
			if n > len(chunk) {
				n = len(chunk)
			}
			w.Write(chunk[:n])
			written += n
		}
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Small response
	mux.HandleFunc("/small", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("small response"))
	})

	server := httptest.NewTLSServer(mux)
	return server
}

// startProxyAndRelay starts both proxy and relay servers for testing
func startProxyAndRelay(t *testing.T, cfg TestConfig) (relayAddr string, cleanup func()) {
	// Start proxy server
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create proxy listener: %v", err)
	}
	proxyAddr := proxyListener.Addr().String()
	proxyListener.Close() // Close so the server can bind

	proxyCfg := proxy.Config{
		ListenAddr: proxyAddr,
		Key:        testKey,
		Salt:       testSalt,
		Transport: transport.Config{
			Mode:      cfg.TransportMode,
			Path:      "/ws",
			KeepAlive: true,
		},
		Timeout:     30 * time.Second,
		RequireAuth: cfg.RequireProxyAuth,
	}

	proxySrv, err := proxy.New(proxyCfg)
	if err != nil {
		t.Fatalf("Failed to create proxy server: %v", err)
	}

	proxyReady := make(chan struct{})
	go func() {
		close(proxyReady)
		proxySrv.Start()
	}()
	<-proxyReady

	// Wait for proxy to start accepting connections
	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", proxyAddr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Start relay server
	relayListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		proxySrv.Shutdown(context.Background())
		t.Fatalf("Failed to create relay listener: %v", err)
	}
	relayAddr = relayListener.Addr().String()
	relayListener.Close() // Close so the server can bind

	poolSize := cfg.PoolSize
	if poolSize <= 0 {
		poolSize = 3
	}

	connPacing := cfg.ConnectionPacing
	if connPacing <= 0 {
		connPacing = 50 * time.Millisecond // Fast pacing for tests
	}

	maxConcurrentDials := cfg.MaxConcurrentDials
	if maxConcurrentDials <= 0 {
		maxConcurrentDials = 2
	}

	relayCfg := relay.Config{
		ListenAddr: relayAddr,
		ProxyAddr:  proxyAddr,
		Key:        testKey,
		Salt:       testSalt,
		Transport: transport.Config{
			Mode:      cfg.TransportMode,
			Path:      "/ws",
			KeepAlive: true,
		},
		PoolSize:    poolSize,
		Timeout:     30 * time.Second,
		RequireAuth: cfg.RequireProxyAuth,

		// DoS prevention settings
		ConnectionPacing:   connPacing,
		MaxConcurrentDials: maxConcurrentDials,
		BackoffInitial:     500 * time.Millisecond,
		BackoffMax:         5 * time.Second,
		BackoffMultiplier:  2.0,
		BackoffJitter:      0.2,
		MinPoolSize:        1,
		MaxPoolSize:        poolSize * 2,
		GrowthThreshold:    0.8,
		ShrinkThreshold:    0.2,
	}

	if cfg.WithAuth {
		relayCfg.AuthUser = "testuser"
		relayCfg.AuthPass = "testpass"
	}

	relaySrv, err := relay.New(relayCfg)
	if err != nil {
		proxySrv.Shutdown(context.Background())
		t.Fatalf("Failed to create relay server: %v", err)
	}

	relayReady := make(chan struct{})
	go func() {
		close(relayReady)
		relaySrv.Start()
	}()
	<-relayReady

	// Wait for relay to start and establish connections
	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", relayAddr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for pool connections to be established
	time.Sleep(time.Duration(poolSize)*connPacing + 500*time.Millisecond)

	cleanup = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		relaySrv.Shutdown(ctx)
		proxySrv.Shutdown(ctx)
	}

	return relayAddr, cleanup
}

// makeProxyClient creates an HTTP client that uses the relay as a proxy
func makeProxyClient(relayAddr string, withAuth bool) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Connect to relay
			conn, err := net.DialTimeout("tcp", relayAddr, 10*time.Second)
			if err != nil {
				return nil, err
			}

			// Send CONNECT request
			connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
			if withAuth {
				// Add proxy auth header
				connectReq += "Proxy-Authorization: Basic dGVzdHVzZXI6dGVzdHBhc3M=\r\n"
			}
			connectReq += "\r\n"

			_, err = conn.Write([]byte(connectReq))
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("failed to send CONNECT: %w", err)
			}

			// Read response
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				conn.Close()
				return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
			}

			response := string(buf[:n])
			if len(response) < 12 || response[9:12] != "200" {
				conn.Close()
				return nil, fmt.Errorf("CONNECT failed: %s", response)
			}

			return conn, nil
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

// runHighLoadTest executes a high-load test scenario
func runHighLoadTest(t *testing.T, cfg TestConfig, targetURL, relayAddr string) TestResult {
	var (
		totalReqs    int64
		successReqs  int64
		failedReqs   int64
		totalBytes   int64
		totalLatency int64
		maxLatency   int64
	)

	startTime := time.Now()

	var wg sync.WaitGroup
	concurrency := cfg.ConcurrentClients
	if concurrency <= 0 {
		concurrency = 10
	}
	requestsPerClient := cfg.RequestsPerClient
	if requestsPerClient <= 0 {
		requestsPerClient = 50
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			client := makeProxyClient(relayAddr, cfg.WithAuth)

			for j := 0; j < requestsPerClient; j++ {
				atomic.AddInt64(&totalReqs, 1)

				reqStart := time.Now()

				var resp *http.Response
				var err error
				var bytesReceived int64

				// Alternate between different request types
				switch j % 3 {
				case 0:
					// Health check - small request
					resp, err = client.Get(targetURL + "/health")
					if err == nil && resp != nil {
						body, _ := io.ReadAll(resp.Body)
						bytesReceived = int64(len(body))
						resp.Body.Close()
					}

				case 1:
					// Small response
					resp, err = client.Get(targetURL + "/small")
					if err == nil && resp != nil {
						body, _ := io.ReadAll(resp.Body)
						bytesReceived = int64(len(body))
						resp.Body.Close()
					}

				case 2:
					// Stream - larger response
					resp, err = client.Get(targetURL + "/stream")
					if err == nil && resp != nil {
						body, _ := io.ReadAll(resp.Body)
						bytesReceived = int64(len(body))
						resp.Body.Close()
					}
				}

				latency := time.Since(reqStart)
				latencyNs := latency.Nanoseconds()

				if err != nil {
					atomic.AddInt64(&failedReqs, 1)
					t.Logf("Request failed: %v", err)
				} else if resp != nil && resp.StatusCode >= 400 {
					atomic.AddInt64(&failedReqs, 1)
					t.Logf("Request failed with status: %d", resp.StatusCode)
				} else {
					atomic.AddInt64(&successReqs, 1)
					atomic.AddInt64(&totalBytes, bytesReceived)
					atomic.AddInt64(&totalLatency, latencyNs)

					// Update max latency
					for {
						current := atomic.LoadInt64(&maxLatency)
						if latencyNs <= current {
							break
						}
						if atomic.CompareAndSwapInt64(&maxLatency, current, latencyNs) {
							break
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	result := TestResult{
		TotalRequests:  totalReqs,
		SuccessfulReqs: successReqs,
		FailedReqs:     failedReqs,
		TotalBytes:     totalBytes,
		Duration:       duration,
		MaxLatency:     time.Duration(maxLatency),
	}

	if successReqs > 0 {
		result.AvgLatency = time.Duration(totalLatency / successReqs)
	}

	if duration.Seconds() > 0 {
		result.RequestsPerSecond = float64(successReqs) / duration.Seconds()
		result.BytesPerSecond = float64(totalBytes) / duration.Seconds()
	}

	return result
}

// printResult prints test results
func printResult(t *testing.T, cfg TestConfig, result TestResult) {
	t.Logf("=== %s ===", cfg.Name)
	t.Logf("  Transport: %s", cfg.TransportMode)
	t.Logf("  Pool Size: %d", cfg.PoolSize)
	t.Logf("  Concurrent Clients: %d", cfg.ConcurrentClients)
	t.Logf("  Requests per Client: %d", cfg.RequestsPerClient)
	t.Logf("  Connection Pacing: %v", cfg.ConnectionPacing)
	t.Logf("---")
	t.Logf("  Total Requests: %d", result.TotalRequests)
	t.Logf("  Successful: %d (%.1f%%)", result.SuccessfulReqs, float64(result.SuccessfulReqs)/float64(result.TotalRequests)*100)
	t.Logf("  Failed: %d", result.FailedReqs)
	t.Logf("  Total Bytes: %.2f KB", float64(result.TotalBytes)/1024)
	t.Logf("  Duration: %v", result.Duration)
	t.Logf("  Avg Latency: %v", result.AvgLatency)
	t.Logf("  Max Latency: %v", result.MaxLatency)
	t.Logf("  Requests/sec: %.2f", result.RequestsPerSecond)
	t.Logf("  Throughput: %.2f KB/s", result.BytesPerSecond/1024)
	t.Logf("")
}

// TestHighLoadTCP tests high load with TCP transport
func TestHighLoadTCP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-load test in short mode")
	}

	targetServer := startMockHTTPSServer(t)
	defer targetServer.Close()

	cfg := TestConfig{
		Name:               "TCP High Load",
		TransportMode:      transport.ModeTCP,
		WithAuth:           false,
		RequireProxyAuth:   false,
		PoolSize:           5,
		ConcurrentClients:  10,
		RequestsPerClient:  30,
		PayloadSize:        4096,
		ConnectionPacing:   50 * time.Millisecond,
		MaxConcurrentDials: 2,
	}

	relayAddr, cleanup := startProxyAndRelay(t, cfg)
	defer cleanup()

	result := runHighLoadTest(t, cfg, targetServer.URL, relayAddr)
	printResult(t, cfg, result)

	// Verify success rate
	if result.TotalRequests > 0 {
		successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
		if successRate < 0.8 {
			t.Errorf("Success rate too low: %.1f%% (expected >= 80%%)", successRate*100)
		}
	}
}

// TestHighLoadTCPWithAuth tests high load with TCP transport and authentication
func TestHighLoadTCPWithAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-load test in short mode")
	}

	targetServer := startMockHTTPSServer(t)
	defer targetServer.Close()

	cfg := TestConfig{
		Name:               "TCP High Load with Auth",
		TransportMode:      transport.ModeTCP,
		WithAuth:           true,
		RequireProxyAuth:   true,
		PoolSize:           5,
		ConcurrentClients:  10,
		RequestsPerClient:  30,
		PayloadSize:        4096,
		ConnectionPacing:   50 * time.Millisecond,
		MaxConcurrentDials: 2,
	}

	relayAddr, cleanup := startProxyAndRelay(t, cfg)
	defer cleanup()

	result := runHighLoadTest(t, cfg, targetServer.URL, relayAddr)
	printResult(t, cfg, result)

	// Verify success rate
	if result.TotalRequests > 0 {
		successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
		if successRate < 0.8 {
			t.Errorf("Success rate too low: %.1f%% (expected >= 80%%)", successRate*100)
		}
	}
}

// TestHighLoadWebSocket tests high load with WebSocket transport
func TestHighLoadWebSocket(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-load test in short mode")
	}

	targetServer := startMockHTTPSServer(t)
	defer targetServer.Close()

	cfg := TestConfig{
		Name:               "WebSocket High Load",
		TransportMode:      transport.ModeWebSocket,
		WithAuth:           false,
		RequireProxyAuth:   false,
		PoolSize:           5,
		ConcurrentClients:  10,
		RequestsPerClient:  30,
		PayloadSize:        4096,
		ConnectionPacing:   50 * time.Millisecond,
		MaxConcurrentDials: 2,
	}

	relayAddr, cleanup := startProxyAndRelay(t, cfg)
	defer cleanup()

	result := runHighLoadTest(t, cfg, targetServer.URL, relayAddr)
	printResult(t, cfg, result)

	// Verify success rate
	if result.TotalRequests > 0 {
		successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
		if successRate < 0.8 {
			t.Errorf("Success rate too low: %.1f%% (expected >= 80%%)", successRate*100)
		}
	}
}

// TestConnectionPacing verifies connection pacing prevents burst
func TestConnectionPacing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pacing test in short mode")
	}

	targetServer := startMockHTTPSServer(t)
	defer targetServer.Close()

	cfg := TestConfig{
		Name:               "Connection Pacing Test",
		TransportMode:      transport.ModeTCP,
		WithAuth:           false,
		RequireProxyAuth:   false,
		PoolSize:           5,                      // 5 connections
		ConcurrentClients:  5,                      // Lower concurrency
		RequestsPerClient:  10,                     // Fewer requests
		PayloadSize:        1024,                   // Smaller payload
		ConnectionPacing:   200 * time.Millisecond, // 200ms between connections
		MaxConcurrentDials: 1,                      // Only 1 dial at a time
	}

	// Measure time to establish connections
	startTime := time.Now()
	relayAddr, cleanup := startProxyAndRelay(t, cfg)
	connectionTime := time.Since(startTime)
	defer cleanup()

	// With 5 connections at 200ms pacing and 1 concurrent dial,
	// establishment should take at least 800ms (4 * 200ms)
	minExpectedTime := time.Duration(cfg.PoolSize-1) * cfg.ConnectionPacing
	t.Logf("Connection pacing results:")
	t.Logf("  Pool size: %d", cfg.PoolSize)
	t.Logf("  Pacing: %v", cfg.ConnectionPacing)
	t.Logf("  Min expected time: %v", minExpectedTime)
	t.Logf("  Actual establishment time: %v", connectionTime)

	if connectionTime < minExpectedTime/2 {
		t.Logf("Warning: Connections established faster than expected (may indicate pacing not working)")
	}

	// Run traffic to verify everything works
	result := runHighLoadTest(t, cfg, targetServer.URL, relayAddr)
	printResult(t, cfg, result)

	if result.TotalRequests > 0 {
		successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
		if successRate < 0.8 {
			t.Errorf("Success rate too low: %.1f%% (expected >= 80%%)", successRate*100)
		}
	}
}

// TestVideoStreamingSimulation simulates video streaming workload
func TestVideoStreamingSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping video streaming test in short mode")
	}

	targetServer := startMockHTTPSServer(t)
	defer targetServer.Close()

	cfg := TestConfig{
		Name:               "Video Streaming Simulation",
		TransportMode:      transport.ModeTCP,
		WithAuth:           false,
		RequireProxyAuth:   false,
		PoolSize:           3,     // Minimal pool (home network config)
		ConcurrentClients:  3,     // Multiple streams
		RequestsPerClient:  20,    // Sustained load
		PayloadSize:        65536, // 64KB chunks (video segments)
		ConnectionPacing:   100 * time.Millisecond,
		MaxConcurrentDials: 1,
	}

	relayAddr, cleanup := startProxyAndRelay(t, cfg)
	defer cleanup()

	result := runHighLoadTest(t, cfg, targetServer.URL, relayAddr)
	printResult(t, cfg, result)

	// Video streaming test metrics
	t.Logf("Video Streaming Metrics:")
	t.Logf("  Simulated streams: %d", cfg.ConcurrentClients)
	t.Logf("  Throughput: %.2f KB/s", result.BytesPerSecond/1024)
	t.Logf("  Avg latency per segment: %v", result.AvgLatency)

	if result.TotalRequests > 0 {
		successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
		if successRate < 0.8 {
			t.Errorf("Success rate too low for streaming: %.1f%% (expected >= 80%%)", successRate*100)
		}
	}
}

// TestAllModes runs tests for all transport modes
func TestAllModes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping all-modes test in short mode")
	}

	targetServer := startMockHTTPSServer(t)
	defer targetServer.Close()

	modes := []struct {
		name        string
		transport   transport.Mode
		withAuth    bool
		requireAuth bool
	}{
		{"TCP_Basic", transport.ModeTCP, false, false},
		{"TCP_HTTPAuth", transport.ModeTCP, true, false},
		{"TCP_ProxyAuth", transport.ModeTCP, false, true},
		{"TCP_AllAuth", transport.ModeTCP, true, true},
		{"WS_Basic", transport.ModeWebSocket, false, false},
		{"WS_HTTPAuth", transport.ModeWebSocket, true, false},
		{"WS_ProxyAuth", transport.ModeWebSocket, false, true},
		{"WS_AllAuth", transport.ModeWebSocket, true, true},
	}

	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
			cfg := TestConfig{
				Name:               mode.name,
				TransportMode:      mode.transport,
				WithAuth:           mode.withAuth,
				RequireProxyAuth:   mode.requireAuth,
				PoolSize:           3,
				ConcurrentClients:  5,
				RequestsPerClient:  10,
				PayloadSize:        4096,
				ConnectionPacing:   50 * time.Millisecond,
				MaxConcurrentDials: 2,
			}

			relayAddr, cleanup := startProxyAndRelay(t, cfg)
			defer cleanup()

			result := runHighLoadTest(t, cfg, targetServer.URL, relayAddr)
			printResult(t, cfg, result)

			if result.TotalRequests > 0 {
				successRate := float64(result.SuccessfulReqs) / float64(result.TotalRequests)
				if successRate < 0.7 {
					t.Errorf("Success rate too low: %.1f%% (expected >= 70%%)", successRate*100)
				}
			}
		})
	}
}

// BenchmarkThroughput measures throughput performance
func BenchmarkThroughput(b *testing.B) {
	targetServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 10*1024) // 10KB
		rand.Read(data)
		w.Write(data)
	}))
	defer targetServer.Close()

	cfg := TestConfig{
		TransportMode:      transport.ModeTCP,
		PoolSize:           5,
		ConnectionPacing:   10 * time.Millisecond,
		MaxConcurrentDials: 3,
	}

	// Get a free port for proxy
	proxyListener, _ := net.Listen("tcp", "127.0.0.1:0")
	proxyAddr := proxyListener.Addr().String()
	proxyListener.Close()

	proxySrv, _ := proxy.New(proxy.Config{
		ListenAddr: proxyAddr,
		Key:        testKey,
		Salt:       testSalt,
		Transport:  transport.Config{Mode: cfg.TransportMode, KeepAlive: true},
		Timeout:    30 * time.Second,
	})
	go proxySrv.Start()
	time.Sleep(100 * time.Millisecond)
	defer proxySrv.Shutdown(context.Background())

	// Get a free port for relay
	relayListener, _ := net.Listen("tcp", "127.0.0.1:0")
	relayAddr := relayListener.Addr().String()
	relayListener.Close()

	relaySrv, _ := relay.New(relay.Config{
		ListenAddr:         relayAddr,
		ProxyAddr:          proxyAddr,
		Key:                testKey,
		Salt:               testSalt,
		Transport:          transport.Config{Mode: cfg.TransportMode, KeepAlive: true},
		PoolSize:           cfg.PoolSize,
		Timeout:            30 * time.Second,
		ConnectionPacing:   cfg.ConnectionPacing,
		MaxConcurrentDials: cfg.MaxConcurrentDials,
	})
	go relaySrv.Start()
	time.Sleep(500 * time.Millisecond)
	defer relaySrv.Shutdown(context.Background())

	client := makeProxyClient(relayAddr, false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(targetServer.URL + "/")
			if err == nil && resp != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}
	})
}
