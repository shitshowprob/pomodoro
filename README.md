# Pomodoro

I am thinking to create bunch of pomodoro timers in different languages. Starting with Go.

## Go

### Installation

### Using `go install`

```bash
go install github.com/shitshowprob/pomodoro/golang/cmd/pomodoro@latest
```

This will install the binary as `pomodoro` in your `$GOPATH/bin` directory.

### Building from source

```bash
git clone https://github.com/shitshowprob/pomodoro.git
cd pomodoro/golang
go build -o pomodoro ./cmd/pomodoro
./pomodoro
```

### Usage

Run the program:

```bash
pomodoro
```

### Controls

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate menu |
| `Enter` / `e` | Select pomodoro setting |
| `p` / `Esc` | Pause/Resume timer |
| `q` / `Ctrl+C` | Quit |

### Available Timer Settings

- **25/5** - 25 minute work session with 5 minute break
- **50/10** - 50 minute work session with 10 minute break

## License

MIT
