package gorule

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// script is the container and keeper of script location and data
type Script struct {
	input  []byte
	offset int
	line   int
}

// NewParser creates a new parser for script
func NewParser(script []byte) *Script {
	return &Script{
		input:  script,
		offset: 0,
	}
}

// EOF returns EOF when the script is finished
func (s *Script) EOF() bool {
	return s.offset >= len(s.input)
}

// Get gets the next byte and increses the offset
func (s *Script) Get() byte {
	out := s.input[s.offset]
	s.offset++
	return out
}

// Word gets the next word and increses the offset
func (s *Script) Word() (string, error) {
	word := []byte{}
	add := 0
	quoted := false

	// continue till we reach EOF
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

// eval evaluates 2 parameters in the script
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
			log.Printf("eval nr %d == %d", n1, n2)
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
		log.Printf("eval str %s == %s", p1, p2)
		return p1 == p2, nil
	case "!=":
		return p1 != p2, nil
	default:
		return false, fmt.Errorf("unknown string validator: %s", v)
	}
}

// Line returns the line number where the script is at
func (s *Script) Line() int {
	return s.line
}

// Parse parses the script, and changes the interfaces defined as input based on that
func parse(i map[string]interface{}, script []byte) error {
	log.Printf("input interface: %+v", i)
	log.Printf("input script: %s", script)

	script = stripComments(script)
	/*script, err := parseVariables(i, script)
	if err != nil {
		return err
	}*/
	log.Printf("cleaned script: %s", script)

	parser := NewParser(script)

	executeFuncCount := 0
	executeFuncMap := map[int]bool{0: true}
	run := true

	// Continue till we reach the EOF
	for !parser.EOF() {
		word, err := parser.Word()
		if err != nil {
			return fmt.Errorf("could not parse script at line:%d error:%s", parser.Line(), err)
		}
		//log.Printf("word: %s", word)
		switch word {
		// if and else means we validate the 3 words after that
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

			if run == false {
				continue
			}

			param1, err = parseVariableStrings(i, param1)
			if err != nil {
				return fmt.Errorf("error parsing value as 1st parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			param2, err = parseVariableStrings(i, param2)
			if err != nil {
				return fmt.Errorf("error parsing value as 1st parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			result, err := parser.eval(param1, validator, param2)
			if err != nil {
				return fmt.Errorf("failed to validate '%s' at line:%d error:%s", word, parser.Line(), err)
			}
			log.Printf("result: %t", result)

			executeFuncMap[executeFuncCount+1] = result

		// log prints the next word (or string) to the output
		case "log":
			param1, err := parser.Word()
			if err != nil {
				return fmt.Errorf("expected string as 1st parameter to 'log' at line:%d error:%s", parser.Line(), err)
			}
			if run == false {
				continue
			}
			log.Printf("Log entry: %s", param1)

		// else means we can do the oposite of the previous block
		case "else":
			// take previous evaluation, and negate that
			executeFuncMap[executeFuncCount+1] = !executeFuncMap[executeFuncCount+1]

		// open curley brackets defines the start of a block, which we may or may not execute
		case "{":
			executeFuncCount++
			run = executeFuncMap[executeFuncCount]

		// closing curley brackets defines the start of a block, which we may or may not execute
		case "}":
			executeFuncCount--
			run = executeFuncMap[executeFuncCount]

		case "var":

			variable, err := parser.Word()
			if err != nil {
				return fmt.Errorf("expected resouce variable as 1st parameter after '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			value, err := parser.Word()
			if err != nil {
				return fmt.Errorf("expected set variable as 2st parameter after '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			if run == false {
				continue
			}

			if _, ok := i[variable]; ok {
				return fmt.Errorf("variable resource with the name '%s' already exists at line:%d", variable, parser.Line())
			}

			log.Printf("setting interface variable: %s to %s", variable, value)

			i[variable] = value

			// only execute if we find this inside a runnable block

		// any other text needs to be seperately interpeted
		default:
			// test if item is a editable variable

			// ignore empty words (this generally happens at the EOF or dus to odd spaces between groups)
			if word == "" {
				continue
			}

			log.Printf("word: %s", word)
			if _, ok := i[word]; ok {

				log.Printf("OK word: %s", word)
			}
			log.Printf("word: %s", word)
			// if a word contains a dot, we asume its a parameter, and we want to set or remove it based on the next parameter
			if _, ok := i[word]; ok || strings.Contains(word, ".") {
				param1 := word
				validator, err := parser.Word()
				if err != nil {
					return fmt.Errorf("expected validator as 1st parameter after variable '%s' at line:%d error:%s", word, parser.Line(), err)
				}
				param2, err := parser.Word()
				if err != nil {
					return fmt.Errorf("expected variable as 2nd parameter after variable '%s' at line:%d error:%s", word, parser.Line(), err)
				}

				// only execute if we find this inside a runnable block
				if run == false {
					continue
				}

				// split and check if it IS a resource
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
					// maybe it was not a resource at all???
					return fmt.Errorf("unknown resource '%s' at line:%d", param1, parser.Line())
				}
			} else {
				// something did not make sense :-(
				return fmt.Errorf("unexpected item in script logic. '%s' does not make sense at line:%d error:%s", word, parser.Line(), err)
			}
		}
	}
	return nil
}

// stripComments removes all comments of type:
// // comments
// # comments
// /* com
//     ments */
func stripComments(script []byte) []byte {
	re := regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/|(?s)#.*?\n")
	return re.ReplaceAll(script, nil)
}

// parseVariableStrings does a regex replace of all $(parameters)
// and queries the translateVariable function to find the value related to the parameter
func parseVariableStrings(i map[string]interface{}, script string) (string, error) {
	var err error
	r := regexp.MustCompile(`\$\(([a-zA-Z.-_]+)\)`)
	script = r.ReplaceAllStringFunc(script, func(m string) string {
		variable := m[2 : len(m)-1]
		var result string
		result, err = translateVariable(i, variable)
		if err != nil {
			return variable
		}
		return result
	})
	return script, err
}

// parseVariables does a regex replace of all $(parameters) to their "value" in Byte
// you can use this to transform the entire script, but that defeats the lexer logic
/*
func parseVariables(i map[string]interface{}, script []byte) ([]byte, error) {
	var err error
	r := regexp.MustCompile(`\$\(([a-zA-Z.-_]+)\)`)
	script = r.ReplaceAllFunc(script, func(m []byte) []byte {
		variable := m[2 : len(m)-1]
		var result string
		result, err = translateVariable(i, string(variable))
		if err != nil {
			return variable
		}
		// we are a string that consists of multiple words
		if strings.Contains(result, " ") {
			b := append([]byte("\""), []byte(result)...)
			b = append(b, '"')
			return b
		}
		return []byte(result)
	})
	return script, err
}
*/

// translateVariable translates a string to the variable in the interfaces
func translateVariable(i map[string]interface{}, variable string) (string, error) {
	resource := strings.Split(variable, ".")
	if r, ok := i[resource[0]]; ok {
		result, err := getInterface(r, resource[1:])
		if err != nil {
			return "", fmt.Errorf("error translating variable '%s of resource '%s': %s", variable, resource[0], err)
		}
		switch result.(type) {
		case string:
			return result.(string), nil
		case int:
			return strconv.Itoa(result.(int)), nil
		case int64:
			return strconv.Itoa(int(result.(int64))), nil
		}
	}
	return "", fmt.Errorf("Unknown resource '%s' used in variable: %s", resource[0], variable)
}

/*
func translateVariable(i map[string]interface{}, variable string) (string, error) {
	charVariable := variable[:][0]
	log.Printf("v: %s 1", variable)
	if (charVariable > 'a' || charVariable < 'z') && (charVariable > 'A' || charVariable < 'Z') {
		log.Printf("v: %s 2", variable)
		resource := strings.Split(variable, ".")
		if len(resource) == 1 {
			log.Printf("v: %s 3", variable)
			return variable, nil
		}
		if r, ok := i[resource[0]]; ok {
			log.Printf("v: %s 4", variable)
			result, err := getInterface(r, resource[1:])
			if err != nil {
				return "", fmt.Errorf("error translating variable '%s of resource '%s': %s", variable, resource[0], err)
			}
			log.Printf("x temp %T", result)
			log.Printf("x temp %v", reflect.ValueOf(result))
			//log.Printf("x temp %v", result.(int))
			switch result.(type) {
			case string:
				return result.(string), nil
			case int:
				return strconv.Itoa(result.(int)), nil
			case int64:
				return strconv.Itoa(int(result.(int64))), nil
			}
		} else {
			return "", fmt.Errorf("Unknown resource '%s' used in variable: %s", resource[0], variable)
		}
	}
	return variable, nil
}
*/
