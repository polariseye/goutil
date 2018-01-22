package xmlUtil

import (
	"testing"
)

var xmlData string = `
<Config>	
	<Student Name='你好' Sex="男的">
		<Book Name='书籍' />
	</Student>
	<Student Name='你好2' Sex="男的2">
		<Book Name='书籍2' />
	</Student>
</Config>

`

// 加载正常测试
func Testloadxml(context *testing.T) {
	xml := `
<Config>	
	<Student Name='你好' Sex="男的">
		<Book Name='书籍' />
	</Student>
</Config>
	`
	_, errMsg := LoadFromString(xml)
	if errMsg != nil {
		context.Error(errMsg)
		return
	}
}

// 加载异常测试
func Testloadxml2(context *testing.T) {
	xml := ``
	_, errMsg := LoadFromString(xml)
	if errMsg != nil {
		context.Error(errMsg)
		return
	}
}
