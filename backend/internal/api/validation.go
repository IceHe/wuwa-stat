package api

import (
	"fmt"
	"strings"
	"time"
)

func validateTacetRecord(input tacetRecordInput) (tacetRecordInput, error) {
	if err := validateDate(input.Date); err != nil {
		return tacetRecordInput{}, err
	}
	input.PlayerID = strings.TrimSpace(input.PlayerID)
	if input.PlayerID == "" {
		return tacetRecordInput{}, fmt.Errorf("player_id 不能为空")
	}
	if input.GoldTubes < 0 || input.PurpleTubes < 0 {
		return tacetRecordInput{}, fmt.Errorf("掉落数量不能为负数")
	}
	if input.ClaimCount == 0 {
		input.ClaimCount = 1
	}
	if input.ClaimCount < 1 || input.ClaimCount > 2 {
		return tacetRecordInput{}, fmt.Errorf("claim_count 必须为 1 或 2")
	}
	if input.SolaLevel == 0 {
		input.SolaLevel = 8
	}
	if input.SolaLevel < 1 {
		return tacetRecordInput{}, fmt.Errorf("sola_level 必须大于 0")
	}
	return input, nil
}

func validateAscensionRecord(input ascensionRecordInput) (ascensionRecordInput, error) {
	if err := validateDate(input.Date); err != nil {
		return ascensionRecordInput{}, err
	}
	input.PlayerID = strings.TrimSpace(input.PlayerID)
	if input.PlayerID == "" {
		return ascensionRecordInput{}, fmt.Errorf("player_id 不能为空")
	}
	if input.DropCount < 0 {
		return ascensionRecordInput{}, fmt.Errorf("drop_count 不能为负数")
	}
	if input.SolaLevel == 0 {
		input.SolaLevel = 8
	}
	if input.SolaLevel < 1 {
		return ascensionRecordInput{}, fmt.Errorf("sola_level 必须大于 0")
	}
	return input, nil
}

func validateResonanceRecord(input resonanceRecordInput) (resonanceRecordInput, error) {
	if err := validateDate(input.Date); err != nil {
		return resonanceRecordInput{}, err
	}
	input.PlayerID = strings.TrimSpace(input.PlayerID)
	if input.PlayerID == "" {
		return resonanceRecordInput{}, fmt.Errorf("player_id 不能为空")
	}
	if input.Gold < 0 || input.Purple < 0 || input.Blue < 0 || input.Green < 0 {
		return resonanceRecordInput{}, fmt.Errorf("掉落数量不能为负数")
	}
	if input.ClaimCount == 0 {
		input.ClaimCount = 1
	}
	if input.ClaimCount < 1 || input.ClaimCount > 2 {
		return resonanceRecordInput{}, fmt.Errorf("claim_count 必须为 1 或 2")
	}
	if input.SolaLevel == 0 {
		input.SolaLevel = 8
	}
	if input.SolaLevel < 1 {
		return resonanceRecordInput{}, fmt.Errorf("sola_level 必须大于 0")
	}
	return input, nil
}

func validateDate(value string) error {
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("date 参数无效")
	}
	return nil
}
