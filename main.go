package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/zalando/go-keyring"
)

const nop = "nop"

var (
	bg     = context.Background()
	model  = ""
	apiKey = os.Getenv("OPENAI_API_KEY")
	dir    = os.Getenv("OPENAI_LOG_DIR")

	password    = flag.String("keyring", "", "store the password in keyring")
	gpt3        = flag.Bool("3", false, "use gpt3.5-turbo")
	gpt4        = flag.Bool("4", false, "use gpt4")
	vim         = flag.Bool("vim", false, "vim mode")
	maxLength   = flag.Int("max", 0, "max length")
	temperature = flag.Float64("t", 0.7, "temperature")
	top         = flag.Float64("p", 1.0, "top_p sampling")
	frequency   = flag.Float64("fp", 0, "frequency penalty")
	presence    = flag.Float64("pp", 0, "presence penalty")

	stderrf = func(format string, a ...interface{}) {
		fmt.Fprintf(os.Stderr, format, a...)
	}
	exit = func(code int) {
		os.Exit(code)
	}
)

const (
	system    = "system"
	user      = "user"
	assistant = "assistant"

	prompt = "\t>>>>>>"
	cont   = "\t<<<<<<"
)

func main() {
	flag.Parse()
	// positional variants of the fine-control flags
	for k, arg := range flag.Args() {
		var err error
		i, intErr := strconv.Atoi(arg)
		j, floatErr := strconv.ParseFloat(arg, 32)
		switch k {
		case 0:
			err = floatErr
			*temperature = j
		case 1:
			err = intErr
			*maxLength = i
		case 2:
			err = floatErr
			*top = j
		case 3:
			err = floatErr
			*frequency = j
		case 4:
			err = floatErr
			*presence = j
		}
		if err != nil {
			stderrf("gpt: pos=%d bad argument %v\n", k, arg)
			exit(1)
		}
	}
	switch {
	case *password != "":
		if err := keyring.Set("gpt", "key", *password); err != nil {
			stderrf("gpt: %v\n", err)
			exit(1)
		}
		fmt.Println(*password)
		return
	case *gpt3:
		model = openai.GPT3Dot5Turbo
	case *gpt4:
		model = openai.GPT4
	default:
		flag.Usage()
		exit(1)
	}
	// if not provided by environment, use the secret from keychain
	if apiKey == "" {
		secret, err := keyring.Get("gpt", "key")
		if err != nil {
			stderrf("gpt: api key: %v\n", err)
			flag.Usage()
			exit(1)
		}
		apiKey = secret
	}
	if dir == "" {
		dir = os.TempDir()
	}

	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		stderrf("gpt: stdin: %v\n", err)
		exit(1)
	}
	var req = request(b)

	client := openai.NewClient(apiKey)
	stream, err := client.CreateChatCompletionStream(bg, req)
	if err != nil {
		panic(err)
	}
	defer stream.Close()
	now := time.Now().Format("2006-01-02T15-04-05")
	f, err := os.Create(path.Join(dir, now+".gpt.txt"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	for _, msg := range req.Messages {
		switch msg.Role {
		case user:
			fmt.Fprintln(f, prompt)
		case assistant:
			fmt.Fprintln(f, cont)
		}
		fmt.Fprintln(f, msg.Content)
		fmt.Fprintln(f)
	}
	fmt.Fprintln(f, cont)
	if *vim {
		fmt.Printf("\n%s\n", cont)
	}
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println()
			if *vim {
				fmt.Printf("\n%s\n", prompt)
			}
			fmt.Fprintln(f)
			return
		}
		if err != nil {
			stderrf("%+v\n", err)
			exit(1)
		}
		chunk := resp.Choices[0].Delta.Content
		fmt.Printf(chunk)
		fmt.Fprintf(f, chunk)
	}
}

func request(b []byte) openai.ChatCompletionRequest {
	s := strings.TrimSpace(string(b))
	if s == "" {
		os.Exit(0)
	}
	return openai.ChatCompletionRequest{
		Model:            model,
		Messages:         parse(s),
		MaxTokens:        *maxLength,
		Temperature:      float32(*temperature),
		TopP:             float32(*top),
		FrequencyPenalty: float32(*frequency),
		PresencePenalty:  float32(*presence),
		Stream:           true,
	}
}

type message = openai.ChatCompletionMessage

// Split is the function that splits the input into system/user/assistant
// messages. The input is provided in the following format:
//
//     system message
//     \t>>>>>>>
//     user message
//     \t<<<<<<<
//     assistant message
func parse(s string) (log []message) {
	var (
		re   = regexp.MustCompile(`\n?[\t\s]+(>{3,}|<{3,})\n`)
		conv = re.Split(s, -1)
		mark = re.FindAllStringSubmatch(s, -1)
		push = func(role string, content string) {
			content = strings.TrimSpace(content)
			if content == "" {
				return
			}
			log = append(log, message{Role: role, Content: content})
		}
	)
	if len(conv) == 1 {
		push(user, conv[0])
		return
	}
	if conv[0] != "" {
		push(system, conv[0])
	}
	for i := range mark {
		switch msg := conv[i+1]; mark[i][1][0] {
		case '>':
			push(user, msg)
		case '<':
			push(assistant, msg)
		default:
			panic("bad parse")
		}
	}
	return
}
