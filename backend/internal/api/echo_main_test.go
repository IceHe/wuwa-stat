package api

import "testing"

func TestValidateEchoMainRecordTrimsAndDefaults(t *testing.T) {
	record, err := validateEchoMainRecord(echoMainRecordInput{
		Date:        "2026-10-01",
		PlayerID:    " 120000001 ",
		C3MainStat:  " 湮灭伤害 ",
		C1MainStat:  " 攻击力 ",
		TacetDomain: " 伤痕无音区 ",
		EchoSet:     " 熔山裂谷 ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record.PlayerID != "120000001" || record.C3MainStat != "湮灭伤害" || record.C1MainStat != "攻击力" || record.TacetDomain != "伤痕无音区" || record.EchoSet != "熔山裂谷" {
		t.Fatalf("record was not normalized: %+v", record)
	}
	if record.SolaLevel != 8 {
		t.Fatalf("default sola level = %d, want 8", record.SolaLevel)
	}
}

func TestValidateEchoMainRecordRequiresAllFields(t *testing.T) {
	base := echoMainRecordInput{
		Date:        "2026-10-01",
		PlayerID:    "120000001",
		C3MainStat:  "湮灭伤害",
		C1MainStat:  "攻击力",
		TacetDomain: "伤痕无音区",
		EchoSet:     "熔山裂谷",
	}
	tests := []struct {
		name string
		edit func(*echoMainRecordInput)
	}{
		{name: "player", edit: func(value *echoMainRecordInput) { value.PlayerID = "" }},
		{name: "c3", edit: func(value *echoMainRecordInput) { value.C3MainStat = "" }},
		{name: "c1", edit: func(value *echoMainRecordInput) { value.C1MainStat = "" }},
		{name: "domain", edit: func(value *echoMainRecordInput) { value.TacetDomain = "" }},
		{name: "set", edit: func(value *echoMainRecordInput) { value.EchoSet = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := base
			test.edit(&value)
			if _, err := validateEchoMainRecord(value); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
