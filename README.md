# gpt
```bash
go install github.com/tucnak/gpt@latest
export OPENAI_API_KEY=sk-...
export OPENAI_LOG_DIR=$HOME/gpt # /<time>.gpt.txt on every run
echo 'Tell a joke.' | gpt -4
```

This is a simple command-line tool to use GPT-3.5/4 streaming chat completion API with a rudimentary [vim-gpt][1] integration. There is a shorthand mode in which optional arguments are set in a particular order:

```bash
# temperature
echo 'Tell a joke.' | gpt -4 0.5
# temperature, max_length
echo 'Tell a joke.' | gpt -4 0.5 512
# temperature, max_length, top_p, fpen, ppen
echo 'Tell a joke.' | gpt -4 0.5 200 1 0 0
```

Any of those in particular may be set normally:

```
$ gpt -help
Usage of gpt:
  -3	use gpt3.5-turbo
  -4	use gpt4
  -fp float
    	frequency penalty
  -keyring string
    	store the password in keyring
  -max int
    	max length
  -p float
    	top_p sampling (default 1)
  -pp float
    	presence penalty
  -t float
    	temperature (default 0.7)
  -vim
    	vim mode
```

### Chat interface

Normally, the stdin input is considered a single USER prompt but `gpt` will also attempt to parse a plaintext conversation format. If the file starts with the guidemark immediately, no System prompt is assumed; the conversation is arbitrary-length, and does not have to adhere to some order. In Vim mode, it will also append an extra prompt marker.

```
System prompt here.

	>>>>>>
User prompt here.

	<<<<<<
Assistant continuation continuation here.
```

### Environment

By default, it will use `OPENAI_API_KEY` environment variable but there's also GNOME Keyring (Keychain on macOS) support that it reverts to, in case the api key wasn't set.

This program will make a record of every API run as individual file in the directory specified by `OPENAI_LOG_DIR` environment variable using a plaintext chat-oriented format that is used by [vim-gpt][1] to easily differentiate between turns.

### License

MIT

[1]: https://github.com/tucnak/vim-gpt
