package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintln(os.Stderr, `YouTube CLI

사용법: youtube <명령어> [옵션]

명령어:
  search      키워드로 영상 검색
  videos      채널 영상 목록 조회
  transcript  영상 자막 추출
  install     AI 에이전트용 스킬 파일 설치

각 명령어에 -h 플래그로 도움말을 볼 수 있습니다.`)
}

func errExit(msg string) {
	fmt.Fprintf(os.Stderr, "오류: %s\n", msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "search":
		runSearch(os.Args[2:])
	case "videos":
		runVideos(os.Args[2:])
	case "transcript":
		runTranscript(os.Args[2:])
	case "install":
		runInstall(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "알 수 없는 명령어: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: youtube search <검색어> [-n 개수] [-f json|text]\n")
		fs.PrintDefaults()
	}
	limit := fs.Int("n", 10, "최대 결과 수")
	format := fs.String("f", "json", "출력 형식 (json|text)")
	_ = fs.Parse(sortFlags(args, map[string]bool{"n": true, "f": true}))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "오류: 검색어가 필요합니다")
		fs.Usage()
		os.Exit(1)
	}
	query := strings.Join(fs.Args(), " ")
	data, err := search(query, *limit)
	if err != nil {
		errExit(err.Error())
	}
	fmt.Println(fmtVideos(data, *format))
}

func runVideos(args []string) {
	fs := flag.NewFlagSet("videos", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: youtube videos <channel> [-n 개수] [-f json|text]\n")
		fs.PrintDefaults()
	}
	limit := fs.Int("n", 20, "최대 영상 수")
	format := fs.String("f", "json", "출력 형식 (json|text)")
	_ = fs.Parse(sortFlags(args, map[string]bool{"n": true, "f": true}))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "오류: 채널이 필요합니다")
		fs.Usage()
		os.Exit(1)
	}
	data, err := channelVideos(fs.Arg(0), *limit)
	if err != nil {
		errExit(err.Error())
	}
	fmt.Println(fmtVideos(data, *format))
}

func runTranscript(args []string) {
	fs := flag.NewFlagSet("transcript", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "사용법: youtube transcript [-l ko,en] [-t] [-f json|text] <video>\n")
		fs.PrintDefaults()
	}
	langs := fs.String("l", "", "자막 언어 우선순위, 콤마 구분 (예: ko,en)")
	textOnly := fs.Bool("t", false, "타임스탬프 없이 텍스트만 출력")
	format := fs.String("f", "json", "출력 형식 (json|text)")
	// sort flags before positional args so that flag parsing works regardless of input order
	_ = fs.Parse(sortFlags(args, map[string]bool{"l": true, "f": true}))

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "오류: 영상 ID 또는 URL이 필요합니다")
		fs.Usage()
		os.Exit(1)
	}

	var langList []string
	if *langs != "" {
		for _, l := range strings.Split(*langs, ",") {
			if l = strings.TrimSpace(l); l != "" {
				langList = append(langList, l)
			}
		}
	}

	data, err := transcript(fs.Arg(0), langList)
	if err != nil {
		errExit(err.Error())
	}
	fmt.Println(fmtTranscript(data, *format, *textOnly))
}

// sortFlags reorders args so that flags come before positional arguments,
// allowing flag.FlagSet to parse them even when the user puts flags after args.
func sortFlags(args []string, flagsWithValue map[string]bool) []string {
	var flags, pos []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if len(arg) > 0 && arg[0] == '-' {
			name := strings.TrimLeft(arg, "-")
			if eq := strings.Index(name, "="); eq >= 0 {
				name = name[:eq]
			}
			flags = append(flags, arg)
			if flagsWithValue[name] && i+1 < len(args) && (len(args[i+1]) == 0 || args[i+1][0] != '-') {
				i++
				flags = append(flags, args[i])
			}
		} else {
			pos = append(pos, arg)
		}
		i++
	}
	return append(flags, pos...)
}

