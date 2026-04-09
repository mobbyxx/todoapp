package domain

import (
	"testing"
)

func TestCalculateLevel(t *testing.T) {
	tests := []struct {
		name     string
		xp       int
		expected int
	}{
		{"negative XP returns level 1", -10, 1},
		{"zero XP is level 1", 0, 1},
		{"1 XP is level 1", 1, 1},
		{"99 XP is level 1", 99, 1},
		{"100 XP is level 2", 100, 2},
		{"150 XP is level 2", 150, 2},
		{"299 XP is level 2", 299, 2},
		{"300 XP is level 3", 300, 3},
		{"500 XP is level 3", 500, 3},
		{"699 XP is level 3", 699, 3},
		{"700 XP is level 4", 700, 4},
		{"1000 XP is level 4", 1000, 4},
		{"1499 XP is level 4", 1499, 4},
		{"1500 XP is level 5", 1500, 5},
		{"2000 XP is level 5", 2000, 5},
		{"2999 XP is level 5", 2999, 5},
		{"3000 XP is level 6", 3000, 6},
		{"4500 XP is level 6", 4500, 6},
		{"5999 XP is level 6", 5999, 6},
		{"6000 XP is level 7", 6000, 7},
		{"9000 XP is level 7", 9000, 7},
		{"11999 XP is level 7", 11999, 7},
		{"12000 XP is level 8", 12000, 8},
		{"15000 XP is level 8", 15000, 8},
		{"100000 XP is level 8", 100000, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateLevel(tt.xp)
			if result != tt.expected {
				t.Errorf("CalculateLevel(%d) = %d, want %d", tt.xp, result, tt.expected)
			}
		})
	}
}

func TestGetXPForNextLevel(t *testing.T) {
	tests := []struct {
		name     string
		currentXP int
		expected int
	}{
		{"level 1 at 0 XP needs 100", 0, 100},
		{"level 1 at 50 XP needs 50", 50, 50},
		{"level 1 at 99 XP needs 1", 99, 1},
		{"level 2 at 100 XP needs 200", 100, 200},
		{"level 2 at 150 XP needs 150", 150, 150},
		{"level 2 at 299 XP needs 1", 299, 1},
		{"level 3 at 300 XP needs 400", 300, 400},
		{"level 3 at 500 XP needs 200", 500, 200},
		{"level 4 at 700 XP needs 800", 700, 800},
		{"level 4 at 1000 XP needs 500", 1000, 500},
		{"level 5 at 1500 XP needs 1500", 1500, 1500},
		{"level 5 at 2000 XP needs 1000", 2000, 1000},
		{"level 6 at 3000 XP needs 3000", 3000, 3000},
		{"level 6 at 4500 XP needs 1500", 4500, 1500},
		{"level 7 at 6000 XP needs 6000", 6000, 6000},
		{"level 7 at 9000 XP needs 3000", 9000, 3000},
		{"level 8 at 12000 XP needs 0 (max)", 12000, 0},
		{"level 8 at 20000 XP needs 0 (max)", 20000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetXPForNextLevel(tt.currentXP)
			if result != tt.expected {
				t.Errorf("GetXPForNextLevel(%d) = %d, want %d", tt.currentXP, result, tt.expected)
			}
		})
	}
}

func TestGetLevelProgress(t *testing.T) {
	tests := []struct {
		name     string
		xp       int
		expected float64
	}{
		{"level 1 at 0 XP is 0%", 0, 0.0},
		{"level 1 at 50 XP is 50%", 50, 50.0},
		{"level 2 at 100 XP is 0%", 100, 0.0},
		{"level 2 at 200 XP is 50%", 200, 50.0},
		{"level 3 at 300 XP is 0%", 300, 0.0},
		{"level 8 at max XP is 100%", 12000, 100.0},
		{"level 8 above max is 100%", 20000, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLevelProgress(tt.xp)
			delta := 0.01
			if result < tt.expected-delta || result > tt.expected+delta {
				t.Errorf("GetLevelProgress(%d) = %f, want %f", tt.xp, result, tt.expected)
			}
		})
	}
}

func TestIsMaxLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected bool
	}{
		{"level 1 is not max", 1, false},
		{"level 7 is not max", 7, false},
		{"level 8 is max", 8, true},
		{"level 9 is max", 9, true},
		{"level 100 is max", 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMaxLevel(tt.level)
			if result != tt.expected {
				t.Errorf("IsMaxLevel(%d) = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestGetLevelName(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{1, "Novice"},
		{2, "Beginner"},
		{3, "Apprentice"},
		{4, "Practitioner"},
		{5, "Expert"},
		{6, "Master"},
		{7, "Grandmaster"},
		{8, "Legend"},
		{9, "Unknown"},
		{0, "Unknown"},
		{-1, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := GetLevelName(tt.level)
			if result != tt.expected {
				t.Errorf("GetLevelName(%d) = %s, want %s", tt.level, result, tt.expected)
			}
		})
	}
}

func TestXPRewardConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"TodoCompleted should be 10", XPRewardTodoCompleted, 10},
		{"StreakBonus7Days should be 50", XPRewardStreakBonus7Days, 50},
		{"StreakBonus30Days should be 200", XPRewardStreakBonus30Days, 200},
		{"PerfectDay should be 25", XPRewardPerfectDay, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, tt.constant)
			}
		})
	}
}

func TestLevelCurve(t *testing.T) {
	expected := []int{0, 100, 300, 700, 1500, 3000, 6000, 12000}

	if len(LevelCurve) != len(expected) {
		t.Errorf("LevelCurve length = %d, want %d", len(LevelCurve), len(expected))
	}

	for i, exp := range expected {
		if i >= len(LevelCurve) {
			break
		}
		if LevelCurve[i] != exp {
			t.Errorf("LevelCurve[%d] = %d, want %d", i, LevelCurve[i], exp)
		}
	}

	if len(LevelCurve) != MaxLevel {
		t.Errorf("LevelCurve length (%d) should match MaxLevel (%d)", len(LevelCurve), MaxLevel)
	}
}

func TestMaxLevel(t *testing.T) {
	if MaxLevel != 8 {
		t.Errorf("MaxLevel = %d, want 8", MaxLevel)
	}
}
