package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	// 解析命令行参数
	templateFile := flag.String("template", "example.txt", "模板文件路径（必需）")
	outputFile := flag.String("output", "", "输出文件路径（默认输出到标准输出）")
	flag.Parse()

	if *templateFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// 读取模板文件
	templateContent, err := ioutil.ReadFile(*templateFile)
	if err != nil {
		panic(err)
	}

	// 读取所有环境变量
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}

	// 准备输出
	var writer *bufio.Writer
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		writer = bufio.NewWriter(file)
	} else {
		writer = bufio.NewWriter(os.Stdout)
	}

	// 简单的模板处理
	content := string(templateContent)

	// 替换环境变量
	for key, value := range env {
		content = strings.ReplaceAll(content, "${"+key+"}", value)
	}

	// 处理简单的if条件
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// 处理if语句
		if strings.HasPrefix(strings.TrimSpace(line), "{{if") {
			// 提取条件
			condition := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "{{if"))
			condition = strings.TrimSuffix(condition, "}}")
			condition = strings.TrimSpace(condition)

			// 检查条件
			parts := strings.Split(condition, "==")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

				if env[key] == value {
					// 条件为真，继续处理下一行
					continue
				} else {
					// 条件为假，跳过直到找到endif
					continue
				}
			}
		}

		// 处理endif
		if strings.TrimSpace(line) == "{{endif}}" {
			continue
		}

		// 输出普通行
		writer.WriteString(line + "\n")
	}

	// 确保所有内容都被写入
	writer.Flush()
}
