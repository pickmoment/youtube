package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed skill.md
var skillContent []byte

const skillName = "ytb"

type installTarget struct {
	agent string
	dir   string
}

var agentTargets = map[string]map[string]installTarget{
	"claude": {
		"global":  {"Claude Code", filepath.Join(os.Getenv("HOME"), ".claude", "skills")},
		"project": {"Claude Code", ".claude/skills"},
	},
	"codex": {
		"global":  {"Codex", filepath.Join(os.Getenv("HOME"), ".agents", "skills")},
		"project": {"Codex", ".agents/skills"},
	},
}

func runInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: youtube install [--agent claude|codex] [--scope global|project] [--yes]")
		fs.PrintDefaults()
	}
	agentFlag := fs.String("agent", "", "에이전트 (claude|codex)")
	scopeFlag := fs.String("scope", "", "설치 범위 (global|project)")
	yes := fs.Bool("yes", false, "확인 프롬프트 없이 설치")
	_ = fs.Parse(args)

	agent := *agentFlag
	if agent == "" {
		agent = promptChoice(os.Stdin, os.Stdout, "에이전트를 선택하세요", []string{"claude", "codex"})
	} else if agentTargets[agent] == nil {
		errExit("알 수 없는 에이전트: " + agent + " (claude|codex 중 선택)")
	}

	scope := *scopeFlag
	if scope == "" {
		scope = promptChoice(os.Stdin, os.Stdout, "설치 범위를 선택하세요 (global: 모든 프로젝트 / project: 현재 프로젝트만)", []string{"global", "project"})
	} else if scope != "global" && scope != "project" {
		errExit("알 수 없는 범위: " + scope + " (global|project 중 선택)")
	}

	t := agentTargets[agent][scope]
	dest := filepath.Join(t.dir, skillName, "SKILL.md")
	fmt.Printf("\n설치 위치: %s\n", dest)

	if !*yes {
		if _, err := os.Stat(dest); err == nil {
			if !promptConfirm(os.Stdin, os.Stdout, "경고: 이미 파일이 존재합니다. 덮어쓸까요? [y/N]: ", false) {
				fmt.Println("취소했습니다.")
				return
			}
		} else {
			if !promptConfirm(os.Stdin, os.Stdout, "설치할까요? [Y/n]: ", true) {
				fmt.Println("취소했습니다.")
				return
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		errExit("디렉토리 생성 실패: " + err.Error())
	}
	if err := os.WriteFile(dest, skillContent, 0644); err != nil {
		errExit("파일 쓰기 실패: " + err.Error())
	}
	fmt.Printf("\n스킬 설치 완료: %s\n", dest)
	fmt.Printf("%s에서 /%s 으로 호출할 수 있습니다.\n", t.agent, skillName)

	if _, err := exec.LookPath("youtube"); err != nil {
		fmt.Println("\n주의: 'youtube' 바이너리가 PATH에 없습니다. 스킬 사용 전에 설치해 주세요.")
	}
}

func promptChoice(in io.Reader, out io.Writer, question string, choices []string) string {
	for {
		fmt.Fprintf(out, "%s [%s]: ", question, strings.Join(choices, "/"))
		var ans string
		fmt.Fscan(in, &ans)
		for _, c := range choices {
			if ans == c {
				return ans
			}
		}
		fmt.Fprintf(out, "  %s 중 하나를 입력하세요.\n", strings.Join(choices, "/"))
	}
}

func promptConfirm(in io.Reader, out io.Writer, question string, defaultYes bool) bool {
	fmt.Fprint(out, question)
	var ans string
	fmt.Fscan(in, &ans)
	ans = strings.ToLower(strings.TrimSpace(ans))
	if ans == "" {
		return defaultYes
	}
	return ans == "y" || ans == "yes"
}
