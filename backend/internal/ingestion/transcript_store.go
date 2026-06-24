package ingestion

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// WriteTranscript writes text to dataDir/transcripts/{podcastID}/{episodeID}.txt.gz
// and returns the relative path to store in the DB.
func WriteTranscript(dataDir string, podcastID, episodeID int, text string) (string, error) {
	relPath := fmt.Sprintf("transcripts/%d/%d.txt.gz", podcastID, episodeID)
	absPath := filepath.Join(dataDir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", fmt.Errorf("mkdir transcript dir: %w", err)
	}
	f, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("create transcript file: %w", err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	if _, err := gz.Write([]byte(text)); err != nil {
		return "", fmt.Errorf("write transcript: %w", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("close gzip: %w", err)
	}
	return relPath, nil
}

// ReadTranscript reads and decompresses a transcript written by WriteTranscript.
// Returns "" if relPath is empty.
func ReadTranscript(dataDir, relPath string) (string, error) {
	if relPath == "" {
		return "", nil
	}
	f, err := os.Open(filepath.Join(dataDir, relPath))
	if err != nil {
		return "", fmt.Errorf("open transcript: %w", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()
	b, err := io.ReadAll(gz)
	if err != nil {
		return "", fmt.Errorf("read transcript: %w", err)
	}
	return string(b), nil
}
