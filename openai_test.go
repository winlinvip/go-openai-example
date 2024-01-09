package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
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
			Model: openai.GPT4TurboPreview,
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
			FilePath: "livestream-25-180152-490.mp4",
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
