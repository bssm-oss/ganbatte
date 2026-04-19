package cmd

import (
	"strings"
	"testing"
)

func TestFindShellInitLines(t *testing.T) {
	tests := []struct {
		name        string
		lines       []string
		wantP10k    int
		wantGnb     int
	}{
		{
			name: "both present, gnb after p10k",
			lines: []string{
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				`  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"`,
				`fi`,
				`eval "$(gnb shell-init)"`,
			},
			wantP10k: 0,
			wantGnb:  3,
		},
		{
			name: "gnb before p10k (correct order)",
			lines: []string{
				`eval "$(gnb shell-init)"`,
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				`fi`,
			},
			wantP10k: 1,
			wantGnb:  0,
		},
		{
			name:     "no p10k",
			lines:    []string{`eval "$(gnb shell-init)"`},
			wantP10k: -1,
			wantGnb:  0,
		},
		{
			name: "no gnb shell-init",
			lines: []string{
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				`fi`,
			},
			wantP10k: 0,
			wantGnb:  -1,
		},
		{
			name:     "neither present",
			lines:    []string{`export PATH="$HOME/bin:$PATH"`},
			wantP10k: -1,
			wantGnb:  -1,
		},
		{
			name: "gnb line commented out",
			lines: []string{
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				`fi`,
				`# eval "$(gnb shell-init)"`,
			},
			wantP10k: 0,
			wantGnb:  -1,
		},
		{
			name: "p10k line commented out",
			lines: []string{
				`# if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				`eval "$(gnb shell-init)"`,
			},
			wantP10k: -1,
			wantGnb:  1,
		},
		{
			name: "only first occurrence matched",
			lines: []string{
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				`  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"`,
				`fi`,
				`eval "$(gnb shell-init)"`,
				`eval "$(gnb shell-init)"`,
			},
			wantP10k: 0,
			wantGnb:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p10k, gnb := findShellInitLines(tt.lines)
			if p10k != tt.wantP10k {
				t.Errorf("p10kLine = %d, want %d", p10k, tt.wantP10k)
			}
			if gnb != tt.wantGnb {
				t.Errorf("gnbLine = %d, want %d", gnb, tt.wantGnb)
			}
		})
	}
}

func TestApplyP10kFix(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		gnbLine  int
		p10kLine int
		wantTop  []string // first N lines of result
	}{
		{
			name: "moves gnb before p10k block with comment",
			lines: []string{
				"# Enable Powerlevel10k instant prompt.",
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				"fi",
				"",
				`export FOO=bar`,
				`eval "$(gnb shell-init)"`,
			},
			gnbLine:  5,
			p10kLine: 1,
			wantTop: []string{
				`eval "$(gnb shell-init)"`,
				"",
				"# Enable Powerlevel10k instant prompt.",
			},
		},
		{
			name: "removes preceding blank line at original position",
			lines: []string{
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				"fi",
				"some other line",
				"",
				`eval "$(gnb shell-init)"`,
			},
			gnbLine:  4,
			p10kLine: 0,
			// blank line at index 3 should be removed
			wantTop: []string{
				`eval "$(gnb shell-init)"`,
				"",
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
			},
		},
		{
			name: "no preceding blank line — no extra removal",
			lines: []string{
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				"fi",
				"some other line",
				`eval "$(gnb shell-init)"`,
			},
			gnbLine:  3,
			p10kLine: 0,
			wantTop: []string{
				`eval "$(gnb shell-init)"`,
				"",
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
			},
		},
		{
			name: "walks back through multiple comment lines",
			lines: []string{
				"# Comment A",
				"# Comment B",
				`if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then`,
				"fi",
				`eval "$(gnb shell-init)"`,
			},
			gnbLine:  4,
			p10kLine: 2,
			wantTop: []string{
				`eval "$(gnb shell-init)"`,
				"",
				"# Comment A",
				"# Comment B",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyP10kFix(tt.lines, tt.gnbLine, tt.p10kLine)
			for i, want := range tt.wantTop {
				if i >= len(result) {
					t.Fatalf("result has only %d lines, want at least %d", len(result), i+1)
				}
				if result[i] != want {
					t.Errorf("line %d = %q, want %q\nfull result:\n%s", i, result[i], want, strings.Join(result, "\n"))
				}
			}
		})
	}
}
