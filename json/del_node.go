package json

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
)

type visitFunc func(*delNode) (stop bool)

const splitter = "."

type delNode struct {
	children map[string]*delNode
	val      string
}

func createDelTree(nodes []string) *delNode {
	root := &delNode{children: make(map[string]*delNode)}
	for _, str := range nodes {
		paths := strings.Split(str, splitter)
		prefix, left := root.findPrefix(paths)
		if len(left) > 0 {
			prefix.add(left...)
		}
	}
	return root
}

func (dt *delNode) hasChildren() bool {
	return len(dt.children) > 0
}

func (dt *delNode) add(steps ...string) {
	parent := dt
	for _, step := range steps {
		n := &delNode{
			children: make(map[string]*delNode),
			val:      step,
		}
		parent.children[step] = n
		parent = n
	}
}

func (dt *delNode) findPrefix(steps []string) (prefix *delNode, left []string) {
	var children []*delNode
	for _, n := range dt.children {
		children = append(children, n)
	}
	return findPrefix(dt, children, steps)
}

func findPrefix(parent *delNode, children []*delNode, steps []string) (prefix *delNode, left []string) {
	if len(steps) == 0 {
		return parent, steps
	}
	step := steps[0]
	var find *delNode
	for _, node := range children {
		if node.val == step {
			find = node
			break
		}
	}
	if find != nil {
		children = children[:0]
		for _, node := range find.children {
			children = append(children, node)
		}
		return findPrefix(find, children, steps[1:len(steps)])
	} else {
		return parent, steps
	}
}

func walkDelTreeBFS(list []*delNode, visit visitFunc) bool {
	size := len(list)
	if size == 0 {
		return false
	}
	for _, node := range list {
		if visit(node) {
			return true
		}
		if node.hasChildren() {
			for _, son := range node.children {
				list = append(list, son)
			}
		}
	}
	newSize := len(list)
	return walkDelTreeBFS(list[size:newSize], visit)
}

func walkJSON(root gjson.Result, dt *delNode) string {
	if !root.IsObject() {
		return root.Raw
	}
	var content []string
	root.ForEach(func(key gjson.Result, val gjson.Result) bool {
		child, ok := dt.children[key.String()]
		if ok {
			if child.hasChildren() {
				if sub := walkJSON(val, child); sub != "" {
					sub = fmt.Sprintf("%s:{%s}", key.Raw, sub)
					content = append(content, sub)
				}
			}
		} else {
			content = append(content, key.Raw+":"+val.Raw)
		}
		return true
	})
	var ret string
	if len(content) != 0 {
		ret = strings.Join(content, ",")
	}
	return ret
}

// TrimJSON 删除json部分属性
func TrimJSON(jsonstr string, paths ...string) string {
	jsonRoot := gjson.Parse(jsonstr)
	delTree := createDelTree(paths)
	res := walkJSON(jsonRoot, delTree)
	return "{" + res + "}"
}
