package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/joho/godotenv"
	"github.com/ossrs/go-oryx-lib/errors"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/asr"
	"github.com/tencentcloud/tencentcloud-speech-sdk-go/common"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
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

// See https://cloud.tencent.com/document/product/1073/37935
func Test_Tencent_TTS(t *testing.T) {
	authGenerateSign := func(secretKey string, requestData map[string]interface{}) string {
		url := "tts.cloud.tencent.com/stream"
		signStr := "POST" + url + "?"

		keys := []string{}
		for k := range requestData {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			signStr = fmt.Sprintf("%s%s=%v&", signStr, k, requestData[k])
		}
		signStr = signStr[:len(signStr)-1]

		h := hmac.New(sha1.New, []byte(secretKey))
		h.Write([]byte(signStr))
		return base64.StdEncoding.EncodeToString(h.Sum(nil))
	}

	if err := func() error {
		appID, err := strconv.ParseInt(AppID, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parse appid %v", AppID)
		}

		requestData := map[string]interface{}{
			"Action":          "TextToStreamAudio",
			"AppId":           int(appID), // replace with your AppId
			"Codec":           "pcm",
			"Expired":         time.Now().Unix() + 3600,
			"ModelType":       0,
			"PrimaryLanguage": 1,
			"ProjectId":       0,
			"SampleRate":      16000,
			"SecretId":        SecretID, // replace with your SecretId
			"SessionId":       "12345678",
			"Speed":           0,
			"Text":            "Hello World!",
			"Timestamp":       time.Now().Unix(),
			"VoiceType":       1009,
			"Volume":          5,
		}

		url := "https://tts.cloud.tencent.com/stream"

		jsonData, err := json.Marshal(requestData)
		if err != nil {
			return errors.Wrapf(err, "marshal json")
		}

		req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
		if err != nil {
			return errors.Wrapf(err, "create request")
		}
		req.Header.Set("Content-Type", "application/json")

		signature := authGenerateSign(SecretKey, requestData) // replace with your SecretKey
		req.Header.Set("Authorization", signature)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return errors.Wrapf(err, "request tts")
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "read body")
		}
		if strings.Contains(string(body), "Error") {
			return errors.Errorf("tts error: %v", string(body))
		}

		out, err := os.Create("test.wav")
		if err != nil {
			return errors.Wrapf(err, "create file")
		}
		defer out.Close()

		// Convert to s16le depth.
		data := make([]int, len(body)/2)
		for i := 0; i < len(body); i += 2 {
			data[i/2] = int(binary.LittleEndian.Uint16(body[i : i+2]))
		}

		enc := wav.NewEncoder(out, 16000, 16, 1, 1)
		defer enc.Close()

		ib := &audio.IntBuffer{
			Data: data, SourceBitDepth: 16,
			Format: &audio.Format{NumChannels: 1, SampleRate: 16000},
		}
		if err := enc.Write(ib); err != nil {
			return errors.Wrapf(err, "copy body")
		}

		return nil
	}(); err != nil {
		t.Errorf("test error: %v", err)
	}
}
