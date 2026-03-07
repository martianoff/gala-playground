package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed examples/*.gala
var exampleFiles embed.FS

type RunRequest struct {
	Code string `json:"code"`
}

type RunResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
	Time   string `json:"time"`
}

type Example struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func findGala() string {
	// Check common locations
	candidates := []string{
		filepath.Join(os.Getenv("USERPROFILE"), ".local", "bin", "gala.exe"),
		filepath.Join(os.Getenv("USERPROFILE"), ".local", "bin", "gala"),
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "gala.exe"),
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "gala"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	// Fallback to PATH
	if p, err := exec.LookPath("gala"); err == nil {
		return p
	}
	if p, err := exec.LookPath("gala.exe"); err == nil {
		return p
	}
	return ""
}

func handleRun(galaBin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RunRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, RunResponse{Error: "Invalid request"})
			return
		}

		if len(req.Code) > 50000 {
			writeJSON(w, RunResponse{Error: "Code too large (max 50KB)"})
			return
		}

		// Create temp directory with gala.mod
		tmpDir, err := os.MkdirTemp("", "gala-playground-*")
		if err != nil {
			writeJSON(w, RunResponse{Error: "Failed to create temp directory"})
			return
		}
		defer os.RemoveAll(tmpDir)

		// Write source file
		srcPath := filepath.Join(tmpDir, "main.gala")
		if err := os.WriteFile(srcPath, []byte(req.Code), 0644); err != nil {
			writeJSON(w, RunResponse{Error: "Failed to write source"})
			return
		}

		// Write gala.mod
		galaMod := "module playground\n\ngala 0.10.0\n"
		if err := os.WriteFile(filepath.Join(tmpDir, "gala.mod"), []byte(galaMod), 0644); err != nil {
			writeJSON(w, RunResponse{Error: "Failed to write gala.mod"})
			return
		}

		// Run with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, galaBin, "run", tmpDir)
		cmd.Dir = tmpDir

		start := time.Now()
		output, err := cmd.CombinedOutput()
		elapsed := time.Since(start)

		resp := RunResponse{
			Time: fmt.Sprintf("%.3fs", elapsed.Seconds()),
		}

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				resp.Error = "Execution timed out (30s limit)"
			} else {
				// Include output even on error (compile errors)
				resp.Error = string(output)
			}
		} else {
			resp.Output = string(output)
		}

		writeJSON(w, resp)
	}
}

func handleExamples(w http.ResponseWriter, r *http.Request) {
	entries, err := fs.ReadDir(exampleFiles, "examples")
	if err != nil {
		http.Error(w, "Failed to read examples", http.StatusInternalServerError)
		return
	}

	var examples []Example
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".gala") {
			continue
		}
		data, err := fs.ReadFile(exampleFiles, "examples/"+entry.Name())
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".gala")
		name = strings.ReplaceAll(name, "_", " ")
		// Title case
		words := strings.Fields(name)
		for i, w := range words {
			if len(w) > 0 {
				words[i] = strings.ToUpper(w[:1]) + w[1:]
			}
		}
		examples = append(examples, Example{
			Name: strings.Join(words, " "),
			Code: string(data),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examples)
}

func handleVersion(galaBin string) http.HandlerFunc {
	// Cache the version at startup
	var version string
	cmd := exec.Command(galaBin, "version")
	if out, err := cmd.Output(); err == nil {
		version = strings.TrimSpace(string(out))
	} else {
		version = "unknown"
	}

	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"version": version})
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

func main() {
	galaBin := findGala()
	if galaBin == "" {
		log.Fatal("GALA binary not found. Install with: gala-install")
	}
	log.Printf("Using GALA binary: %s", galaBin)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/run", handleRun(galaBin))
	mux.HandleFunc("/api/examples", handleExamples)
	mux.HandleFunc("/api/version", handleVersion(galaBin))

	// Static files
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// Determine bind address: 0.0.0.0 for Docker, 127.0.0.1 for local
	bind := "127.0.0.1"
	if os.Getenv("BIND_ALL") != "" {
		bind = "0.0.0.0"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	listenAddr := bind + ":" + port
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		// Fallback to random port on localhost
		listener, err = net.Listen("tcp", bind+":0")
		if err != nil {
			log.Fatalf("Failed to find available port: %v", err)
		}
	}

	addr := listener.Addr().String()
	url := "http://" + addr

	server := &http.Server{Handler: mux}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	fmt.Println()
	fmt.Println("  ╔═══════════════════════════════════════╗")
	fmt.Println("  ║         GALA Playground Server        ║")
	fmt.Printf("  ║  Running at: %-24s║\n", url)
	fmt.Println("  ╚═══════════════════════════════════════╝")
	fmt.Println()

	if bind == "127.0.0.1" {
		openBrowser(url)
	}

	if err := server.Serve(listener); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
