package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/ossrs/go-oryx-lib/errors"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/asr"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/common"
	"io/ioutil"
	"os"
	"testing"
)

var AppID, SecretID, SecretKey string

func init() {
	if err := godotenv.Load(".env.tencent"); err != nil {
		panic(err)
	}

	AppID = os.Getenv("TENCENT_SPEECH_APPID")
	SecretID = os.Getenv("TENCENT_SECRET_ID")
	SecretKey = os.Getenv("TENCENT_SECRET_KEY")
}

// MySpeechRecognitionListener implementation of SpeechRecognitionListener
type MySpeechRecognitionListener struct {
}

func (listener *MySpeechRecognitionListener) OnRecognitionStart(response *asr.SpeechRecognitionResponse) {
}

func (listener *MySpeechRecognitionListener) OnSentenceBegin(response *asr.SpeechRecognitionResponse) {
}

func (listener *MySpeechRecognitionListener) OnRecognitionResultChange(response *asr.SpeechRecognitionResponse) {
}

func (listener *MySpeechRecognitionListener) OnSentenceEnd(response *asr.SpeechRecognitionResponse) {
	fmt.Printf("ASR result: %s\n", response.Result.VoiceTextStr)
}

func (listener *MySpeechRecognitionListener) OnRecognitionComplete(response *asr.SpeechRecognitionResponse) {
}

func (listener *MySpeechRecognitionListener) OnFail(response *asr.SpeechRecognitionResponse, err error) {
}

func Test_Tencent_ASR(t *testing.T) {
	EngineModelType, inputFile, inputFormat := "16k_zh", "./test.wav", asr.AudioFormatWav
	if err := func() error {
		recognizer := asr.NewSpeechRecognizer(
			AppID, common.NewCredential(SecretID, SecretKey),
			EngineModelType, &MySpeechRecognitionListener{},
		)
		recognizer.VoiceFormat = inputFormat
		if err := recognizer.Start(); err != nil {
			return errors.Wrapf(err, "recognizer start failed")
		}
		defer recognizer.Stop()

		audio, err := os.Open(inputFile)
		if err != nil {
			return errors.Wrapf(err, "open file error")
		}
		defer audio.Close()

		if b, err := ioutil.ReadFile(inputFile); err != nil {
			return errors.Wrapf(err, "read file error")
		} else {
			if err = recognizer.Write(b); err != nil {
				return errors.Wrapf(err, "recognizer write error")
			}
		}

		return nil
	}(); err != nil {
		t.Errorf("test error: %v", err)
	}
}
