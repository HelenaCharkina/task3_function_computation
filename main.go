package main

import (
	"encoding/json"
	"github.com/apaxa-go/eval"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"strconv"
	"strings"
)

func main() {
	fileName := os.Args[1]

	var inputData map[string]interface{}

	inputFile, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer inputFile.Close()

	var data []byte
	data, err = ioutil.ReadAll(inputFile)

	err = json.Unmarshal(data, &inputData)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			log.Println("Input is not JSON")
			os.Exit(1)
		case *json.UnmarshalTypeError:
			log.Println("Incorrect input data format")
			os.Exit(2)
		default:
			log.Fatal("Необработанная ошибка парсинга json файла: ", err)
		}

	}

	operations := inputData["operations"].(map[string]interface{})

	var operationsArray []string
	for key, _ := range operations {
		operationsArray = append(operationsArray, key)
	}

	function := inputData["function"].(string)
	for _, op := range operationsArray {
		value := operations[op]
		switch value.(type) {
		case string:
			function = strings.ReplaceAll(function, op, operations[op].(string))
		case float64:
			function = strings.ReplaceAll(function, op, strconv.FormatFloat(operations[op].(float64), 'f', -1, 64))
		default:
			log.Fatal("Неизвестное значение в массиве операций")
		}

	}
	splitFunctions := getFunctions(function)

	var result []string
	for _, function := range splitFunctions {
		s := calculate(function)

		expr, err := eval.ParseString(s, "")
		if err != nil {
			log.Fatal("eval.ParseString error", err)
		}

		r, err := expr.EvalToInterface(nil)
		if err != nil {
			log.Fatal("eval.ParseString error", err)
		}

		var r_string string
		switch r.(type) {
		case int:
			r_string = strconv.Itoa(r.(int))
		case float64:
			r_string = strconv.FormatFloat(r.(float64), 'f', -1, 64)
		}

		result = append(result, r_string)
	}
	var res string
	res += "["
	res += strings.Join(result, ",")
	res += "]"

	err = ioutil.WriteFile("output.json", []byte(res), 0644)
	if err != nil {
		log.Fatal("ioutil.WriteFile error : ", err)
	}
}

func getFunctions(s string) []string {
	var functions []string
	lvl := 0
	i := 0
	s = strings.ReplaceAll(s, " ", "")
	for len(s) > 0 {
		if len(s) == i {
			functions = append(functions, s)
			s = ""
			return functions
		}
		if s[i] == ',' && lvl == 0 {
			functions = append(functions, s[:i])
			s = s[i+1:]
			i = 0
		} else if s[i] == '(' {
			lvl++
		} else if s[i] == ')' {
			lvl--
		} else if len(s) < i {
			functions = append(functions, s)
			s = ""
		}

		if lvl < 0 {
			log.Fatal("parameter incorrect")
		}

		i++
	}

	return functions
}

func calculate(s string) string {

	regex, err := regexp.Compile(`.\([0-9.,+*]+\)`)
	if err != nil {
		log.Fatal("regexp.Compile error", err)
	}
	str := strings.ReplaceAll(s, "exp(", "math.Exp(")

	for regex.MatchString(str) {
		matched := regex.FindAllString(str, -1)

		for _, item := range matched {
				s := item[0]
			if s == 'p' {
				tmp := strings.ReplaceAll(string(item), ",", "+")
				tmp = strings.ReplaceAll(tmp, "(", "[")
				tmp = strings.ReplaceAll(tmp, ")", "]")
				str = strings.ReplaceAll(str, string(item), tmp)
			} else {
				tmp := item[1:]
				tmpStr := strings.ReplaceAll(string(tmp), ",", string(s))
				tmpStr = strings.ReplaceAll(tmpStr, "(", "")
				tmpStr = strings.ReplaceAll(tmpStr, ")", "")
				str = strings.ReplaceAll(str, string(item), tmpStr)
			}
		}
	}
	str = strings.ReplaceAll(str, "[", "(")
	str = strings.ReplaceAll(str, "]", ")")

	return str
}
