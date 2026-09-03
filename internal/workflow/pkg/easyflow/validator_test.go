package easyflow

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConditionExpression(t *testing.T) {
	valid := []string{
		"$environment = 'prod'",
		"$environment in ('backend', 'frontend')",
		"($level = 'P1' || $level = 'P2') && $owner != 'guest'",
		"$region not in ('test', 'dev')",
	}
	for _, expression := range valid {
		t.Run("valid_"+expression, func(t *testing.T) {
			require.NoError(t, ValidateConditionExpression(expression))
		})
	}

	invalid := []string{
		"1 = 1",
		"$environment = 'prod' OR SLEEP(1)",
		"$environment = 'prod' --",
		"$environment = 'prod\\' OR '1'='1'",
		"$environment == 'prod'",
		"$environment in ()",
	}
	for _, expression := range invalid {
		t.Run("invalid_"+expression, func(t *testing.T) {
			require.Error(t, ValidateConditionExpression(expression))
		})
	}
}

func TestValidateWorkflow(t *testing.T) {
	t.Run("valid serial workflow", func(t *testing.T) {
		require.NoError(t, ValidateWorkflow(validWorkflow()))
	})

	t.Run("rejects disconnected and cyclic graph", func(t *testing.T) {
		wf := validWorkflow()
		wf.FlowData.Nodes = append(wf.FlowData.Nodes,
			map[string]interface{}{"id": "u2", "type": "user", "properties": map[string]interface{}{"approved": []string{"operator"}}},
			map[string]interface{}{"id": "u3", "type": "user", "properties": map[string]interface{}{"approved": []string{"operator"}}},
		)
		wf.FlowData.Edges = append(wf.FlowData.Edges,
			map[string]interface{}{"id": "cycle1", "sourceNodeId": "u2", "targetNodeId": "u3"},
			map[string]interface{}{"id": "cycle2", "sourceNodeId": "u3", "targetNodeId": "u2"},
		)

		err := ValidateWorkflow(wf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "循环")
		assert.Contains(t, err.Error(), "无法从开始节点到达")
	})

	t.Run("rejects unsafe condition", func(t *testing.T) {
		wf := Workflow{
			Name: "condition",
			FlowData: LogicFlow{
				Nodes: []map[string]interface{}{
					{"id": "start", "type": "start"},
					{"id": "condition", "type": "condition"},
					{"id": "end", "type": "end"},
				},
				Edges: []map[string]interface{}{
					{"id": "e1", "sourceNodeId": "start", "targetNodeId": "condition"},
					{"id": "e2", "sourceNodeId": "condition", "targetNodeId": "end", "properties": map[string]interface{}{"expression": "$name = 'admin' OR SLEEP(1)"}},
				},
			},
		}

		err := ValidateWorkflow(wf)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "表达式无效")
	})

	t.Run("rejects incomplete automation", func(t *testing.T) {
		wf := validWorkflow()
		wf.FlowData.Nodes[1] = map[string]interface{}{"id": "user", "type": "automation", "properties": map[string]interface{}{"codebook_uid": "deploy"}}

		err := ValidateWorkflow(wf)
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "执行标签"), err.Error())
	})
}

func TestConverterUsesWorkflowIdentityInEngineSource(t *testing.T) {
	converter := NewDefaultConverter()
	converter.Register(&StartNodeHandler{})
	converter.Register(&EndNodeHandler{})
	converter.Register(&UserNodeHandler{})

	first := validWorkflow()
	first.Id = 101
	second := validWorkflow()
	second.Id = 102

	firstProcess, err := converter.Convert(first)
	require.NoError(t, err)
	secondProcess, err := converter.Convert(second)
	require.NoError(t, err)

	assert.Equal(t, "工单系统:101", firstProcess.Source)
	assert.Equal(t, "工单系统:102", secondProcess.Source)
	assert.NotEqual(t, firstProcess.Source, secondProcess.Source)
}

func validWorkflow() Workflow {
	return Workflow{
		Name: "serial",
		FlowData: LogicFlow{
			Nodes: []map[string]interface{}{
				{"id": "start", "type": "start"},
				{"id": "user", "type": "user", "properties": map[string]interface{}{"approved": []string{"operator"}}},
				{"id": "end", "type": "end"},
			},
			Edges: []map[string]interface{}{
				{"id": "e1", "sourceNodeId": "start", "targetNodeId": "user"},
				{"id": "e2", "sourceNodeId": "user", "targetNodeId": "end"},
			},
		},
	}
}
