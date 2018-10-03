package xmlUtil

import (
	"bytes"
	"container/list"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/polariseye/goutil/stringUtil"
)

// If X2jCharsetReader != nil, it will be used to decode the doc or stream if required
//   import charset "code.google.com/p/go-charset/charset"
//   ...
//   x2j.X2jCharsetReader = charset.NewReader
//   s, err := x2j.DocToJson(doc)
var CharsetReader func(charset string, input io.Reader) (io.Reader, error)

// 从文件加载
// filePath:文件路径
// 返回值:
// *XmlDocument:文档对象
// error:错误信息
func LoadFromFile(filePath string) (*XmlNode, error) {
	data, errMsg := ioutil.ReadFile(filePath)
	if errMsg != nil {
		return nil, errMsg
	}

	return LoadFromByte(data)
}

// 从字节数组加载
// data:文档数据
// 返回值:
// *XmlDocument:文档对象
// error:错误信息
func LoadFromByte(data []byte) (*XmlNode, error) {
	if len(data) <= 0 {
		return nil, fmt.Errorf("xml source have no data")
	}

	return LoadFromString(string(data))
}

// 文档字符串
// doc:文档字符串
// 返回值:
// *XmlDocument:文档对象
// error:错误信息
func LoadFromString(doc string) (*XmlNode, error) {
	// xml.Decoder doesn't properly handle whitespace in some doc
	// see songTextString.xml test case ...
	reg, _ := regexp.Compile("[ \t\n\r]*<")
	doc = reg.ReplaceAllString(doc, "<")
	if stringUtil.IsEmpty(doc) {
		return nil, fmt.Errorf("xml source have no data")
	}

	// 创建解码对象
	b := bytes.NewBufferString(doc)
	decoder := xml.NewDecoder(b)
	decoder.CharsetReader = CharsetReader

	// 把xml转换成树形对象
	return xml2Tree(decoder)
}

// xml转换成树形结构
// decoder:解码对象
// 返回值:
// *XmlNode:结果节点
// error:错误信息
func xml2Tree(decoder *xml.Decoder) (*XmlNode, error) {
	stack := list.New()

	var root *XmlNode
	var nowNode *XmlNode = nil
	for {
		// 获取一个标签
		nowToken, errMsg := decoder.Token()
		if errMsg == io.EOF || nowToken == nil {
			break
		}

		// 如果获取过程存在错误，则直接返回
		if errMsg != nil {
			return root, errMsg
		}

		switch nowToken.(type) {
		case xml.StartElement: //// 起始标签读取
			tmpNowNode := newXmlNode(nowNode)
			if nowNode != nil {
				stack.PushFront(nowNode)                                  //// 入栈
				nowNode.chilldren = append(nowNode.chilldren, tmpNowNode) //// 构造树形结构
			} else {
				// 记录根节点
				root = tmpNowNode
			}

			// 使当前节点指向创建的新节点
			nowNode = tmpNowNode
			startElement := nowToken.(xml.StartElement)
			nowNode.elementName = startElement.Name.Local

			// 属性解析
			if startElement.Attr == nil || len(startElement.Attr) <= 0 {
				continue
			}

			for _, item := range startElement.Attr {
				nowNode.attribute[item.Name.Local] = item.Value
			}

		case xml.EndElement:
			// 已解析完所有节点，则跳出循环
			if stack.Len() <= 0 {
				break
			}

			// 出栈
			isOk := false
			nowNode, isOk = stack.Remove(stack.Front()).(*XmlNode)
			if isOk == false {
				return root, nil
			}

		case xml.CharData:
			// 解析内部文本
			charData := nowToken.(xml.CharData)
			nowNode.innerText = string(charData)
		}
	}

	return root, nil
}
