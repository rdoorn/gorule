package gorule

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// Script is the container and keeper of script location and data
type script struct {
	input  []byte
	offset int
	line   int
}

// new creates a new parser for script
func new(s []byte) *script {
	return &script{
		input:  s,
		offset: 0,
	}
}

// eof returns eof when the script is finished
func (s *script) eof() bool {
	return s.offset >= len(s.input)
}

// get gets the next byte and increses the offset
func (s *script) get() byte {
	out := s.input[s.offset]
	s.offset++
	return out
}

// word gets the next word and increses the offset
func (s *script) word() (string, error) {
	word := []byte{}
	add := 0
	quoted := false

	// continue till we reach eof
	for !s.eof() {
		g := s.get()
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

	// if we reached eof, but have word, return the word first without error
	if len(word) > 0 {
		return string(word), nil
	}
	return string(word), nil //, io.eof
}

// eval evaluates 2 parameters in the script
func (s *script) eval(p1, v, p2 string) (bool, error) {
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
		//log.Printf("eval str %s == %s", p1, p2)
		return p1 == p2, nil
	case "!=":
		return p1 != p2, nil
	case "regex":
		//log.Printf("doing regex check of %s on %s", p2, p1)
		re, err := regexp.Compile(p2)
		if err != nil {
			return false, err
		}
		return re.MatchString(p1), nil
	default:
		return false, fmt.Errorf("unknown string validator: %s", v)
	}
}

// Line returns the line number where the script is at
func (s *script) Line() int {
	return s.line
}

// Parse parses the script, and changes the interfaces defined as input based on that
func Parse(i map[string]interface{}, script []byte) error {
	//log.Printf("input interface: %+v", i)
	//log.Printf("input script: %s", script)

	script = stripComments(script)
	/*script, err := parseVariables(i, script)
	if err != nil {
		return err
	}*/
	//log.Printf("cleaned script: %s", script)

	parser := new(script)

	executeFuncCount := 0
	executeFuncMap := map[int]bool{0: true}
	ifTracker := map[int]bool{0: false} // true if we had a match in the if statements at this level
	run := true

	// Continue till we reach the eof
	for !parser.eof() {
		word, err := parser.word()
		if err != nil {
			return fmt.Errorf("could not parse script at line:%d error:%s", parser.Line(), err)
		}
		//log.Printf("word: %s", word)
		switch word {
		// if and else means we validate the 3 words after that
		case "if", "elseif":
			param1, err := parser.word()
			if err != nil {
				return fmt.Errorf("expected value as 1st parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}
			validator, err := parser.word()
			if err != nil {
				return fmt.Errorf("expected validator as 2nd parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}
			param2, err := parser.word()
			if err != nil {
				return fmt.Errorf("expected value as 3st parameter to '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			if run == false {
				continue
			}

			if word == "if" { // reset tracker on IF
				ifTracker[executeFuncCount+1] = false
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
			//log.Printf("result: %t", result)

			executeFuncMap[executeFuncCount+1] = result
			if result == true {
				ifTracker[executeFuncCount+1] = true
			}

		// log prints the next word (or string) to the output
		case "log":
			param1, err := parser.word()
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
			executeFuncMap[executeFuncCount+1] = !ifTracker[executeFuncCount+1]

		// open curley brackets defines the start of a block, which we may or may not execute
		case "{":
			executeFuncCount++
			run = executeFuncMap[executeFuncCount]

		// closing curley brackets defines the start of a block, which we may or may not execute
		case "}":
			executeFuncCount--
			run = executeFuncMap[executeFuncCount]

		case "var":

			variable, err := parser.word()
			if err != nil {
				return fmt.Errorf("expected resouce variable as 1st parameter after '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			value, err := parser.word()
			if err != nil {
				return fmt.Errorf("expected set variable as 2st parameter after '%s' at line:%d error:%s", word, parser.Line(), err)
			}

			if run == false {
				continue
			}

			if _, ok := i[variable]; ok {
				return fmt.Errorf("variable resource with the name '%s' already exists at line:%d", variable, parser.Line())
			}

			//log.Printf("setting interface variable: %s to %s", variable, value)

			i[variable] = value

			// only execute if we find this inside a runnable block

		// any other text needs to be seperately interpeted
		default:
			// test if item is a editable variable

			// ignore empty words (this generally happens at the eof or dus to odd spaces between groups)
			if word == "" {
				continue
			}

			//log.Printf("word: %s", word)
			// if a word contains a dot, we asume its a parameter, and we want to set or remove it based on the next parameter
			if _, ok := i[word]; ok || strings.Contains(word, ".") {
				param1 := word
				validator, err := parser.word()
				if err != nil {
					return fmt.Errorf("expected validator as 1st parameter after variable '%s' at line:%d error:%s", word, parser.Line(), err)
				}
				param2, err := parser.word()
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
						//log.Printf("Set entry: %s to %s", param1, param2)
						if len(resource) == 1 {
							i[resource[0]] = param2
							//log.Printf("Set ---- %s to %s ", resource[0], i[resource[0]])
							continue
						}
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
