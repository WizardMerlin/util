package container

const (
	Less ComparisonResult = iota
	Equal
	Greater
)

type (
	ComparisonResult int
	Node             struct {
		Data     interface{}
		Children [2]*Node
	}
	Compare func(a, b interface{}) ComparisonResult
	Tree    struct {
		Compare Compare
		Root    Node
	}
)

func (n *Node) find(data interface{}, cmp Compare, child int, parent *Node) (rchild int, retparent, node *Node) {
	if n.Data == nil {
		return child, parent, n
	}
	switch c := cmp(data, n.Data); c {
	case Equal:
		return child, parent, n
	case Less:
		if n.Children[0] == nil {
			return 0, n, n.Children[0]
		} else {
			return n.Children[0].find(data, cmp, 0, n)
		}
	case Greater:
		if n.Children[1] == nil {
			return 1, n, n.Children[1]
		} else {
			return n.Children[1].find(data, cmp, 1, n)
		}
	default:
		panic(c)
	}
}

func (n *Node) Find(data interface{}, cmp Compare) (child int, parent, node *Node) {
	return n.find(data, cmp, -1, nil)
}

func (n *Node) Walk(ch chan interface{}) {
	if n.Children[0] != nil {
		n.Children[0].Walk(ch)
	}
	if n.Data != nil {
		ch <- n.Data
	}
	if n.Children[1] != nil {
		n.Children[1].Walk(ch)
	}
}

func (n *Node) delete(child int, parent *Node) {
	a, b := n.Children[0], n.Children[1]
	switch {
	case a == nil && b == nil:
		if parent != nil {
			parent.Children[child] = nil
		} else {
			n.Data = nil
		}
	case a == nil && b != nil:
		*n = *b
	case a != nil && b != nil:
		*n = *a
	default:
		if ac := a.Children[1]; ac != nil {
			n.Data = ac.Data
			ac.delete(1, a)
		} else if bc := b.Children[0]; bc != nil {
			n.Data = bc.Data
			bc.delete(0, b)
		}
	}
}

func (t *Tree) Add(data interface{}) {
	child, p, n := t.Root.Find(data, t.Compare)
	if n != nil {
		n.Data = data
	} else if p.Data != nil {
		p.Children[child] = &Node{Data: data}
	} else {
		panic("Both parent and child was null")
	}
}

func (t *Tree) Delete(data interface{}) {
	child, p, n := t.Root.Find(data, t.Compare)
	if n == nil {
		panic("Unable to find that node")
	} else {
		n.delete(child, p)
	}
}