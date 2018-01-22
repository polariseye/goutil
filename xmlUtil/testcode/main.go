package main

import (
	"fmt"

	"github.com/polariseye/xmldocument"
)

func main() {
	Test1()
}

func Test1() {
	str := `
	<config>
		<student name="你好" sex="男的">
			<H1 name ="asd"/>
		</student>
	</config>
`
	xml, errMsg := xmldocument.LoadFromString(str)
	if errMsg != nil {
		fmt.Println(errMsg)
		return
	}

	node := xml.GetElement("config/student/H1")

	fmt.Println(node.ElementName)
}
