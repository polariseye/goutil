package query

import "github.com/polariseye/goutil/xmlUtil/gxpath/xpath"

type Iterator interface {
	Current() xpath.NodeNavigator
}
