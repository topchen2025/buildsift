package analyzer

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Diagnosis struct {
	Found      bool
	Tool       string
	Summary    string
	Confidence string
	Evidence   []Evidence
	Cascades   int
	NextCheck  string
}

type Evidence struct {
	Line int
	Text string
}

type rule struct {
	id        string
	tool      string
	pattern   *regexp.Regexp
	score     int
	nextCheck string
	summary   func([]string, string) string
}

type candidate struct {
	rule    rule
	line    int
	text    string
	summary string
}

type sourceLine struct {
	number int
	text   string
}

var rules = []rule{
	{
		id:        "docker-daemon",
		tool:      "docker",
		pattern:   regexp.MustCompile(`(?i)cannot connect to the docker daemon[^\r\n]*`),
		score:     100,
		nextCheck: "docker info && docker context show",
		summary:   wholeMatch,
	},
	{
		id:        "docker-registry-timeout",
		tool:      "docker",
		pattern:   regexp.MustCompile(`(?i)(?:head|get) ["']?https?://[^\s"']+[^\r\n]*(?:context deadline exceeded|i/o timeout|connection refused)`),
		score:     100,
		nextCheck: "docker info && curl -I --connect-timeout 10 https://<registry>/v2/",
		summary:   wholeMatch,
	},
	{
		id:        "docker-port",
		tool:      "docker",
		pattern:   regexp.MustCompile(`(?i)(?:port is already allocated|bind[^\r\n]*address already in use)`),
		score:     100,
		nextCheck: "docker compose ps && lsof -nP -iTCP:<port> -sTCP:LISTEN",
		summary:   wholeMatch,
	},
	{
		id:        "docker-pull",
		tool:      "docker",
		pattern:   regexp.MustCompile(`(?i)(?:pull access denied|manifest unknown|no matching manifest)[^\r\n]*`),
		score:     96,
		nextCheck: "docker pull <image> && docker image inspect <image>",
		summary:   wholeMatch,
	},
	{
		id:        "docker-build",
		tool:      "docker",
		pattern:   regexp.MustCompile(`(?i)(?:error:\s*)?failed to solve:\s*(.+)`),
		score:     86,
		nextCheck: "docker build --progress=plain .",
		summary:   firstGroup,
	},
	{
		id:        "maven-symbol",
		tool:      "maven",
		pattern:   regexp.MustCompile(`(?i)(?:compilation failure[^:]*:\s*)?(?:cannot find symbol|package [^ ]+ does not exist)[^\r\n]*`),
		score:     99,
		nextCheck: "mvn -DskipTests compile -e",
		summary:   wholeMatch,
	},
	{
		id:        "maven-resolution",
		tool:      "maven",
		pattern:   regexp.MustCompile(`(?i)(?:could not resolve dependencies|non-resolvable parent POM|failed to read artifact descriptor)[^\r\n]*`),
		score:     98,
		nextCheck: "mvn dependency:tree -U -e",
		summary:   wholeMatch,
	},
	{
		id:        "maven-goal",
		tool:      "maven",
		pattern:   regexp.MustCompile(`(?i)failed to execute goal\s+([^\s]+)[^:]*:\s*(.+?)(?:\s*->\s*\[Help[^\]]*\])?$`),
		score:     78,
		nextCheck: "mvn -e -X",
		summary: func(match []string, _ string) string {
			return cleanSummary(fmt.Sprintf("%s: %s", match[1], match[2]))
		},
	},
	{
		id:        "gradle-cause",
		tool:      "gradle",
		pattern:   regexp.MustCompile(`^\s*>\s+(.+)$`),
		score:     92,
		nextCheck: "./gradlew <task> --stacktrace --info",
		summary:   firstGroup,
	},
	{
		id:        "gradle-task",
		tool:      "gradle",
		pattern:   regexp.MustCompile(`(?i)execution failed for task\s+['\"]?([^'\"]+)['\"]?\.?`),
		score:     76,
		nextCheck: "./gradlew <task> --stacktrace --info",
		summary: func(match []string, _ string) string {
			return cleanSummary("task " + match[1] + " failed")
		},
	},
	{
		id:        "npm-module",
		tool:      "npm",
		pattern:   regexp.MustCompile(`(?i)(?:module not found|cannot find module|could not resolve (?:module|package))[^\r\n]*`),
		score:     97,
		nextCheck: "npm ls && npm install",
		summary:   wholeMatch,
	},
	{
		id:        "pnpm-error",
		tool:      "npm",
		pattern:   regexp.MustCompile(`(?i)(ERR_PNPM_[A-Z_]+[^\r\n]*)`),
		score:     94,
		nextCheck: "pnpm install --reporter=append-only",
		summary:   firstGroup,
	},
	{
		id:        "npm-error",
		tool:      "npm",
		pattern:   regexp.MustCompile(`(?i)^\s*npm (?:ERR!|error)\s+(.+)$`),
		score:     68,
		nextCheck: "npm doctor && npm run <script> -- --verbose",
		summary:   firstGroup,
	},
	{
		id:        "caused-by",
		tool:      "",
		pattern:   regexp.MustCompile(`(?i)caused by:\s*(.+)$`),
		score:     90,
		nextCheck: "",
		summary:   firstGroup,
	},
	{
		id:        "actionable-generic",
		tool:      "",
		pattern:   regexp.MustCompile(`(?i)(?:nosuchfileexception|filenotfoundexception|no such file or directory|file not found|cannot find symbol|does not exist|context deadline exceeded|connection refused|address already in use)[^\r\n]*`),
		score:     95,
		nextCheck: "",
		summary:   wholeMatch,
	},
	{
		id:        "generic-error",
		tool:      "",
		pattern:   regexp.MustCompile(`(?i)(?:^|\s)(?:error|fatal):\s*(.+)$`),
		score:     58,
		nextCheck: "",
		summary:   firstGroup,
	},
}

var (
	ansiPattern       = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
	ghaPrefixPattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z\s+`)
	secretPattern     = regexp.MustCompile(`(?i)\b(token|password|passwd|secret|api[_-]?key)\b(\s*[:=]\s*)("[^"]*"|'[^']*'|[^\s,;]+)`)
	bearerPattern     = regexp.MustCompile(`(?i)(authorization:\s*bearer\s+)[^\s]+`)
	knownTokenPattern = regexp.MustCompile(`\b(?:gh[pousr]_[A-Za-z0-9_]{20,}|github_pat_[A-Za-z0-9_]{20,}|sk-[A-Za-z0-9_-]{16,})\b`)
	urlAuthPattern    = regexp.MustCompile(`(https?://)[^/@\s:]+:[^/@\s]+@`)
	unixHomePattern   = regexp.MustCompile(`/(?:Users|home)/[^/\s]+/`)
	windowsHome       = regexp.MustCompile(`(?i)[A-Z]:\\Users\\[^\\\s]+\\`)
)

// 2026-07-18：规则优先选择可操作的底层错误，避免把级联失败当成根因。
func Analyze(raw string) Diagnosis {
	lines := splitLines(raw)
	dominantTool := detectTool(lines)
	candidates := findCandidates(lines, dominantTool)
	if len(candidates) == 0 {
		return Diagnosis{
			Found:      false,
			Tool:       dominantTool,
			Confidence: "low",
			NextCheck:  defaultCheck(dominantTool),
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].rule.score != candidates[j].rule.score {
			return candidates[i].rule.score > candidates[j].rule.score
		}
		return candidates[i].line < candidates[j].line
	})

	chosen := candidates[0]
	tool := chosen.rule.tool
	if tool == "" {
		tool = dominantTool
	}
	evidence := selectEvidence(chosen, candidates, tool)
	nextCheck := chosen.rule.nextCheck
	if nextCheck == "" {
		nextCheck = defaultCheck(tool)
	}

	return Diagnosis{
		Found:      true,
		Tool:       tool,
		Summary:    chosen.summary,
		Confidence: confidence(chosen.rule.score),
		Evidence:   evidence,
		Cascades:   max(0, len(candidates)-len(evidence)),
		NextCheck:  nextCheck,
	}
}

func splitLines(raw string) []sourceLine {
	parts := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	lines := make([]sourceLine, 0, len(parts))
	for i, part := range parts {
		part = ansiPattern.ReplaceAllString(part, "")
		part = ghaPrefixPattern.ReplaceAllString(part, "")
		part = Redact(part)
		if strings.TrimSpace(part) == "" {
			continue
		}
		lines = append(lines, sourceLine{number: i + 1, text: part})
	}
	return lines
}

func detectTool(lines []sourceLine) string {
	scores := map[string]int{"maven": 0, "gradle": 0, "npm": 0, "docker": 0}
	for _, line := range lines {
		lower := strings.ToLower(line.text)
		if strings.Contains(lower, "failed to execute goal") || strings.Contains(lower, "reactor summary") || strings.Contains(lower, "maven-") {
			scores["maven"] += 3
		}
		if strings.Contains(lower, "execution failed for task") || strings.Contains(lower, "failure: build failed with an exception") || strings.Contains(lower, "> task :") {
			scores["gradle"] += 3
		}
		if strings.Contains(lower, "npm err!") || strings.Contains(lower, "npm error") || strings.Contains(lower, "err_pnpm_") || strings.Contains(lower, "elifecycle") {
			scores["npm"] += 3
		}
		if strings.Contains(lower, "failed to solve") || strings.Contains(lower, "docker daemon") || strings.Contains(lower, "pull access denied") || strings.Contains(lower, "manifest unknown") {
			scores["docker"] += 3
		}
	}

	tool := "unknown"
	best := 0
	for _, name := range []string{"maven", "gradle", "npm", "docker"} {
		if scores[name] > best {
			best = scores[name]
			tool = name
		}
	}
	return tool
}

func findCandidates(lines []sourceLine, dominantTool string) []candidate {
	var candidates []candidate
	for _, line := range lines {
		for _, currentRule := range rules {
			match := currentRule.pattern.FindStringSubmatch(line.text)
			if match == nil {
				continue
			}
			if currentRule.id == "gradle-cause" && dominantTool != "gradle" {
				continue
			}
			if currentRule.id == "gradle-cause" && strings.HasPrefix(strings.ToLower(strings.TrimSpace(match[1])), "task ") {
				continue
			}
			candidates = append(candidates, candidate{
				rule:    currentRule,
				line:    line.number,
				text:    line.text,
				summary: currentRule.summary(match, line.text),
			})
		}
	}
	return deduplicateCandidates(candidates)
}

func deduplicateCandidates(candidates []candidate) []candidate {
	bestByLine := make(map[int]candidate)
	for _, item := range candidates {
		current, exists := bestByLine[item.line]
		if !exists || item.rule.score > current.rule.score {
			bestByLine[item.line] = item
		}
	}
	lineNumbers := make([]int, 0, len(bestByLine))
	for line := range bestByLine {
		lineNumbers = append(lineNumbers, line)
	}
	sort.Ints(lineNumbers)
	result := make([]candidate, 0, len(lineNumbers))
	for _, line := range lineNumbers {
		result = append(result, bestByLine[line])
	}
	return result
}

func selectEvidence(chosen candidate, candidates []candidate, tool string) []Evidence {
	result := []Evidence{{Line: chosen.line, Text: chosen.text}}
	seenLines := map[int]bool{chosen.line: true}
	others := append([]candidate(nil), candidates...)
	sort.SliceStable(others, func(i, j int) bool {
		distanceI := abs(others[i].line - chosen.line)
		distanceJ := abs(others[j].line - chosen.line)
		if distanceI != distanceJ {
			return distanceI < distanceJ
		}
		return others[i].rule.score > others[j].rule.score
	})
	for _, item := range others {
		itemTool := item.rule.tool
		if itemTool == "" {
			itemTool = tool
		}
		if seenLines[item.line] || itemTool != tool {
			continue
		}
		result = append(result, Evidence{Line: item.line, Text: item.text})
		seenLines[item.line] = true
		if len(result) == 3 {
			break
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Line < result[j].Line })
	return result
}

func Redact(input string) string {
	input = bearerPattern.ReplaceAllString(input, "${1}[REDACTED]")
	input = secretPattern.ReplaceAllString(input, "${1}${2}[REDACTED]")
	input = knownTokenPattern.ReplaceAllString(input, "[REDACTED]")
	input = urlAuthPattern.ReplaceAllString(input, "${1}[REDACTED]@")
	input = unixHomePattern.ReplaceAllString(input, "~/")
	input = windowsHome.ReplaceAllString(input, `~\`)
	return input
}

func confidence(score int) string {
	if score >= 90 {
		return "high"
	}
	if score >= 70 {
		return "medium"
	}
	return "low"
}

func defaultCheck(tool string) string {
	switch tool {
	case "maven":
		return "mvn -e -X"
	case "gradle":
		return "./gradlew <task> --stacktrace --info"
	case "npm":
		return "npm doctor && npm run <script> -- --verbose"
	case "docker":
		return "docker info && docker compose config"
	default:
		return "rerun the failing command with verbose or stacktrace output"
	}
}

func wholeMatch(match []string, _ string) string {
	return cleanSummary(match[0])
}

func firstGroup(match []string, text string) string {
	if len(match) > 1 && strings.TrimSpace(match[1]) != "" {
		return cleanSummary(match[1])
	}
	return cleanSummary(text)
}

func cleanSummary(input string) string {
	input = strings.TrimSpace(input)
	input = regexp.MustCompile(`^(?:\[[A-Z]+\]|[A-Z]+[:!])\s*`).ReplaceAllString(input, "")
	if len(input) > 220 {
		return input[:217] + "..."
	}
	return input
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
