package subtitles

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type whisperClient struct {
	apiKey string
	hc     *http.Client
}

func newWhisper(apiKey string) *whisperClient {
	return &whisperClient{apiKey: apiKey, hc: &http.Client{Timeout: 5 * time.Minute}}
}

type whisperVerbose struct {
	Words []struct {
		Word  string  `json:"word"`
		Start float64 `json:"start"`
		End   float64 `json:"end"`
	} `json:"words"`
	Segments []struct {
		Words []struct {
			Word  string  `json:"word"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		} `json:"words"`
	} `json:"segments"`
}

func (w *whisperClient) Transcribe(ctx context.Context, audioPath string) (*Output, error) {
	f, err := os.Open(audioPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("model", "whisper-1")
	_ = mw.WriteField("response_format", "verbose_json")
	_ = mw.WriteField("timestamp_granularities[]", "word")
	part, err := mw.CreateFormFile("file", filepath.Base(audioPath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/audio/transcriptions", &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	res, err := w.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("whisper http %d: %s", res.StatusCode, string(raw))
	}

	var parsed whisperVerbose
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	return wordsFromWhisper(parsed), nil
}

func wordsFromWhisper(v whisperVerbose) *Output {
	var words []Word
	for _, w := range v.Words {
		text := strings.TrimSpace(w.Word)
		if text == "" {
			continue
		}
		words = append(words, Word{
			Text:    text,
			StartMs: int(w.Start * 1000),
			EndMs:   int(w.End * 1000),
		})
	}
	if len(words) == 0 {
		for _, seg := range v.Segments {
			for _, w := range seg.Words {
				text := strings.TrimSpace(w.Word)
				if text == "" {
					continue
				}
				words = append(words, Word{
					Text:    text,
					StartMs: int(w.Start * 1000),
					EndMs:   int(w.End * 1000),
				})
			}
		}
	}
	return &Output{Style: "tiktok-bold", Words: words, Source: "whisper"}
}

func isMockAudio(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return strings.HasPrefix(string(data), "MOCK_MP3")
}
