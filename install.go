package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

//go:embed skill.md
var skillContent []byte

const skillName = "ytb"

func init() {
	re := regexp.MustCompile(`(?m)^name:\s*(\S+)`)
	m := re.FindSubmatch(skillContent)
	if m == nil || string(m[1]) != skillName {
		panic("skill.md의 name이 skillName(" + skillName + ")과 일치하지 않습니다")
	}
}

type installTarget struct {
	agent string
	dir   string
}

func resolveTarget(agent, scope, home string) (installTarget, error) {
	switch agent {
	case "claude":
		dir := filepath.Join(home, ".claude", "skills")
		if scope == "project" {
			dir = ".claude/skills"
		}
		return installTarget{"Claude Code", dir}, nil
	case "codex":
		dir := filepath.Join(home, ".codex", "skills")
		if scope == "project" {
			dir = ".codex/skills"
		}
		return installTarget{"Codex", dir}, nil
	default:
		return installTarget{}, fmt.Errorf("알 수 없는 에이전트: %s (claude|codex 중 선택)", agent)
	}
}

type prompter struct {
	sc  *bufio.Scanner
	out io.Writer
}

func newPrompter(in io.Reader, out io.Writer) *prompter {
	return &prompter{sc: bufio.NewScanner(in), out: out}
}

func (p *prompter) line(q string) (string, bool) {
	fmt.Fprint(p.out, q)
	if !p.sc.Scan() {
		return "", false
	}
	return strings.TrimSpace(p.sc.Text()), true
}

func (p *prompter) choice(question string, choices []string) string {
	for {
		ans, ok := p.line(fmt.Sprintf("%s [%s]: ", question, strings.Join(choices, "/")))
		if !ok {
			errExit("비대화형 환경입니다. --agent/--scope 플래그를 지정하세요.")
		}
		for _, c := range choices {
			if ans == c {
				return ans
			}
		}
		fmt.Fprintf(p.out, "  %s 중 하나를 입력하세요.\n", strings.Join(choices, "/"))
	}
}

func (p *prompter) confirm(question string, defaultYes bool) bool {
	ans, ok := p.line(question)
	if !ok {
		errExit("비대화형 환경입니다. --yes 플래그를 사용하세요.")
	}
	if ans == "" {
		return defaultYes
	}
	return strings.ToLower(ans) == "y" || strings.ToLower(ans) == "yes"
}

func writeSkillAtomic(dest string, content []byte) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".SKILL.md.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

func runInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: youtube install [--agent claude|codex] [--scope global|project] [--yes] [--backup]")
		fs.PrintDefaults()
	}
	agentFlag := fs.String("agent", "", "에이전트 (claude|codex)")
	scopeFlag := fs.String("scope", "", "설치 범위 (global|project)")
	yes := fs.Bool("yes", false, "확인 프롬프트 없이 설치")
	backup := fs.Bool("backup", false, "기존 파일을 .bak으로 백업 후 덮어쓰기")
	_ = fs.Parse(args)

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		errExit("홈 디렉토리를 찾을 수 없습니다")
	}

	p := newPrompter(os.Stdin, os.Stdout)

	agent := *agentFlag
	if agent == "" {
		agent = p.choice("에이전트를 선택하세요", []string{"claude", "codex"})
	} else if agent != "claude" && agent != "codex" {
		errExit("알 수 없는 에이전트: " + agent + " (claude|codex 중 선택)")
	}

	scope := *scopeFlag
	if scope == "" {
		scope = p.choice("설치 범위를 선택하세요 (global: 모든 프로젝트 / project: 현재 프로젝트만)", []string{"global", "project"})
	} else if scope != "global" && scope != "project" {
		errExit("알 수 없는 범위: " + scope + " (global|project 중 선택)")
	}

	t, err := resolveTarget(agent, scope, home)
	if err != nil {
		errExit(err.Error())
	}
	dest := filepath.Join(t.dir, skillName, "SKILL.md")
	fmt.Printf("\n설치 위치: %s\n", dest)

	if _, pathErr := exec.LookPath("youtube"); pathErr != nil {
		fmt.Println("주의: 'youtube' 바이너리가 PATH에 없습니다. 스킬 사용 전에 설치해 주세요.")
	}

	if !*yes {
		if _, statErr := os.Stat(dest); statErr == nil {
			if !p.confirm("경고: 이미 파일이 존재합니다. 덮어쓸까요? [y/N]: ", false) {
				fmt.Println("취소했습니다.")
				return
			}
		} else {
			if !p.confirm("설치할까요? [Y/n]: ", true) {
				fmt.Println("취소했습니다.")
				return
			}
		}
	}

	if *backup {
		if _, statErr := os.Stat(dest); statErr == nil {
			bak := dest + ".bak"
			if err := os.Rename(dest, bak); err != nil {
				errExit("백업 실패: " + err.Error())
			}
			fmt.Printf("기존 파일을 백업했습니다: %s\n", bak)
		}
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		errExit("디렉토리 생성 실패: " + err.Error())
	}
	if err := writeSkillAtomic(dest, skillContent); err != nil {
		errExit("파일 쓰기 실패: " + err.Error())
	}
	fmt.Printf("\n스킬 설치 완료: %s\n", dest)
	fmt.Printf("%s에서 /%s 으로 호출할 수 있습니다.\n", t.agent, skillName)

	if scope == "project" {
		fmt.Println("팁: 팀과 공유하려면 이 파일을 커밋하세요.")
	}
}
