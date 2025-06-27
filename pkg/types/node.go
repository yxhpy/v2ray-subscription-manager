package types

// Node 表示一个V2Ray节点
type Node struct {
	Name       string            `json:"name"`
	Protocol   string            `json:"protocol"`
	Server     string            `json:"server"`
	Port       string            `json:"port"`
	UUID       string            `json:"uuid,omitempty"`
	Method     string            `json:"method,omitempty"`
	Password   string            `json:"password,omitempty"`
	Parameters map[string]string `json:"parameters"`
}

// NodeList 节点列表类型
type NodeList []*Node

// GetByIndex 根据索引获取节点
func (nl NodeList) GetByIndex(index int) *Node {
	if index < 0 || index >= len(nl) {
		return nil
	}
	return nl[index]
}

// Count 返回节点数量
func (nl NodeList) Count() int {
	return len(nl)
}

// FilterByProtocol 根据协议过滤节点
func (nl NodeList) FilterByProtocol(protocol string) NodeList {
	var result NodeList
	for _, node := range nl {
		if node.Protocol == protocol {
			result = append(result, node)
		}
	}
	return result
}
