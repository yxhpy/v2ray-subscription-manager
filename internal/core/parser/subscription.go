package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// fetchSubscription 从URL获取订阅内容
func fetchSubscription(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("获取订阅失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应内容失败: %v", err)
	}

	return string(body), nil
}

// decodeBase64 智能解码base64内容
func decodeBase64(content string) (string, error) {
	// 去除空白字符
	content = strings.TrimSpace(content)

	// 检查是否看起来像base64编码
	// base64字符集: A-Z, a-z, 0-9, +, /, =
	isBase64Like := true
	for _, char := range content {
		if !((char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '+' || char == '/' || char == '=' ||
			char == '\n' || char == '\r' || char == ' ') {
			isBase64Like = false
			break
		}
	}

	// 如果不像base64编码，或者内容包含协议前缀，直接返回原始内容
	if !isBase64Like ||
		strings.Contains(content, "://") ||
		strings.HasPrefix(content, "vmess://") ||
		strings.HasPrefix(content, "vless://") ||
		strings.HasPrefix(content, "ss://") ||
		strings.HasPrefix(content, "trojan://") ||
		strings.HasPrefix(content, "hysteria2://") {
		fmt.Printf("📄 内容不是base64编码，直接解析\n")
		return content, nil
	}

	// 尝试base64解码
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		// 如果base64解码失败，返回原始内容
		fmt.Printf("⚠️ base64解码失败，使用原始内容: %v\n", err)
		return content, nil
	}

	fmt.Printf("✅ base64解码成功\n")
	return string(decoded), nil
}

// parseHysteria2 解析hysteria2协议链接
func parseHysteria2(link string) (*types.Node, error) {
	// hysteria2://user@server:port?params#name
	link = strings.TrimPrefix(link, "hysteria2://")

	// 分离锚点（名称）
	parts := strings.Split(link, "#")
	if len(parts) != 2 {
		return nil, fmt.Errorf("hysteria2链接格式错误")
	}

	mainPart := parts[0]
	name, _ := url.QueryUnescape(parts[1])

	// 分离参数
	urlParts := strings.Split(mainPart, "?")
	addressPart := urlParts[0]

	// 解析地址部分 user@server:port
	atIndex := strings.LastIndex(addressPart, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("hysteria2地址格式错误")
	}

	user := addressPart[:atIndex]
	serverPort := addressPart[atIndex+1:]

	// 分离服务器和端口
	colonIndex := strings.LastIndex(serverPort, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("端口格式错误")
	}

	server := serverPort[:colonIndex]
	port := serverPort[colonIndex+1:]

	// 解析参数
	parameters := make(map[string]string)
	if len(urlParts) > 1 {
		queryParams, _ := url.ParseQuery(urlParts[1])
		for key, values := range queryParams {
			if len(values) > 0 {
				parameters[key] = values[0]
			}
		}
	}

	return &types.Node{
		Name:       name,
		Protocol:   "hysteria2",
		Server:     server,
		Port:       port,
		UUID:       user,
		Parameters: parameters,
	}, nil
}

// parseVless 解析vless协议链接
func parseVless(link string) (*types.Node, error) {
	// vless://uuid@server:port?params#name
	link = strings.TrimPrefix(link, "vless://")

	// 分离锚点（名称）
	parts := strings.Split(link, "#")
	if len(parts) != 2 {
		return nil, fmt.Errorf("vless链接格式错误")
	}

	mainPart := parts[0]
	name, _ := url.QueryUnescape(parts[1])

	// 分离参数
	urlParts := strings.Split(mainPart, "?")
	addressPart := urlParts[0]

	// 解析地址部分 uuid@server:port
	atIndex := strings.LastIndex(addressPart, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("vless地址格式错误")
	}

	uuid := addressPart[:atIndex]
	serverPort := addressPart[atIndex+1:]

	// 分离服务器和端口
	colonIndex := strings.LastIndex(serverPort, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("端口格式错误")
	}

	server := serverPort[:colonIndex]
	port := serverPort[colonIndex+1:]

	// 解析参数
	parameters := make(map[string]string)
	if len(urlParts) > 1 {
		queryParams, _ := url.ParseQuery(urlParts[1])
		for key, values := range queryParams {
			if len(values) > 0 {
				parameters[key] = values[0]
			}
		}
	}

	return &types.Node{
		Name:       name,
		Protocol:   "vless",
		Server:     server,
		Port:       port,
		UUID:       uuid,
		Parameters: parameters,
	}, nil
}

// parseSS 解析ss协议链接
func parseSS(link string) (*types.Node, error) {
	// ss://base64编码或method:password@server:port#name
	link = strings.TrimPrefix(link, "ss://")

	// 分离锚点（名称）
	parts := strings.Split(link, "#")
	if len(parts) != 2 {
		return nil, fmt.Errorf("ss链接格式错误")
	}

	mainPart := parts[0]
	name, _ := url.QueryUnescape(parts[1])

	var server, port, method, password string
	var parameters map[string]string = make(map[string]string)

	// 检查是否有查询参数
	urlParts := strings.Split(mainPart, "?")
	addressPart := urlParts[0]

	// 解析参数
	if len(urlParts) > 1 {
		queryParams, _ := url.ParseQuery(urlParts[1])
		for key, values := range queryParams {
			if len(values) > 0 {
				parameters[key] = values[0]
			}
		}
	}

	// 尝试解析 method:password@server:port 格式
	atIndex := strings.LastIndex(addressPart, "@")
	if atIndex != -1 {
		// 格式: method:password@server:port
		methodPassword := addressPart[:atIndex]
		serverPort := addressPart[atIndex+1:]

		colonIndex := strings.Index(methodPassword, ":")
		if colonIndex != -1 {
			method = methodPassword[:colonIndex]
			password = methodPassword[colonIndex+1:]
		} else {
			// 可能是base64编码的部分
			decoded, err := base64.StdEncoding.DecodeString(methodPassword)
			if err == nil {
				decodedStr := string(decoded)
				colonIndex := strings.Index(decodedStr, ":")
				if colonIndex != -1 {
					method = decodedStr[:colonIndex]
					password = decodedStr[colonIndex+1:]
				}
			}
		}

		// 解析server:port
		colonIndex = strings.LastIndex(serverPort, ":")
		if colonIndex != -1 {
			server = serverPort[:colonIndex]
			port = serverPort[colonIndex+1:]
		}
	} else {
		// 尝试整体base64解码
		decoded, err := base64.StdEncoding.DecodeString(addressPart)
		if err != nil {
			return nil, fmt.Errorf("ss链接解码失败: %v", err)
		}

		decodedStr := string(decoded)
		// 格式应该是 method:password@server:port
		atIndex = strings.LastIndex(decodedStr, "@")
		if atIndex == -1 {
			return nil, fmt.Errorf("ss解码后格式错误")
		}

		methodPassword := decodedStr[:atIndex]
		serverPort := decodedStr[atIndex+1:]

		colonIndex := strings.Index(methodPassword, ":")
		if colonIndex != -1 {
			method = methodPassword[:colonIndex]
			password = methodPassword[colonIndex+1:]
		}

		colonIndex = strings.LastIndex(serverPort, ":")
		if colonIndex != -1 {
			server = serverPort[:colonIndex]
			port = serverPort[colonIndex+1:]
		}
	}

	return &types.Node{
		Name:       name,
		Protocol:   "ss",
		Server:     server,
		Port:       port,
		Method:     method,
		Password:   password,
		Parameters: parameters,
	}, nil
}

// parseVmess 解析vmess协议链接
func parseVmess(link string) (*types.Node, error) {
	// vmess://base64编码的JSON配置
	link = strings.TrimPrefix(link, "vmess://")

	// 解码base64
	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		return nil, fmt.Errorf("vmess链接base64解码失败: %v", err)
	}

	// 解析JSON
	var config map[string]interface{}
	if err := json.Unmarshal(decoded, &config); err != nil {
		return nil, fmt.Errorf("vmess配置JSON解析失败: %v", err)
	}

	// 提取基本信息
	name := ""
	if ps, ok := config["ps"].(string); ok {
		name = ps
	}

	server := ""
	if add, ok := config["add"].(string); ok {
		server = add
	}

	port := ""
	if p, ok := config["port"].(float64); ok {
		port = fmt.Sprintf("%.0f", p)
	} else if p, ok := config["port"].(string); ok {
		port = p
	}

	uuid := ""
	if id, ok := config["id"].(string); ok {
		uuid = id
	}

	// 提取其他参数
	parameters := make(map[string]string)

	if aid, ok := config["aid"].(float64); ok {
		parameters["aid"] = fmt.Sprintf("%.0f", aid)
	}
	if scy, ok := config["scy"].(string); ok {
		parameters["scy"] = scy
	}
	if net, ok := config["net"].(string); ok {
		parameters["net"] = net
	}
	if vType, ok := config["type"].(string); ok {
		parameters["type"] = vType
	}
	if tls, ok := config["tls"].(string); ok {
		parameters["tls"] = tls
	}
	if host, ok := config["host"].(string); ok {
		parameters["host"] = host
	}
	if path, ok := config["path"].(string); ok {
		parameters["path"] = path
	}
	if v, ok := config["v"].(string); ok {
		parameters["v"] = v
	}
	if alpn, ok := config["alpn"].(string); ok {
		parameters["alpn"] = alpn
	}
	if fp, ok := config["fp"].(string); ok {
		parameters["fp"] = fp
	}
	if sni, ok := config["sni"].(string); ok {
		parameters["sni"] = sni
	}

	return &types.Node{
		Name:       name,
		Protocol:   "vmess",
		Server:     server,
		Port:       port,
		UUID:       uuid,
		Parameters: parameters,
	}, nil
}

// parseTrojan 解析trojan协议链接
func parseTrojan(link string) (*types.Node, error) {
	// trojan://password@server:port?params#name
	link = strings.TrimPrefix(link, "trojan://")

	// 分离锚点（名称）
	parts := strings.Split(link, "#")
	if len(parts) != 2 {
		return nil, fmt.Errorf("trojan链接格式错误")
	}

	mainPart := parts[0]
	name, _ := url.QueryUnescape(parts[1])

	// 分离参数
	urlParts := strings.Split(mainPart, "?")
	addressPart := urlParts[0]

	// 解析地址部分 password@server:port
	atIndex := strings.LastIndex(addressPart, "@")
	if atIndex == -1 {
		return nil, fmt.Errorf("trojan地址格式错误")
	}

	password := addressPart[:atIndex]
	serverPort := addressPart[atIndex+1:]

	// 分离服务器和端口
	colonIndex := strings.LastIndex(serverPort, ":")
	if colonIndex == -1 {
		return nil, fmt.Errorf("端口格式错误")
	}

	server := serverPort[:colonIndex]
	port := serverPort[colonIndex+1:]

	// 解析参数
	parameters := make(map[string]string)
	if len(urlParts) > 1 {
		queryParams, _ := url.ParseQuery(urlParts[1])
		for key, values := range queryParams {
			if len(values) > 0 {
				parameters[key] = values[0]
			}
		}
	}

	return &types.Node{
		Name:       name,
		Protocol:   "trojan",
		Server:     server,
		Port:       port,
		Password:   password,
		Parameters: parameters,
	}, nil
}

// parseLinks 解析所有链接
func parseLinks(content string) ([]*types.Node, error) {
	var nodes []*types.Node
	var errors []string

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var node *types.Node
		var err error

		if strings.HasPrefix(line, "hysteria2://") {
			node, err = parseHysteria2(line)
		} else if strings.HasPrefix(line, "vless://") {
			node, err = parseVless(line)
		} else if strings.HasPrefix(line, "ss://") {
			node, err = parseSS(line)
		} else if strings.HasPrefix(line, "vmess://") {
			node, err = parseVmess(line)
		} else if strings.HasPrefix(line, "trojan://") {
			node, err = parseTrojan(line)
		} else {
			errors = append(errors, fmt.Sprintf("第%d行: 不支持的协议 %s", i+1, line[:min(20, len(line))]))
			continue
		}

		if err != nil {
			errors = append(errors, fmt.Sprintf("第%d行解析失败: %v", i+1, err))
			continue
		}

		if node != nil {
			nodes = append(nodes, node)
		}
	}

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "解析警告:\n")
		for _, errMsg := range errors {
			fmt.Fprintf(os.Stderr, "  %s\n", errMsg)
		}
	}

	return nodes, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParseSubscription 解析订阅链接并返回JSON结果
func ParseSubscription(subscriptionURL string) error {
	// 获取订阅内容
	fmt.Fprintf(os.Stderr, "正在获取订阅内容...\n")
	content, err := fetchSubscription(subscriptionURL)
	if err != nil {
		return fmt.Errorf("获取订阅失败: %v", err)
	}

	// 解码base64
	fmt.Fprintf(os.Stderr, "正在解码内容...\n")
	decoded, err := decodeBase64(content)
	if err != nil {
		return fmt.Errorf("解码失败: %v", err)
	}

	// 解析所有链接
	fmt.Fprintf(os.Stderr, "正在解析链接...\n")
	nodes, err := parseLinks(decoded)
	if err != nil {
		return fmt.Errorf("解析失败: %v", err)
	}

	// 输出JSON
	result := map[string]interface{}{
		"total": len(nodes),
		"nodes": nodes,
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	fmt.Println(string(jsonOutput))
	fmt.Fprintf(os.Stderr, "解析完成，共找到 %d 个节点\n", len(nodes))
	return nil
}
