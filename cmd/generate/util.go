package generate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/chries19921104/gin-admin-cli/v1/util"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	delimiter = "\n"
)

// 获取模块头
func getModuleHeader(moduleName string, imports ...string) *bytes.Buffer {
	buf := new(bytes.Buffer)

	buf.WriteString(fmt.Sprintf("package %s", moduleName))
	buf.WriteString(delimiter)
	buf.WriteString(delimiter)

	if len(imports) > 0 {
		buf.WriteString("import (")
		buf.WriteString(delimiter)

		for _, s := range imports {
			buf.WriteByte('\t')
			buf.WriteString(s)
			buf.WriteString(delimiter)
		}

		buf.WriteByte(')')
		buf.WriteString(delimiter)
		buf.WriteString(delimiter)
	}

	return buf
}

// 写入文件
func createFile(ctx context.Context, name string, buf *bytes.Buffer) error {
	exists := true
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			exists = false
		} else {
			return err
		}
	}

	if exists {
		return fmt.Errorf("file has been exists: %s", name)
	}

	os.MkdirAll(filepath.Dir(name), 0777)
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, _ = io.Copy(file, buf)
	return nil
}

// 执行go代码格式化
func execGoFmt(name string) error {
	cmd := exec.Command("gofmt", "-w", name, name)
	return cmd.Run()
}

// 执行解析模板
func execParseTpl(tpl string, data interface{}) (*bytes.Buffer, error) {
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"condition": func(field TplFieldItem) string {
			if field.ConditionArray {
				return "len(v) > 0"
			} else if field.StructFieldType == "string" {
				return `v != ""`
			} else if strings.Index(field.StructFieldType, "int") > -1 ||
				strings.Index(field.StructFieldType, "float") > -1 {
				return `v > 0`
			}
			return "v"
		},
		"fieldToPluralAndLowerUnderlinedName": func(v string) string {
			return util.ToLowerUnderlinedNamer(util.ToPlural(v))
		},
		"fieldToPlural": func(v string) string {
			return util.ToPlural(v)
		},
		"fieldToLowerUnderlinedName": func(v string) string {
			return util.ToLowerUnderlinedNamer(v)
		},
	}).Parse(tpl))

	buf := new(bytes.Buffer)
	err := t.Execute(buf, data)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// 读取文件内容
func readFile(name string) (*bytes.Buffer, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, file)
	return buf, nil
}

// 写入文件
func writeFile(name string, buf *bytes.Buffer) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, _ = io.Copy(file, buf)
	return nil
}

// insertContent 插入文件内容
// fn 回调当前行数据，返回-1为插入当前行之前，1为插入当前行之后
func insertContent(name string, fn func(string) (string, int, bool)) error {
	buf, err := readFile(name)
	if err != nil {
		return err
	}

	nbuf := new(bytes.Buffer)
	scanner := bufio.NewScanner(buf)

	for scanner.Scan() {
		cline := scanner.Text()
		data, flag, ok := fn(cline)
		if ok {
			if flag == -1 {
				nbuf.WriteString(data)
				nbuf.WriteString(delimiter)
				nbuf.WriteString(cline)
				nbuf.WriteString(delimiter)
				continue
			}
			nbuf.WriteString(cline)
			nbuf.WriteString(delimiter)
			nbuf.WriteString(data)
			nbuf.WriteString(delimiter)
			continue
		}
		nbuf.WriteString(cline)
		nbuf.WriteString(delimiter)
	}

	return writeFile(name, nbuf)
}
