package execution

import "anserflow/internal/convert"

type AcceptanceCheckInput struct {
	Type          string `json:"type"`
	Command       string `json:"command,omitempty"`
	PassCondition string `json:"pass_condition"`
}

type RetryPolicyInput struct {
	MaxAttempts     int  `json:"max_attempts"`
	AllowAutoRepair bool `json:"allow_auto_repair"`
}

type MergePolicyInput struct {
	AllowAutoPR        bool `json:"allow_auto_pr"`
	AllowAutoMerge     bool `json:"allow_auto_merge"`
	RequireHumanReview bool `json:"require_human_review"`
}

type RollbackPolicyInput struct {
	Strategy          string   `json:"strategy"`
	TriggerConditions []string `json:"trigger_conditions"`
}

func ToAcceptanceCheckData(items []AcceptanceCheckInput) []convert.AcceptanceCheckData {
	out := make([]convert.AcceptanceCheckData, 0, len(items))
	for _, item := range items {
		out = append(out, convert.AcceptanceCheckData{
			Type:          item.Type,
			Command:       item.Command,
			PassCondition: item.PassCondition,
		})
	}
	return out
}

func ToRetryPolicyData(in RetryPolicyInput) convert.RetryPolicyData {
	return convert.RetryPolicyData{
		MaxAttempts:     in.MaxAttempts,
		AllowAutoRepair: in.AllowAutoRepair,
	}
}

func ToMergePolicyData(in MergePolicyInput) convert.MergePolicyData {
	return convert.MergePolicyData{
		AllowAutoPR:        in.AllowAutoPR,
		AllowAutoMerge:     in.AllowAutoMerge,
		RequireHumanReview: in.RequireHumanReview,
	}
}

func ToRollbackPolicyData(in RollbackPolicyInput) convert.RollbackPolicyData {
	return convert.RollbackPolicyData{
		Strategy:          in.Strategy,
		TriggerConditions: in.TriggerConditions,
	}
}
