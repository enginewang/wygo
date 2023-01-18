package wygo

import "strings"

type node struct {
	pattern  string  // 待匹配路由，例如 /p/:lang
	part     string  // 路由中的一部分，例如 :lang
	children []*node // 子节点，例如 [doc, tutorial, intro]
	isWild   bool    // 是否精确匹配，part 含有 : 或 * 时为true
}

// 从子节点中找到第一个匹配成功的节点，用于后续的插入
func (n *node) matchChild(part string) *node {
	// 遍历子节点
	for _, child := range n.children {
		// 如果part相同，或者说是*或者:，都算匹配
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 从子节点中找到所有匹配成功的节点，用于后续的查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// 插入到树中
func (n *node) insert(pattern string, parts []string, height int) {
	// 如果当前的高度跟parts一样多，就说明到了最底端，结束递归
	if len(parts) == height {
		// 将这个节点的pattern设置为这个pattern
		n.pattern = pattern
		return
	}
	// 看这一层的part是否有子节点匹配
	part := parts[height]
	child := n.matchChild(part)
	// 没有 电话，就创建一个child，并插入作为它的子节点
	if child == nil {
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}
	// 到下一层
	child.insert(pattern, parts, height+1)
}

func (n *node) search(parts []string, height int) *node {
	// 如果当前的高度跟parts一样多，或者这个节点的前缀有*，就结束递归
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		// 看它的pattern，理论上是有的，没有就返回nil
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	children := n.matchChildren(part)
	for _, child := range children {
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
