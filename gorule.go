package gorule

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Script struct {
	input  []byte
	offset int
	line   int
}

func NewParser(script []byte) *Script {
	return &Script{
		input:  script,
		offset: 0,
	}
}

func (s *Script) EOF() bool {
	return s.offset >= len(s.input)
}

func (s *Script) Get() byte {
	out := s.input[s.offset]
	s.offset++
	return out
}

func (s *Script) Word() (string, error) {
	word := []byte{}
	add := 0
	quoted := false
	for !s.EOF() {
		g := s.Get()
		add++
		// on enter increse line count, and return recovered words so far (maybe take in to account \ ending of lines to continue)
		if g == '\n' {
			s.line++
			return string(word), nil
		}
		// skip prefixed spaces
		if len(word) == 0 && (g == ' ' || g == '\t') {
			continue
		}

		// string parsing
		if g == '"' {
			// if word is not empty, its not start of string
			if len(word) > 0 {
				// if prev was \ then new char is quote
				if word[len(word)-1] == '\\' {
					word[len(word)-1] = g
					continue
				} else {
					quoted = false
					return string(word), nil
				}

			} else {
				quoted = true
				continue
			}
		}

		// if we hit space now, we have a word
		if len(word) > 0 && (g == ' ') && !quoted {
			return string(word), nil
		}
		// append letter
		word = append(word, g)
	}
	// if we reached EOF, but have word, return the word first without error
	if len(word) > 0 {
		return string(word), nil
	}
	return string(word), nil //, io.EOF
}

func (s *Script) eval(p1, v, p2 string) (bool, error) {
	// test if p1 is number
	if n1, err := strconv.Atoi(p1); err == nil {
		// n2 is a number, so p2 should be a number too
		n2, err := strconv.Atoi(p2)
		if err != nil {
			// p2 is not a number, return error
			return false, fmt.Errorf("cannot compare string with number")
		}

		switch v {
		case "==":
			return n1 == n2, nil
		case "<=":
			return n1 <= n2, nil
		case ">=":
			return n1 >= n2, nil
		case "!=":
			return n1 != n2, nil
		default:
			return false, fmt.Errorf("unknown number validator: %s", v)
		}
	}

	// so we are comparing strings
	switch v {
	case "==":
		return p1 == p2, nil
	case "!=":
		return p1 != p2, nil
	default:
		return false, fmt.Errorf("unknown string validator: %s", v)
	}
}

func (s *Script) Line() int {
	return s.line
}

func parse(i map[string]interface{}, script []byte) error {
	log.Printf("input interface: %+v", i)
	log.Printf("input script: %s", script)

	parser := NewParser(script)

	executeFuncCount := 0
	executeFuncMap := map[int]bool{0: true}
	ignoreline := false
	lastLine := 0
	run := true

	for !parser.EOF() {
		if lastLine != parser.Line() {
			lastLine = parser.Line()
			ignoreline = false
		}
		word, err := parser.Word()
		if err != nil {
			return fmt.Errorf("could not parse script at line:%d error:%s", parser.Line(), err)
		}
		//log.Printf("word: %s", word)
		switch word {
		case "if", "elseif":
			param1, err := parser.Word()
			if err != nil {
				return fmt.Errorf("expected value as 1st parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}
			validator, err := parser.Word()
			if err != nil {
				return fmt.Errorf("expected validator as 2nd parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}
			param2, err := parser.Word()
			if err != nil {
				return fmt.Errorf("expected value as 3st parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			if run == false || ignoreline == true {
				continue
			}
			result, err := parser.eval(param1, validator, param2)
			if err != nil {
				return fmt.Errorf("failed to validate '%s' at line:%d error:%s", word, parser.Line(), err)
			}
			log.Printf("result: %t", result)

			executeFuncMap[executeFuncCount+1] = result
		case "log":
			param1, err := parser.Word()
			if err != nil {
				return fmt.Errorf("expected string as 1st parameter to 'log' at line:%d error:%s", parser.Line(), err)
			}
			if run == false || ignoreline == true {
				continue
			}
			log.Printf("Log entry: %s", param1)
		case "else":
			if ignoreline == true {
				continue
			}
			// take previous evaluation, and negate that
			executeFuncMap[executeFuncCount+1] = !executeFuncMap[executeFuncCount+1]
		case "{":
			if ignoreline == true {
				continue
			}
			executeFuncCount++
			run = executeFuncMap[executeFuncCount]
		case "}":
			if ignoreline == true {
				continue
			}
			executeFuncCount--
			run = executeFuncMap[executeFuncCount]
		default:
			if ignoreline == true {
				continue
			}
			// test if item is a editable variable
			if word == "" {
				// silently ignore
				continue
			}

			if strings.HasPrefix(word, "#") {
				ignoreline = true
				continue
			}
			if strings.Contains(word, ".") {
				param1 := word
				validator, err := parser.Word()
				if err != nil {
					return fmt.Errorf("expected validator as 1st parameter after variable '%s' at line:%d error:%s", word, parser.Line(), err)
				}
				param2, err := parser.Word()
				if err != nil {
					return fmt.Errorf("expected variable as 2nd parameter after variable '%s' at line:%d error:%s", word, parser.Line(), err)
				}

				if run == false {
					continue
				}

				resource := strings.Split(param1, ".")
				if r, ok := i[resource[0]]; ok {

					switch validator {
					case "=":
						log.Printf("Set entry: %s to %s", param1, param2)
						err := modifyInterface(r, resource[1:], param2)
						if err != nil {
							return fmt.Errorf("error modifing '%s' to '%s' at line:%d error:%s", param1, param2, parser.Line(), err)
						}
					}
				} else {
					return fmt.Errorf("unknown resource '%s' at line:%d", param1, parser.Line())
				}
			} else {
				return fmt.Errorf("unexpected item in script logic. '%s' does not make sense at line:%d error:%s", word, parser.Line(), err)
			}
		}
	}
	log.Printf("EOF: %t", parser.EOF())
	return nil
}
