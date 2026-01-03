package logging

import (
	"encoding/json"
	"fmt"
	"maps"
	"time"
)

type StdoutLogger struct{}

func (l *StdoutLogger) log(level, msg string, fields map[string]any) {
	entry := map[string]any{
		"level": level,
		"msg":   msg,
		"time":  time.Now().UTC().Format(time.RFC3339),
	}

	maps.Copy(entry, fields)

	b, _ := json.Marshal(entry)
	fmt.Println(string(b))
}

func (l *StdoutLogger) Info(msg string, fields map[string]any) {
	l.log("INFO", msg, fields)
}

func (l *StdoutLogger) Error(msg string, fields map[string]any) {
	l.log("ERROR", msg, fields)
}
