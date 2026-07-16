package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/glebarez/go-sqlite"
	"github.com/spf13/cobra"
)

var db *sql.DB

var rootCmd = &cobra.Command{
	Use:   "snare",
	Short: "snare is a local webhook interceptor and replayer",
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "file:trap.db?_pragma=journal_mode(WAL)")
	if err != nil {
		fmt.Printf("Failed to open DB: %v\n", err)
		os.Exit(1)
	}

	query := `
	CREATE TABLE IF NOT EXISTS requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		method TEXT,
		path TEXT,
		headers_json TEXT,
		body_blob TEXT,
		captured_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(query); err != nil {
		fmt.Printf("Failed to migrate DB: %v\n", err)
		os.Exit(1)
	}
}

var port string
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Start trapping incoming requests",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, _ := io.ReadAll(r.Body)
			headersBytes, _ := json.Marshal(r.Header)

			_, err := db.Exec(
				"INSERT INTO requests (method, path, headers_json, body_blob) VALUES (?, ?, ?, ?)",
				r.Method, r.URL.Path+"?"+r.URL.RawQuery, string(headersBytes), string(bodyBytes),
			)

			if err != nil {
				fmt.Printf("Failed to log request: %v\n", err)
				w.WriteHeader(500)
				return
			}

			fmt.Printf("🪤 Trapped: [%s] %s (%d bytes)\n", r.Method, r.URL.Path, len(bodyBytes))
			w.WriteHeader(200)
			w.Write([]byte(`{"status": "captured"}`))
		})

		fmt.Printf("🎧 Listening for webhooks on http://localhost:%s\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Printf("Server failed: %v\n", err)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List recently trapped requests",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		rows, err := db.Query("SELECT id, method, path, length(body_blob), captured_at FROM requests ORDER BY id DESC LIMIT 15")
		if err != nil {
			fmt.Println("Error querying DB:", err)
			return
		}
		defer rows.Close()

		columns := []table.Column{
			{Title: "ID", Width: 5},
			{Title: "Method", Width: 8},
			{Title: "Path", Width: 30},
			{Title: "Size", Width: 10},
			{Title: "Time", Width: 20},
		}

		var tableRows []table.Row
		for rows.Next() {
			var id, size int
			var method, path, tsStr string
			rows.Scan(&id, &method, &path, &size, &tsStr)
			
			ts, _ := time.Parse(time.RFC3339, tsStr)

			tableRows = append(tableRows, table.Row{
				fmt.Sprintf("%d", id), method, path, fmt.Sprintf("%dB", size), ts.Format(time.RFC822),
			})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(tableRows),
			table.WithHeight(len(tableRows)+1),
		)
		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Bold(true)
		t.SetStyles(s)

		fmt.Println("\nRecent Traps:\n" + t.View())
	},
}

var replayId int
var replayTarget string
var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay a trapped request against a target URL",
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		var method, path, headersJson, bodyBlob string
		err := db.QueryRow("SELECT method, path, headers_json, body_blob FROM requests WHERE id = ?", replayId).
			Scan(&method, &path, &headersJson, &bodyBlob)

		if err != nil {
			fmt.Printf("Could not find request ID %d: %v\n", replayId, err)
			return
		}

		fullUrl := replayTarget + path
		fmt.Printf("🔄 Replaying request to: %s\n", fullUrl)

		req, err := http.NewRequest(method, fullUrl, bytes.NewBuffer([]byte(bodyBlob)))
		if err != nil {
			fmt.Printf("Failed to build request: %v\n", err)
			return
		}

		var headers map[string][]string
		json.Unmarshal([]byte(headersJson), &headers)
		for k, v := range headers {
			for _, val := range v {
				req.Header.Add(k, val)
			}
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Replay failed: %v\n", err)
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("✅ Target responded with Status: %d\n", resp.StatusCode)
		fmt.Printf("Body: %s\n", string(respBody))
	},
}

func init() {
	listenCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
	replayCmd.Flags().IntVarP(&replayId, "id", "i", 0, "ID of the request to replay")
	replayCmd.Flags().StringVarP(&replayTarget, "target", "t", "http://localhost:3000", "Target base URL")
	replayCmd.MarkFlagRequired("id")

	rootCmd.AddCommand(listenCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(replayCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
