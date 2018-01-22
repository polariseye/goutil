package xmlUtil

// xpath结构对象
type XPath struct {
	path     string
	pathNode []*xPathNode
}

// xpath解析出的节点
type xPathNode struct {
	PathNodeString string
	AttrData       map[string]string
	Index          int
}

func (this *XPath) Parse() {
	for _, code := range this.path {
		switch code {
		case '/':

		case '[':
		case ']':
		case '=':
		case '@':
		default:
		}
	}
}

// 创建新的xpath节点对象
// _pathNodeString:当前节点对应的节点内容文本
// _index:需要的索引序号
func newXpathNode(_pathNodeString string, _index int) *xPathNode {
	return &xPathNode{
		PathNodeString: _pathNodeString,
		Index:          _index,
		AttrData:       make(map[string]string),
	}
}

// 创建新的Xpath解析对象
// _path:xpath原始文本
// 返回值:
// *XPath:xpath解析对象
func NewXpath(_path string) *XPath {
	return &XPath{
		path:     _path,
		pathNode: make([]*xPathNode, 0),
	}
}
