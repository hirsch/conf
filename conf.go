// Package conf offers read access to conf files.
// Configuration file format:
// 	#comment
// 	;comment
// 	[section]
// 	value=key
package conf

import (
	"errors"
	"io"
	"os"
)

type Conf struct {
	filename string
	data     map[string]map[string]string
}

const (
	stateStart = iota
	stateMid
	stateComment
	stateSection
	stateKey
	stateValue
	stateError
	stateEOF
)

type lexer struct {
	file *os.File

	bufferSection string
	bufferKey     string
	bufferValue   string
	bufferError   string
	buffer        string

	data map[string]map[string]string
}

// Open opens and parses a conf file.
func Open(filename string) (*Conf, error) {
	conf := &Conf{filename: filename}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	state := stateStart
	lex := &lexer{file, "", "", "", "", "", make(map[string]map[string]string)}
	for {
		switch state {
		case stateStart:
			state = lex.doStart()
		case stateMid:
			state = lex.doMid()
		case stateComment:
			state = lex.doComment()
		case stateSection:
			state = lex.doSection()
		case stateKey:
			state = lex.doKey()
		case stateValue:
			state = lex.doValue()
		case stateError:
			return nil, lex.doError()
		case stateEOF:
			conf.data = lex.data
			return conf, nil
		}
	}

}

func (lex *lexer) doStart() int {
	switch lex.look() {
	case "":
		return stateEOF
	case " ", "	", "\n":
		lex.add()
		return stateStart
	case "[":
		lex.add()
		lex.flush()
		return stateSection
	case "#", ";":
		lex.add()
		return stateComment
	}
	lex.add()
	lex.bufferError = "key not in section: " + lex.buffer
	return stateError
}

func (lex *lexer) doMid() int {
	switch lex.look() {
	case "":
		return stateEOF
	case " ", "	", "\n":
		lex.add()
		return stateMid
	case "[":
		lex.add()
		lex.flush()
		return stateSection
	case "#", ";":
		lex.add()
		return stateComment
	}
	lex.flush()
	return stateKey
}

func (lex *lexer) doComment() int {
	switch lex.look() {
	case "":
		return stateEOF
	case "\n":
		lex.add()
		if lex.bufferSection == "" {
			return stateStart
		}
		return stateMid
	}

	lex.add()
	return stateComment
}

func (lex *lexer) doSection() int {
	switch lex.look() {
	case "\n", "":
		lex.add()
		lex.bufferError = "broken section name: " + lex.buffer
		return stateError
	case "]":
		lex.bufferSection = lex.flush()
		lex.data[lex.bufferSection] = make(map[string]string)
		lex.add()
		return stateMid
	}
	lex.add()
	return stateSection
}

func (lex *lexer) doKey() int {
	switch lex.look() {
	case "\n", "":
		lex.add()
		lex.bufferError = "broken key name: " + lex.buffer
		return stateError
	case "=":
		lex.bufferKey = lex.flush()
		lex.add()
		lex.flush()
		return stateValue
	}
	lex.add()
	return stateKey
}

func (lex *lexer) doValue() int {
	switch lex.look() {
	case "\n":
		lex.bufferValue = lex.flush()
		lex.add()
		lex.data[lex.bufferSection][lex.bufferKey] = lex.bufferValue
		return stateMid
	case "":
		lex.bufferValue = lex.flush()
		lex.add()
		lex.data[lex.bufferSection][lex.bufferKey] = lex.bufferValue
		return stateEOF
	}
	lex.add()
	return stateValue
}

func (lex *lexer) doError() error {
	return errors.New(lex.bufferError)
}

func (lex *lexer) get() string {
	chr := make([]byte, 1)
	_, err := io.ReadFull(lex.file, chr)
	if err != nil {
		return ""
	}
	if string(chr[0]) == "\r" {
		return lex.get()
	}
	return string(chr[0])
}

func (lex *lexer) add() string {
	chr := lex.get()
	lex.buffer += chr
	return chr
}

func (lex *lexer) look() string {
	chr := lex.get()
	lex.file.Seek(-1, 1)
	return chr
}

func (lex *lexer) flush() string {
	save := lex.buffer
	lex.buffer = ""
	return save
}

// Read returns the value to a given section and key.
// An error will be returned if a key or section does not exist.
func (conf *Conf) Read(section, key string) (string, error) {
	value, exists := conf.data[section][key]
	if !exists {
		return "", errors.New("key or section does not exist")
	}
	return value, nil
}

