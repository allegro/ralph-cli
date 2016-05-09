package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

const (
	loggingEndpoint    = ""
	logstashDateFormat = "2006-01-02T15:04:05.999Z07:00"
)

// Message represents information that will be send to logger.
type Message struct {
	Message    string `json:"message"`
	SourceHost string `json:"SourceHost"`
}

// LogstashWriter is an alias for io.Writer interface with added config field
// (for ralph-cli config).
type LogstashWriter struct {
	config Config // this should be moved somewhere else, I guess.
	writer io.Writer
}

// Write implements io.Writer interface for LogstashWriter.
func (lw LogstashWriter) Write(p []byte) (int, error) {
	json, err := lw.buildJSON(p)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 0, err
	}
	data := fmt.Sprintf("%s\n", json)
	return fmt.Fprint(lw.writer, data)
}

func (lw LogstashWriter) buildJSON(message []byte) ([]byte, error) {
	m, err := lw.createLogstashMessage(string(message))
	if err != nil {
		return nil, err
	}
	json, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func (lw LogstashWriter) createLogstashMessage(message string) (*Message, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	return &Message{
		Message:    message,
		SourceHost: hostname,
	}, nil
}

// NewLogstashWriter creates new LogstashWriter based on ralph-cli config.
func NewLogstashWriter(config Config) *LogstashWriter {
	writer, err := net.Dial("udp", loggingEndpoint)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return &LogstashWriter{
		config: config,
		writer: writer,
	}
}
