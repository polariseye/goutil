package xmlUtil

import (
	"container/list"
	"fmt"
	"regexp"
)

// xml节点对象
type XmlNode struct {
	// 节点名
	elementName string

	// 属性集合
	attribute map[string]string

	// 文本数据
	innerText string

	// 子节点
	chilldren []*XmlNode

	// 父节点
	parent *XmlNode
}

// 元素名
func (this *XmlNode) ElementName() string {
	return this.elementName
}

// 查看指定属性
// attrName:待查看的属性名
// 返回值:
// string：属性值
// bool:是否存在指定属性
func (this *XmlNode) Attribute(attrName string) (string, bool) {
	result, isExist := this.attribute[attrName]

	return result, isExist
}

// 获取所有属性
// 返回值:
// map[string]string:属性列表
func (this *XmlNode) ALLAttribute() map[string]string {
	return this.attribute
}

// 内部文本信息
// 返回值:
// string:内部文本信息
func (this *XmlNode) InnerText() string {
	return this.innerText
}

// 子节点
// 返回值:
// []*XmlNode:子节点集合
func (this *XmlNode) Children() []*XmlNode {
	return this.chilldren[:]
}

// 子节点数量
// 返回值:
// int:子节点数量
func (this *XmlNode) ChildLen() int {
	return len(this.chilldren)
}

// 父节点对象
// 返回值:
// *XmlNode:父节点对象，如果没有父节点，则为nil
func (this *XmlNode) Parent() *XmlNode {
	return this.parent
}

// 是否是根节点
// 返回值:
// bool:是否是根节点
func (this *XmlNode) IsRoot() bool {
	return this.parent == nil
}

// 获取指定节点
// xpath:xpath格式的路径信息
// 返回值:
// *XmlNode:结果节点
func (this *XmlNode) GetElement(xpath string) *XmlNode {
	reg, _ := regexp.Compile("/|\\\\")
	elementList := reg.Split(xpath, -1)

	nowNodeList := make([]*XmlNode, 0)
	nowNodeList = append(nowNodeList, this.chilldren...)

	var preOkNode *XmlNode
	for i := 0; i < len(elementList); i++ {
		if nowNodeList == nil || len(nowNodeList) <= 0 {
			return nil
		}

		isFind := false
		for _, childNode := range nowNodeList {
			if elementList[i] == childNode.elementName {
				preOkNode = childNode
				nowNodeList = childNode.chilldren

				isFind = true
			}
		}

		if isFind == false {
			return nil
		}
	}

	return preOkNode
}

// 返回根结点
func (this *XmlNode) Root() *XmlNode {
	parent := this
	for parent.parent != nil {
		parent = parent.parent
	}

	return parent
}

// 输出所有
func (this *XmlNode) OutALL() {
	stack := list.New()
	stack.PushFront(this)

	for {
		if stack.Len() <= 0 {
			break
		}

		nowNode := stack.Front().Value.(*XmlNode)
		stack.Remove(stack.Front())
		for _, item := range nowNode.chilldren {
			stack.PushFront(item)
		}

		fmt.Println("name:", nowNode.elementName, " attr:", nowNode.attribute)
	}
}

// 创建一个新节点
func newXmlNode(parent *XmlNode) *XmlNode {
	return &XmlNode{
		attribute: make(map[string]string),
		chilldren: make([]*XmlNode, 0),
		parent:    parent,
	}
}
