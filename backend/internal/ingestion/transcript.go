package ingestion

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// fetchTranscript 依優先序取得最準確的逐字稿：
//  1. YouTube auto-captions（yt-dlp，最快）
//  2. podcast:transcript URL（直接下載）
//  3. Whisper 音頻轉文字（最準但最慢，需設定 WHISPER_MODEL）
//  4. 都失敗回傳 ""，caller 會 fallback 到 show notes
func fetchTranscript(ep episodeData, whisperModel string) (string, string) {
	// 1. YouTube auto-captions
	if ep.VideoID != "" {
		text, err := fetchYouTubeCaptions(ep.VideoID)
		if err == nil && len(text) > 100 {
			return text, "yt_captions"
		}
		// YouTube 沒字幕，若有 whisper 嘗試抓音頻轉寫
		if whisperModel != "" && ep.AudioURL == "" {
			audioPath, err := downloadYouTubeAudio(ep.VideoID)
			if err != nil {
				log.Printf("[transcript] yt audio download failed: %v", err)
			} else {
				defer os.Remove(audioPath)
				text, err := runWhisper(audioPath, whisperModel)
				if err != nil {
					log.Printf("[transcript] whisper failed: %v", err)
				} else if len(text) > 100 {
					return text, "whisper_yt"
				} else {
					log.Printf("[transcript] whisper returned too little text (%d chars)", len(text))
				}
			}
		}
	}

	// 2. podcast:transcript URL
	if ep.TranscriptURL != "" {
		text, err := fetchRemoteTranscript(ep.TranscriptURL)
		if err == nil && len(text) > 50 {
			return text, "remote"
		}
	}

	// 3. Whisper on podcast audio (RSS enclosure)
	if whisperModel != "" && ep.AudioURL != "" {
		audioPath, err := downloadAudio(ep.AudioURL)
		if err != nil {
			log.Printf("[transcript] rss audio download failed: %v", err)
		} else {
			defer os.Remove(audioPath)
			text, err := runWhisper(audioPath, whisperModel)
			if err != nil {
				log.Printf("[transcript] whisper failed: %v", err)
			} else if len(text) > 100 {
				return text, "whisper_rss"
			} else {
				log.Printf("[transcript] whisper returned too little text (%d chars)", len(text))
			}
		}
	}

	return "", ""
}

// ─── YouTube captions ────────────────────────────────────────────────────────

func fetchYouTubeCaptions(videoID string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "yt-cap-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	log.Printf("[transcript] fetching yt captions: %s", videoID)
	args := []string{
		"--write-auto-sub",
		"--sub-langs", "zh-Hant,zh-Hans,zh,en",
		"--skip-download",
		"--convert-subs", "srt",
		"--no-playlist",
		"--quiet", "--no-warnings",
		"-o", filepath.Join(tmpDir, "%(id)s.%(ext)s"),
		"https://www.youtube.com/watch?v=" + videoID,
	}
	cmd := exec.Command("yt-dlp", args...)
	cmd.Env = append(os.Environ(), "HOME=/tmp")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp captions %s: %w (%s)", videoID, err, strings.TrimSpace(string(out)))
	}

	for _, lang := range []string{"zh-Hant", "zh-Hans", "zh", "en"} {
		matches, _ := filepath.Glob(filepath.Join(tmpDir, "*."+lang+".srt"))
		if len(matches) > 0 {
			b, err := os.ReadFile(matches[0])
			if err == nil {
				return parseSRT(string(b)), nil
			}
		}
	}
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "*.srt"))
	if len(matches) > 0 {
		b, err := os.ReadFile(matches[0])
		if err == nil {
			return parseSRT(string(b)), nil
		}
	}
	return "", fmt.Errorf("no captions for %s", videoID)
}

// ─── Whisper ─────────────────────────────────────────────────────────────────

func downloadYouTubeAudio(videoID string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "yt-audio-")
	if err != nil {
		return "", err
	}
	log.Printf("[transcript] downloading yt audio for whisper: %s", videoID)
	outTmpl := filepath.Join(tmpDir, "audio.%(ext)s")
	args := []string{
		"--extract-audio", "--audio-format", "mp3", "--audio-quality", "5",
		"--no-playlist", "--quiet",
		"-o", outTmpl,
		"https://www.youtube.com/watch?v=" + videoID,
	}
	cmd := exec.Command("yt-dlp", args...)
	cmd.Env = append(os.Environ(), "HOME=/tmp")
	out, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("yt-dlp audio %s: %w (%s)", videoID, err, strings.TrimSpace(string(out)))
	}
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "audio.*"))
	if len(matches) == 0 {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("no audio file for %s", videoID)
	}
	return matches[0], nil
}

func downloadAudio(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	ext := ".mp3"
	ct := resp.Header.Get("Content-Type")
	if strings.Contains(ct, "ogg") {
		ext = ".ogg"
	} else if strings.Contains(ct, "mp4") || strings.Contains(ct, "m4a") {
		ext = ".m4a"
	}

	f, err := os.CreateTemp("", "podcast-*"+ext)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(f, io.LimitReader(resp.Body, 150<<20)); err != nil { // 150MB limit
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func runWhisper(audioPath, model string) (string, error) {
	log.Printf("[transcript] running whisper model=%s on %s", model, audioPath)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", "/app/scripts/whisper_transcribe.py", model, audioPath)
	cmd.Env = append(os.Environ(), "HOME=/tmp")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("whisper: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("whisper: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ─── Remote transcript ───────────────────────────────────────────────────────

func fetchRemoteTranscript(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}
	content := string(body)
	if strings.Contains(content, "-->") {
		return parseSRT(content), nil
	}
	if strings.HasPrefix(content, "WEBVTT") {
		return parseVTT(content), nil
	}
	return strings.TrimSpace(content), nil
}

// ─── SRT / VTT parsers ───────────────────────────────────────────────────────

func parseSRT(srt string) string {
	var lines []string
	for _, line := range strings.Split(srt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, err := strconv.Atoi(line); err == nil {
			continue
		}
		if strings.Contains(line, "-->") {
			continue
		}
		if strings.HasPrefix(line, "<") && strings.HasSuffix(line, ">") {
			continue
		}
		lines = append(lines, line)
	}
	deduped := make([]string, 0, len(lines))
	for i, l := range lines {
		if i == 0 || l != lines[i-1] {
			deduped = append(deduped, l)
		}
	}
	return strings.Join(deduped, " ")
}

func parseVTT(vtt string) string {
	var lines []string
	for _, line := range strings.Split(vtt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "WEBVTT" || strings.Contains(line, "-->") ||
			strings.HasPrefix(line, "NOTE") || strings.HasPrefix(line, "STYLE") {
			continue
		}
		lines = append(lines, line)
	}
	deduped := make([]string, 0, len(lines))
	for i, l := range lines {
		if i == 0 || l != lines[i-1] {
			deduped = append(deduped, l)
		}
	}
	return strings.Join(deduped, " ")
}
