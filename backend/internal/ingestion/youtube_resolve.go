package ingestion

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ResolveYouTubeURL 把各種 YouTube 網址轉成 Atom feed URL。
// 非 YouTube 網址原樣回傳。
//
// 支援格式：
//   - https://www.youtube.com/@handle
//   - https://www.youtube.com/channel/UCxxxxxxxx
//   - https://www.youtube.com/c/channelname
//   - https://www.youtube.com/user/username
func ResolveYouTubeURL(rawURL string) (string, error) {
	if !strings.Contains(rawURL, "youtube.com") {
		return rawURL, nil
	}

	channelID, err := resolveChannelID(rawURL)
	if err != nil {
		return "", err
	}
	return "https://www.youtube.com/feeds/videos.xml?channel_id=" + channelID, nil
}

// resolveChannelID 從 YouTube 網址取得 channelId。
// 對於 /channel/UCxxx 直接解析；其他格式用 yt-dlp 抓第一支影片的 metadata。
func resolveChannelID(pageURL string) (string, error) {
	// /channel/UCxxx 直接從 URL 解析
	if idx := strings.Index(pageURL, "/channel/UC"); idx >= 0 {
		rest := pageURL[idx+len("/channel/"):]
		// 截到下一個 / 或結尾
		if end := strings.IndexByte(rest, '/'); end >= 0 {
			rest = rest[:end]
		}
		if strings.HasPrefix(rest, "UC") && len(rest) > 10 {
			return rest, nil
		}
	}

	// 其他格式（@handle、/c/、/user/）：用 yt-dlp 抓第一支影片 JSON metadata
	channelID, err := channelIDViaYtDlp(pageURL)
	if err != nil {
		return "", fmt.Errorf("無法解析 YouTube 頻道：%w\n提示：請確認網址為頻道頁面（如 youtube.com/@channelname）", err)
	}
	return channelID, nil
}

func channelIDViaYtDlp(pageURL string) (string, error) {
	// yt-dlp --flat-playlist -I 1 -j 只抓第一筆 entry 的 JSON，不實際下載
	done := make(chan struct{})
	var out []byte
	var cmdErr error

	go func() {
		defer close(done)
		cmd := exec.Command("yt-dlp",
			"--playlist-items", "1",
			"-j",
			"--quiet",
			"--no-warnings",
			pageURL,
		)
		cmd.Env = append(os.Environ(), "HOME=/tmp")
		out, cmdErr = cmd.Output()
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		return "", fmt.Errorf("yt-dlp timeout")
	}

	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*exec.ExitError); ok {
			return "", fmt.Errorf("yt-dlp: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", cmdErr
	}

	var info struct {
		ChannelID string `json:"channel_id"`
	}
	// 只取第一行（channel 頁可能有多行 JSON）
	firstLine := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	if err := json.Unmarshal([]byte(firstLine), &info); err != nil {
		return "", fmt.Errorf("parse yt-dlp output: %w", err)
	}
	if info.ChannelID == "" {
		return "", fmt.Errorf("channel_id not found in yt-dlp output")
	}
	return info.ChannelID, nil
}
