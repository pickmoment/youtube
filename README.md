# youtube

YouTube CLI — API 키 없이 영상 검색, 영상·채널 정보 조회, 채널 목록 조회, 자막 추출을 하는 커맨드라인 도구.

YouTube InnerTube 내부 API를 직접 호출하므로 별도의 API 키나 인증이 필요 없습니다.

## 설치

```bash
go install github.com/pickmoment/youtube@latest
```

또는 소스에서 직접 빌드:

```bash
go build -o youtube .
mv youtube /usr/local/bin/youtube
```

## 명령어

```
youtube <명령어> [옵션]

명령어:
  search      키워드로 영상 검색
  video       영상 기본 정보 조회
  channel     채널 정보 조회
  videos      채널 영상 목록 조회
  transcript  영상 자막 추출
  install     AI 에이전트용 스킬 파일 설치
```

각 명령어에 `-h` 플래그로 상세 도움말을 볼 수 있습니다.

---

### search — 영상 검색

```bash
youtube search <검색어> [-n 개수] [-f json|text]
```

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-n` | 10 | 최대 결과 수 |
| `-f` | json | 출력 형식 (`json` \| `text`) |

```bash
youtube search "파이썬 튜토리얼" -n 5
youtube search "machine learning" -f text
```

---

### video — 영상 정보

```bash
youtube video <video> [-f json|text]
```

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-f` | json | 출력 형식 (`json` \| `text`) |

영상은 ID(11자), 전체 URL, 단축 URL(`youtu.be/...`) 모두 지원합니다.

반환 필드: `id`, `title`, `channel`, `channel_id`, `channel_url`, `duration`, `publish_date`, `upload_date`, `view_count`, `is_live`, `is_private`, `description`, `tags`, `thumbnail`, `url`

| 필드 | 설명 |
|------|------|
| `publish_date` | 게시일 (예: `2026. 5. 13.`) |
| `upload_date` | 상대 날짜 (예: `1일 전`, `2주 전`) |
| `view_count` | 조회수 (순수 숫자 문자열, 예: `"5214"`) |
| `is_live` | 라이브 방송 여부 (`true`일 때만 포함) |
| `is_private` | 비공개 여부 (`true`일 때만 포함) |

```bash
youtube video dQw4w9WgXcQ
youtube video https://youtu.be/dQw4w9WgXcQ -f text
```

---

### channel — 채널 정보

```bash
youtube channel <channel> [-f json|text]
```

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-f` | json | 출력 형식 (`json` \| `text`) |

채널은 핸들(`@veritasium`), 채널 ID(`UCHnyfMqiRRG1u-2MsSQLbXA`), 전체 URL 모두 지원합니다.

반환 필드: `id`, `name`, `handle`, `description`, `subscribers`, `video_count`, `url`

```bash
youtube channel @veritasium
youtube channel UCHnyfMqiRRG1u-2MsSQLbXA -f text
```

---

### videos — 채널 영상 목록

```bash
youtube videos <channel> [-n 개수] [-f json|text]
```

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-n` | 20 | 최대 영상 수 |
| `-f` | json | 출력 형식 (`json` \| `text`) |

채널은 핸들(`@veritasium`), 채널 ID(`UCHnyfMqiRRG1u-2MsSQLbXA`), 전체 URL 모두 지원합니다.

```bash
youtube videos @veritasium -n 10
youtube videos @MKBHD -f text
youtube videos https://www.youtube.com/@veritasium
```

---

### transcript — 자막 추출

```bash
youtube transcript <video> [-l ko,en] [-t] [-f json|text]
```

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-l` | (자동) | 자막 언어 우선순위, 콤마 구분 |
| `-t` | false | 타임스탬프 없이 텍스트만 출력 |
| `-f` | json | 출력 형식 (`json` \| `text`) |

영상은 ID(11자), 전체 URL, 단축 URL(`youtu.be/...`) 모두 지원합니다.  
언어 미지정 시 수동 자막을 우선 선택하고, 없으면 자동 생성 자막(asr)을 사용합니다.

```bash
youtube transcript dQw4w9WgXcQ
youtube transcript https://youtu.be/dQw4w9WgXcQ -l ko,en
youtube transcript dQw4w9WgXcQ -t                # 타임스탬프 없는 텍스트
youtube transcript dQw4w9WgXcQ -f text           # 코드블록 형식
```

---

### install — AI 에이전트 스킬 설치

Claude Code 또는 Codex에서 `/ytb` 명령어로 이 도구를 사용할 수 있도록 스킬 파일을 설치합니다.

```bash
youtube install
```

에이전트(Claude Code / Codex)와 설치 범위(global / project)를 대화형으로 선택합니다.

## 출력 형식

모든 명령어는 `-f` 옵션으로 출력 형식을 선택할 수 있습니다.

- `-f json` (기본): JSON 출력
- `-f text`: 사람이 읽기 쉬운 텍스트 (코드블록)
- `transcript -t`: 타임스탬프 없는 순수 텍스트 (요약 등에 활용)

## 주의사항

- YouTube 내부 API를 직접 호출하므로 구조 변경 시 동작이 불안정해질 수 있습니다.
- 일부 영상은 자막을 제공하지 않거나 자동 생성 자막만 있을 수 있습니다.
