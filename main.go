package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	openaiApiType := viper.GetString("openai-api-type")
	openaiApiKey := viper.GetString("openai-api-key")
	openaiApiUrl := viper.GetString("openai-api-url")
	openaiModel := viper.GetString("openai-model")
	accounts := viper.GetStringMapString("accounts")
	switch strings.ToLower(openaiApiType) {
	case "openai":
	case "azure":
	default:
		panic(fmt.Errorf("fatal error config file: unknow api type: %s", openaiApiType))
	}

	// 初始化 gin 引擎
	r := gin.Default()

	// 添加验证中间件
	authenticated := r.Group("/", gin.BasicAuth(accounts))

	// 添加反向代理路由
	authenticated.POST("/", func(c *gin.Context) {
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		defer cancel()

		reqBody, err := io.ReadAll(c.Request.Body)
		log.Println(reqBody)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code": STATUS_ERR_INT,
				"msg":  "Can't get request body: " + err.Error(),
				"data": nil,
			})
		}
		if raw != "" {
			path = path + "?" + raw
		}

		var input Input
		err = json.Unmarshal(reqBody, &input)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code": STATUS_ERR_INT,
				"msg":  "Can't unmarshal request body: " + err.Error(),
				"data": nil,
			})
		}

		// 		var funcs string
		// 		for i := 0; i < len(input.CrashFuncAIArr); i++ {
		// 			funcs += "\n```\n"
		// 			funcs += input.CrashFuncAIArr[i].Content
		// 			funcs += "\n```\n"
		// 		}
		//
		// 		prompts := promptHeader + input.Stacktrace + funcs

		prompts := promptHeader + input.RepoUrl + "To test the target function " + input.FuncSig + generationDesc + outputFormat
		// log.Println("prompts:", prompts)
		requestBody, err := json.Marshal(Request{
			Model: openaiModel,
			Messages: []Message{
				{Role: "user", Content: prompts},
			},
		})
		// log.Println("request:", requestBody)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code": STATUS_ERR_INT,
				"msg":  "Can't marshal request body to OpenAI: " + err.Error(),
				"data": nil,
			})
		}

		gin.DefaultWriter.Write([]byte(fmt.Sprintf("%s%v %v:\n%s\n", c.Request.Header["Authorization"], c.Request.Method, path, requestBody)))

		// 创建HTTP客户端
		client := &http.Client{}

		// 创建POST请求
		req, err := http.NewRequest("POST", openaiApiUrl, bytes.NewBuffer(requestBody))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code": STATUS_ERR_INT,
				"msg":  "Can't create request to OpenAI: " + err.Error(),
				"data": nil,
			})
		}

		// 设置请求头
		switch strings.ToLower(openaiApiType) {
		case "openai":
			req.Header.Set("Authorization", "Bearer "+openaiApiKey)
		case "azure":
			req.Header.Set("api-key", openaiApiKey)
		}

		req.Header.Set("Content-Type", "application/json")

		// 发送请求并获取响应
		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code": STATUS_ERR_INT,
				"msg":  "Error on connecting OpenAI: " + err.Error(),
				"data": nil,
			})
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code": STATUS_ERR_INT,
				"msg":  "Can't get response body: " + err.Error(),
				"data": nil,
			})
		}

		select {
		case <-ctx.Done():
			// 超时了，返回错误信息
			msg := gin.H{
				"code": STATUS_TIMEOUT,
				"msg":  "Timeout",
				"data": nil,
			}
			c.AbortWithStatusJSON(http.StatusOK, msg)
			return
		default:
			// 返回响应给客户端
			data := make(map[string]interface{})
			err = json.Unmarshal(respBody, &data)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code": STATUS_ERR_INT,
					"msg":  "Can't unmarshal response body: " + err.Error(),
					"data": nil,
				})
			}
			_, ok := data["error"]
			if ok {
				var errResp OpenAIErrorResp
				json.Unmarshal(respBody, &errResp)
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code": STATUS_ERR_API,
					"msg":  "OpenAI API error: " + errResp.Error.Code,
					"data": nil,
				})
				return
			}

			var resp ChatCompletionResponse
			err := json.Unmarshal(respBody, &resp)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code": STATUS_ERR_INT,
					"msg":  "Can't unmarshal response body: " + err.Error(),
					"data": nil,
				})
			}

			content := resp.Choices[0].Message.Content
			var generatedTestcase GeneratedTestcase
			err = yaml.Unmarshal([]byte(content), &generatedTestcase)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code": STATUS_ERR_INT,
					"msg":  "Can't unmarshal response content: " + err.Error(),
					"data": nil,
				})
			}
			jsonData, err := json.Marshal(generatedTestcase)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"code": STATUS_ERR_INT,
					"msg":  "Can't marshal response content to json: " + err.Error(),
					"data": nil,
				})
			}

			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{
				"code": STATUS_SUCCESS,
				"msg":  "Success",
				"data": string(jsonData),
			})
		}
	})

	// 启动服务
	log.Fatal(r.Run(":8080"))
}
