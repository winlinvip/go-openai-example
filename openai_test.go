package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
	"strings"
	"testing"
)

var conf openai.ClientConfig

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	conf = openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	conf.BaseURL = os.Getenv("OPENAI_PROXY")
}

func Test_OpenAI_Chat(t *testing.T) {
	client := openai.NewClientWithConfig(conf)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "Hello!"},
			},
		},
	)
	if err != nil {
		t.Errorf("Completion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}

func Test_OpenAI_ASR(t *testing.T) {
	client := openai.NewClientWithConfig(conf)
	resp, err := client.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:    openai.Whisper1,
			FilePath: "25-audio-f669a768-4099-4a73-910b-ab5d57e03545.m4a",
			Format:   openai.AudioResponseFormatVerboseJSON,
		},
	)
	if err != nil {
		t.Errorf("Transcription error: %v\n", err)
		return
	}

	fmt.Println(resp.Text)
	fmt.Println(fmt.Sprintf("Task: %v, Language: %v, Duration: %v, Segments: %v",
		resp.Task, resp.Language, resp.Duration, len(resp.Segments)))
	for _, s := range resp.Segments {
		fmt.Println(fmt.Sprintf("  #%v: [%.2f, %.2f] \"%v\" seek=%v, tokens=%v, temp=%v, avg=%v, comp=%v, nos=%v, trans=%v",
			s.ID, s.Start, s.End, s.Text, s.Seek, len(s.Tokens), s.Temperature, s.AvgLogprob, s.CompressionRatio, s.NoSpeechProb, s.Transient))
	}
}

func Test_OpenAI_TTS(t *testing.T) {
	client := openai.NewClientWithConfig(conf)
	resp, err := client.CreateSpeech(
		context.Background(),
		openai.CreateSpeechRequest{
			Model: openai.TTSModel1, Input: "Hello, AI world!",
			Voice: openai.VoiceNova, ResponseFormat: openai.SpeechResponseFormatAac,
		},
	)
	if err != nil {
		t.Errorf("TTS error: %v\n", err)
		return
	}
	defer resp.Close()

	out, err := os.Create("test.aac")
	if err != nil {
		t.Errorf("Unable to create the file for writing")
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, resp); err != nil {
		t.Errorf("Unable to write data to the file")
		return
	}

	fmt.Println("Done.")
}

// https://platform.openai.com/docs/guides/vision/what-type-of-files-can-i-upload
// We currently support PNG (.png), JPEG (.jpeg and .jpg), WEBP (.webp), and non-animated GIF (.gif).
func Test_OpenAI_Vision_JPEG(t *testing.T) {
	data, err := os.ReadFile("srs-image.jpg")
	if err != nil {
		t.Errorf("Unable to read the file")
		return
	}

	bd := base64.StdEncoding.EncodeToString(data)

	client := openai.NewClientWithConfig(conf)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "Recognize the text in the image."},
				{Role: openai.ChatMessageRoleUser, MultiContent: []openai.ChatMessagePart{
					{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{
						Detail: openai.ImageURLDetailLow, URL: fmt.Sprintf("data:image/jpeg;base64,%v", bd),
					}},
				}},
			},
		},
	)
	if err != nil {
		t.Errorf("Completion error: %v\n", err)
		return
	}

	content := resp.Choices[0].Message.Content
	if !strings.Contains(content, "simple, high-efficiency, real-time video server") {
		t.Errorf("Expected text not found: %v", content)
		return
	}

	fmt.Println(content)
}

func Test_OpenAI_Vision_PNG(t *testing.T) {
	data, err := os.ReadFile("srs-image.png")
	if err != nil {
		t.Errorf("Unable to read the file")
		return
	}

	bd := base64.StdEncoding.EncodeToString(data)

	client := openai.NewClientWithConfig(conf)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "Convert the image to text."},
				{Role: openai.ChatMessageRoleUser, MultiContent: []openai.ChatMessagePart{
					{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{
						Detail: openai.ImageURLDetailLow, URL: fmt.Sprintf("data:image/png;base64,%v", bd),
					}},
				}},
			},
		},
	)
	if err != nil {
		t.Errorf("Completion error: %v\n", err)
		return
	}

	content := resp.Choices[0].Message.Content
	if !strings.Contains(content, "simple, high-efficiency, real-time video server") {
		t.Errorf("Expected text not found: %v", content)
		return
	}

	fmt.Println(content)
}
