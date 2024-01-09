# go-openai-example

This is an example of using the [OpenAI API](https://platform.openai.com/docs/api-reference).

## Usage

First, create `.env` and setup the API key, from [here](https://platform.openai.com/api-keys) you can create and get the API key.

```bash
OPENAI_API_KEY=sk-xxxxxx
OPENAI_PROXY=https://api.openai.com/v1
```

Then, run test for OpenAI service:

```bash
go test openai_test.go -v
```

If want to test Tencent AI service, create a `.env.tencent` file:

```bash
TENCENT_SPEECH_APPID=xxxx
TENCENT_SECRET_ID=xxxx
TENCENT_SECRET_KEY=xxxx
```

Then, run test for Tencent AI service:

```bash
go test tencent_test.go -v
```
