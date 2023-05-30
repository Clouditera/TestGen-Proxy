package main

import "time"

const Timeout = 5 * time.Minute

const (
	STATUS_TIMEOUT = 500
	STATUS_SUCCESS = 200
	STATUS_ERR_API = 400
	STATUS_ERR_INT = 300
)

type ChatCompletionResponse struct {
	ID     string `json:"id"`
	Object string `json:"object"`
	// 其他字段省略
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		// 其他字段省略
	} `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Input struct {
	FuncSig string `json:"func_signature"`
	RepoUrl string `json:"repository_url"`
}

type OpenAIError struct {
	Code           string  `json:"code,omitempty"`
	Message        string  `json:"message"`
	Param          *string `json:"param,omitempty"`
	Type           string  `json:"type"`
	HTTPStatusCode int     `json:"-"`
}

type OpenAIErrorResp struct {
	Error *OpenAIError `json:"error,omitempty"`
}

type GeneratedTestcase struct {
	TestDesc      string           `yaml:"test_desc" json:"test_desc"`             // 测试用例生成原因
	TestcaseAIArr []TestCaseAIYAML `yaml:"testcase_ai_arr" json:"testcase_ai_arr"` // 测试用例多样化生成内容
}

type TestCaseAIYAML struct {
	Index   string `yaml:"index" json:"index"`
	Content string `yaml:"content" json:"content"`
}

const promptHeader = "Now you are a testing and vulnerability expert. Given a code repository:"

const generationDesc = `you should find then learn usage examples for this function, therefore you write out code snippets to use this function with symbolic function in the form of fuzz<its type> e.g. fuzzstring(0) or fuzzUint64(0) or fuzzInt32(0) or fuzzfile(0).
                    The argument of symbolic function is the index of new symbolic value, thus the arguments of same symbolic function should be incremental, e.g. fuzzstring(0), fuzzstring(1), and fuzzstring(2) for string type and char* type.
                    The basic types of C language are string, char, int, file, etc. In the output code snippets, if the values of them appears in the generative code snippets, then replaced by symbolic value fuzz<its type>.
                    The symbols of output code snippet should be complete, they may be an included header or a function declaration or an function implemetation, try to find out or inline the missing symbols from the code repository.
                    For symbolic functions, add the header \"#include<easyfuzz.h>\".
                    When generating code snippets, replacing symbolic function for variables diversifiably.`

const outputFormat = "Finally, output the generated code snippets of testcases as you are an API responding a yaml. The yaml format of testcase generation response is as follows:\n" +
	"```go\n" +
	"type GeneratedTestcase struct { \n" +
	"	TestDesc String `yaml:\"test_desc\"`  // 测试用例生成原因\n" +
	"	TestcaseAIArr []TestCaseAIYAML `yaml:\"testcase_ai_arr\"` // 测试用例多样化生成内容\n" +
	"}\n" +
	"\n" +
	"type TestCaseAIYAML struct {\n" +
	"	Index    string `yaml:\"index\"`    // 生成的测试用例的序号\n" +
	"	Content string `yaml:\"content\"` // 生成的测试用例代码\n" +
	"}\n" +
	"```\n"
