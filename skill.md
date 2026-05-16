---
name: ytb
description: YouTube 영상·채널 정보 조회, 영상 검색, 채널 영상 목록, 자막 추출을 제공하는 스킬. API 키 불필요. 사용자가 유튜브 영상·채널·검색·자막 관련 질문을 하거나 /ytb를 입력했을 때 사용.
---

# ytb

YouTube를 직접 스크래핑하는 CLI 도구 `youtube`을 실행해 영상·채널 정보, 검색, 자막 데이터를 가져오는 스킬. API 키 없이 동작한다.

## CLI 위치

`youtube` 바이너리가 PATH에 설치되어 있어야 합니다.

모든 명령은 `youtube <subcommand>` 형태로 실행합니다.

## 영상 ID / URL 규칙

| 입력 형식 | 예시 |
|-----------|------|
| 영상 ID (11자) | `dQw4w9WgXcQ` |
| 전체 URL | `https://www.youtube.com/watch?v=dQw4w9WgXcQ` |
| 단축 URL | `https://youtu.be/dQw4w9WgXcQ` |

## 채널 입력 규칙

| 입력 형식 | 예시 |
|-----------|------|
| 핸들 | `@veritasium` 또는 `veritasium` |
| 채널 ID | `UCHnyfMqiRRG1u-2MsSQLbXA` |
| 전체 URL | `https://www.youtube.com/@veritasium` |

## 서브커맨드

### search — 키워드 영상 검색

```bash
youtube search "<검색어>" [-n 개수] [-f json|text]
```

- `-n`: 최대 결과 수, 기본 10
- 반환 필드: id, title, channel, duration, upload_date, view_count, url

```bash
youtube search "파이썬 튜토리얼" -n 5
youtube search "machine learning"
```

### video — 영상 기본 정보

```bash
youtube video <video> [-f json|text]
```

- 반환 필드: id, title, channel, channel_id, channel_url, duration, publish_date, upload_date, view_count, is_live, is_private, description, tags, thumbnail, url
- `is_live: true`인 영상은 자막 추출이 불가능할 수 있음
- `publish_date`는 절대 날짜(예: `2024. 3. 5.`), `upload_date`는 상대 표현(예: `1년 전`)

```bash
youtube video dQw4w9WgXcQ
youtube video https://youtu.be/dQw4w9WgXcQ
```

### channel — 채널 정보

```bash
youtube channel <channel> [-f json|text]
```

- 반환 필드: id, name, handle, description, subscribers, video_count, url

```bash
youtube channel @veritasium
youtube channel UCHnyfMqiRRG1u-2MsSQLbXA
```

### videos — 채널 영상 목록

```bash
youtube videos <channel> [-n 개수] [-f json|text]
```

- `-n`: 최대 영상 수, 기본 20
- 반환 필드: id, title, duration, upload_date, view_count, url

```bash
youtube videos @veritasium -n 10
youtube videos @MKBHD
```

### transcript — 영상 자막 추출

```bash
youtube transcript <video> [-l ko,en] [-t] [-f json|text]
```

- `-l`: 자막 언어 우선순위, 콤마 구분 (예: `-l ko,en`). 미지정 시 자동 선택
- `-t` 또는 `-f text`: 타임스탬프 없이 자막 본문(`text` 필드)만 출력 (둘이 동일한 결과)
- `-f json` (기본): 세그먼트(start/duration/text)와 전체 text를 함께 반환
- 수동 자막 우선, 없으면 자동 생성 자막(asr) 사용
- 반환 필드: video_id, url, lang, transcript(start/duration/text 리스트), text

```bash
youtube transcript dQw4w9WgXcQ
youtube transcript https://youtu.be/dQw4w9WgXcQ -l ko,en
youtube transcript dQw4w9WgXcQ -t
```

## 출력 형식

모든 커맨드에 `-f` 옵션 사용 가능:
- `-f json` (기본): JSON 출력 — 모든 필드 포함, 잘림 없음
- `-f text`: 사람이 읽기 좋은 텍스트 — `video`/`channel`은 description 300자, tags 10개로 잘림. `transcript`는 타임스탬프 없이 본문만 반환
- JSON에서 빈 값은 필드 자체가 누락될 수 있음 (omitempty)

LLM이 데이터를 받아 사용자에게 정리할 때는 `-f json`을 사용해 완전한 필드를 받는다.

## 사용 패턴

**"이 영상 정보 알려줘" (URL 또는 ID 제공)**
```bash
youtube video <video_id_or_url>
```

**"이 채널 정보 알려줘"**
```bash
youtube channel <handle_or_id>
```

**"이 영상 자막 뽑아줘"**
```bash
youtube transcript <video_id_or_url>
```

**"한국어 자막으로 추출해줘"**
```bash
youtube transcript <video_id> -l ko,en
```

**"자막 내용 요약해줘"** — `-t`로 타임스탬프 없이 받아 요약
```bash
youtube transcript <video_id> -t
```

**"veritasium 최근 영상 목록 보여줘"**
```bash
youtube videos @veritasium -n 10
```

**"파이썬 관련 유튜브 영상 찾아줘"**
```bash
youtube search "파이썬" -n 10
```

**자막 없는 영상 처리**
- 자막이 아예 없으면 `자막을 찾을 수 없습니다 (video: ...)` 오류 발생 → 사용자에게 자막 없는 영상임을 알린다.
- 특정 언어가 없을 때는 `요청한 언어의 자막을 찾을 수 없습니다` → `-l` 없이 재시도해 자동 선택된 자막을 받는다.
- 라이브 스트리밍 영상(`is_live: true`)은 자막이 제공되지 않을 수 있으므로 `video`로 먼저 확인하면 좋다.

## 주의사항

- API 없이 YouTube 내부 API 직접 호출 (구조 변경 시 동작 이상 가능)
- 일부 영상은 자막을 제공하지 않거나 자동 생성 자막만 있을 수 있습니다.
