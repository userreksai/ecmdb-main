package easyflow

import (
	"fmt"
	"strings"
)

const maxConditionExpressionLength = 4096

var supportedNodeTypes = map[string]struct{}{
	NodeTypeStart:     {},
	NodeTypeEnd:       {},
	NodeTypeUser:      {},
	NodeTypeCondition: {},
	NodeTypeParallel:  {},
	NodeTypeInclusion: {},
	NodeTypeSelective: {},
	NodeTypeAuto:      {},
	NodeTypeChat:      {},
}

// ValidateWorkflow validates the persisted LogicFlow graph before it is handed
// to easy-workflow. Drafts may remain incomplete, but an invalid graph must not
// be deployed because the engine itself does not validate graph structure.
func ValidateWorkflow(wf Workflow) error {
	var problems []string
	if strings.TrimSpace(wf.Name) == "" {
		problems = append(problems, "流程名称不能为空")
	}

	nodes, err := ParseNodes(wf.FlowData.Nodes)
	if err != nil {
		return fmt.Errorf("流程节点解析失败: %w", err)
	}
	edges, err := parseEdges(wf.FlowData.Edges)
	if err != nil {
		return fmt.Errorf("流程连线解析失败: %w", err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("流程校验失败: 流程至少需要一个开始节点和一个结束节点")
	}

	nodeByID := make(map[string]Node, len(nodes))
	startIDs := make([]string, 0, 1)
	endIDs := make([]string, 0, 1)
	for _, node := range nodes {
		if strings.TrimSpace(node.ID) == "" {
			problems = append(problems, "存在 ID 为空的节点")
			continue
		}
		if _, exists := nodeByID[node.ID]; exists {
			problems = append(problems, fmt.Sprintf("节点 ID 重复: %s", node.ID))
			continue
		}
		if _, supported := supportedNodeTypes[node.Type]; !supported {
			problems = append(problems, fmt.Sprintf("节点 %s 使用了不支持的类型: %s", node.ID, node.Type))
		}
		nodeByID[node.ID] = node
		switch node.Type {
		case NodeTypeStart:
			startIDs = append(startIDs, node.ID)
		case NodeTypeEnd:
			endIDs = append(endIDs, node.ID)
		}
	}

	if len(startIDs) != 1 {
		problems = append(problems, fmt.Sprintf("流程必须且只能有一个开始节点，当前为 %d 个", len(startIDs)))
	}
	if len(endIDs) != 1 {
		problems = append(problems, fmt.Sprintf("流程必须且只能有一个结束节点，当前为 %d 个", len(endIDs)))
	}

	adjacency := make(map[string][]string, len(nodes))
	reverse := make(map[string][]string, len(nodes))
	outgoingEdges := make(map[string][]Edge, len(nodes))
	edgeIDs := make(map[string]struct{}, len(edges))
	edgePairs := make(map[string]struct{}, len(edges))
	for _, edge := range edges {
		if strings.TrimSpace(edge.ID) == "" {
			problems = append(problems, "存在 ID 为空的连线")
		} else if _, exists := edgeIDs[edge.ID]; exists {
			problems = append(problems, fmt.Sprintf("连线 ID 重复: %s", edge.ID))
		} else {
			edgeIDs[edge.ID] = struct{}{}
		}

		source, sourceOK := nodeByID[edge.SourceNodeId]
		target, targetOK := nodeByID[edge.TargetNodeId]
		if !sourceOK || !targetOK {
			problems = append(problems, fmt.Sprintf("连线 %s 引用了不存在的节点", edge.ID))
			continue
		}
		if edge.SourceNodeId == edge.TargetNodeId {
			problems = append(problems, fmt.Sprintf("节点 %s 不能连接自身", edge.SourceNodeId))
			continue
		}
		if source.Type == NodeTypeEnd {
			problems = append(problems, fmt.Sprintf("结束节点 %s 不能连接下级节点", source.ID))
		}
		if target.Type == NodeTypeStart {
			problems = append(problems, fmt.Sprintf("开始节点 %s 不能连接上级节点", target.ID))
		}

		pair := edge.SourceNodeId + "\x00" + edge.TargetNodeId
		if _, exists := edgePairs[pair]; exists {
			problems = append(problems, fmt.Sprintf("节点 %s 到 %s 存在重复连线", edge.SourceNodeId, edge.TargetNodeId))
			continue
		}
		edgePairs[pair] = struct{}{}
		adjacency[edge.SourceNodeId] = append(adjacency[edge.SourceNodeId], edge.TargetNodeId)
		reverse[edge.TargetNodeId] = append(reverse[edge.TargetNodeId], edge.SourceNodeId)
		outgoingEdges[edge.SourceNodeId] = append(outgoingEdges[edge.SourceNodeId], edge)
	}

	for id, node := range nodeByID {
		if node.Type != NodeTypeStart && len(reverse[id]) == 0 {
			problems = append(problems, fmt.Sprintf("节点 %s 没有上级连线", id))
		}
		if node.Type != NodeTypeEnd && len(adjacency[id]) == 0 {
			problems = append(problems, fmt.Sprintf("节点 %s 没有下级连线", id))
		}
	}

	if len(startIDs) == 1 {
		reachable := walkGraph(startIDs[0], adjacency)
		for id := range nodeByID {
			if !reachable[id] {
				problems = append(problems, fmt.Sprintf("节点 %s 无法从开始节点到达", id))
			}
		}
	}
	if len(endIDs) == 1 {
		canReachEnd := walkGraph(endIDs[0], reverse)
		for id := range nodeByID {
			if !canReachEnd[id] {
				problems = append(problems, fmt.Sprintf("节点 %s 无法到达结束节点", id))
			}
		}
	}
	if hasCycle(nodeByID, adjacency) {
		problems = append(problems, "流程中存在循环连线")
	}

	for _, node := range nodes {
		switch node.Type {
		case NodeTypeCondition:
			problems = append(problems, validateConditionNode(node, outgoingEdges[node.ID])...)
		case NodeTypeSelective:
			for _, targetID := range adjacency[node.ID] {
				if target, ok := nodeByID[targetID]; ok && target.Type != NodeTypeCondition {
					problems = append(problems, fmt.Sprintf("条件并行节点 %s 只能连接条件节点", node.ID))
				}
			}
		case NodeTypeUser:
			property, propertyErr := ToNodeProperty[UserProperty](node)
			if propertyErr != nil {
				problems = append(problems, fmt.Sprintf("人工节点 %s 配置解析失败", node.ID))
				continue
			}
			if !hasUsableAssignee(property.NormalizeAssignees()) {
				problems = append(problems, fmt.Sprintf("人工节点 %s 未配置审批人策略", node.ID))
			}
		case NodeTypeAuto:
			property, propertyErr := ToNodeProperty[AutomationProperty](node)
			if propertyErr != nil {
				problems = append(problems, fmt.Sprintf("自动化节点 %s 配置解析失败", node.ID))
				continue
			}
			if strings.TrimSpace(property.CodebookUid) == "" {
				problems = append(problems, fmt.Sprintf("自动化节点 %s 未配置代码模板", node.ID))
			}
			if strings.TrimSpace(property.Tag) == "" {
				problems = append(problems, fmt.Sprintf("自动化节点 %s 未配置执行标签", node.ID))
			}
			if property.IsTiming {
				switch property.ExecMethod {
				case "hand":
					if property.Unit < 1 || property.Unit > 3 || property.Quantity <= 0 {
						problems = append(problems, fmt.Sprintf("自动化节点 %s 的定时参数无效", node.ID))
					}
				case "template":
					if property.TemplateId <= 0 || strings.TrimSpace(property.TemplateField) == "" {
						problems = append(problems, fmt.Sprintf("自动化节点 %s 的定时模板配置无效", node.ID))
					}
				default:
					problems = append(problems, fmt.Sprintf("自动化节点 %s 未配置定时方式", node.ID))
				}
			}
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("流程校验失败: %s", strings.Join(uniqueStrings(problems), "；"))
	}
	return nil
}

func validateConditionNode(node Node, edges []Edge) []string {
	problems := make([]string, 0)
	defaultBranches := 0
	for _, edge := range edges {
		property, err := ToEdgeProperty(edge)
		if err != nil {
			problems = append(problems, fmt.Sprintf("条件节点 %s 的连线 %s 配置解析失败", node.ID, edge.ID))
			continue
		}
		expression := strings.TrimSpace(property.Expression)
		if expression == "" {
			defaultBranches++
			continue
		}
		if err = ValidateConditionExpression(expression); err != nil {
			problems = append(problems, fmt.Sprintf("条件节点 %s 的连线 %s 表达式无效: %v", node.ID, edge.ID, err))
		}
	}
	if defaultBranches > 1 {
		problems = append(problems, fmt.Sprintf("条件节点 %s 最多只能配置一条无条件分支", node.ID))
	}
	return problems
}

func hasUsableAssignee(assignees []Assignee) bool {
	for _, assignee := range assignees {
		if assignee.Rule == FOUNDER || assignee.Rule == LEADER || assignee.Rule == MAIN_LEADER {
			return true
		}
		for _, value := range assignee.Values {
			if strings.TrimSpace(value) != "" {
				return true
			}
		}
	}
	return false
}

func walkGraph(start string, graph map[string][]string) map[string]bool {
	visited := map[string]bool{start: true}
	queue := []string{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range graph[current] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return visited
}

func hasCycle(nodes map[string]Node, graph map[string][]string) bool {
	const (
		unvisited = iota
		visiting
		visited
	)
	states := make(map[string]int, len(nodes))
	var visit func(string) bool
	visit = func(id string) bool {
		states[id] = visiting
		for _, next := range graph[id] {
			if states[next] == visiting || states[next] == unvisited && visit(next) {
				return true
			}
		}
		states[id] = visited
		return false
	}
	for id := range nodes {
		if states[id] == unvisited && visit(id) {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// ValidateConditionExpression accepts only the expression language generated
// by the UI: $variable comparisons joined by && / || and parentheses. This is
// intentionally strict because easy-workflow evaluates the final expression in
// MySQL; accepting SQL syntax or function calls here would be unsafe.
func ValidateConditionExpression(expression string) error {
	if strings.TrimSpace(expression) == "" {
		return fmt.Errorf("表达式不能为空")
	}
	if len(expression) > maxConditionExpressionLength {
		return fmt.Errorf("表达式长度不能超过 %d", maxConditionExpressionLength)
	}
	p := expressionParser{input: expression}
	if err := p.parseOr(); err != nil {
		return err
	}
	p.skipSpaces()
	if !p.eof() {
		return fmt.Errorf("位置 %d 存在不支持的语法", p.pos+1)
	}
	return nil
}

type expressionParser struct {
	input string
	pos   int
}

func (p *expressionParser) parseOr() error {
	if err := p.parseAnd(); err != nil {
		return err
	}
	for {
		p.skipSpaces()
		if !p.consume("||") {
			return nil
		}
		if err := p.parseAnd(); err != nil {
			return err
		}
	}
}

func (p *expressionParser) parseAnd() error {
	if err := p.parsePrimary(); err != nil {
		return err
	}
	for {
		p.skipSpaces()
		if !p.consume("&&") {
			return nil
		}
		if err := p.parsePrimary(); err != nil {
			return err
		}
	}
}

func (p *expressionParser) parsePrimary() error {
	p.skipSpaces()
	if p.consume("(") {
		if err := p.parseOr(); err != nil {
			return err
		}
		p.skipSpaces()
		if !p.consume(")") {
			return fmt.Errorf("位置 %d 缺少右括号", p.pos+1)
		}
		return nil
	}
	return p.parseComparison()
}

func (p *expressionParser) parseComparison() error {
	p.skipSpaces()
	if !p.consume("$") || !p.consumeIdentifier() {
		return fmt.Errorf("位置 %d 必须使用 $变量名", p.pos+1)
	}
	p.skipSpaces()

	if p.consumeWord("not") {
		p.skipSpaces()
		if !p.consumeWord("in") {
			return fmt.Errorf("位置 %d 缺少 in", p.pos+1)
		}
		return p.parseStringList()
	}
	if p.consumeWord("in") {
		return p.parseStringList()
	}
	if !(p.consume("!=") || p.consume("=") || p.consume(">") || p.consume("<")) {
		return fmt.Errorf("位置 %d 使用了不支持的比较符", p.pos+1)
	}
	return p.parseQuotedString()
}

func (p *expressionParser) parseStringList() error {
	p.skipSpaces()
	if !p.consume("(") {
		return fmt.Errorf("位置 %d 的 in 条件必须使用值列表", p.pos+1)
	}
	if err := p.parseQuotedString(); err != nil {
		return err
	}
	for {
		p.skipSpaces()
		if !p.consume(",") {
			break
		}
		if err := p.parseQuotedString(); err != nil {
			return err
		}
	}
	p.skipSpaces()
	if !p.consume(")") {
		return fmt.Errorf("位置 %d 的值列表缺少右括号", p.pos+1)
	}
	return nil
}

func (p *expressionParser) parseQuotedString() error {
	p.skipSpaces()
	if !p.consume("'") {
		return fmt.Errorf("位置 %d 的比较值必须使用单引号", p.pos+1)
	}
	for !p.eof() {
		char := p.input[p.pos]
		switch char {
		case '\'', '\\', '\r', '\n', 0:
			if char == '\'' {
				p.pos++
				return nil
			}
			return fmt.Errorf("位置 %d 的比较值包含不安全字符", p.pos+1)
		default:
			p.pos++
		}
	}
	return fmt.Errorf("比较值缺少结束单引号")
}

func (p *expressionParser) consumeIdentifier() bool {
	start := p.pos
	for !p.eof() {
		char := p.input[p.pos]
		if !isIdentifierChar(char) {
			break
		}
		p.pos++
	}
	return p.pos > start
}

func (p *expressionParser) consumeWord(word string) bool {
	p.skipSpaces()
	end := p.pos + len(word)
	if end > len(p.input) || !strings.EqualFold(p.input[p.pos:end], word) {
		return false
	}
	if end < len(p.input) && isIdentifierChar(p.input[end]) {
		return false
	}
	p.pos = end
	return true
}

func (p *expressionParser) consume(token string) bool {
	if strings.HasPrefix(p.input[p.pos:], token) {
		p.pos += len(token)
		return true
	}
	return false
}

func (p *expressionParser) skipSpaces() {
	for !p.eof() {
		switch p.input[p.pos] {
		case ' ', '\t', '\r', '\n':
			p.pos++
		default:
			return
		}
	}
}

func (p *expressionParser) eof() bool {
	return p.pos >= len(p.input)
}

func isIdentifierChar(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_'
}
