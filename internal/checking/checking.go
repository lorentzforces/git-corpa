func CheckChanges() (CheckReport, error) {
	checkData, err := gatherState()
	if err != nil {
		return CheckReport{}, err
	}

	return reportChecks(checkData), nil
}

type CheckReport struct {
	Errors []CheckFlag
	Warnings []CheckFlag
}

type CheckFlag interface {
	Message() string
}

type StashEntryFlag struct {
	Number uint
}

func (flag StashEntryFlag) Message() string {
	return fmt.Sprintf("Stash entry {%d} has stashed changes from your current branch", flag.Number)
}

type LineIndentFlag struct {
	FileName string
	LineNumber uint
	FileIndents IndentKind
	LineIndents IndentKind
}

func (flag LineIndentFlag) Message() string {
	return fmt.Sprintf(
		"%s:%d | line has indentation (%s) inconsistent with the rest of the file (%s)",
		flag.FileName, flag.LineNumber, flag.FileIndents, flag.LineIndents,
	)
}

type KeywordPresenceFlag struct {
	FileName string
	LineNumber uint
	Keyword string
	LineContent string
}

func (flag KeywordPresenceFlag) Message() string {
	return fmt.Sprintf(
		"%s:%d | line contains keyword \"%s\": %s",
		flag.FileName, flag.LineNumber, flag.Keyword, flag.LineContent,
	)
}

type checkData struct {
	CurrentBranch string
	Files []diffFile
	StashEntries []stashEntry
}

type diffFile struct {
	FileName string
	Indents IndentKind
	ChangedLines []diffLine
}

type diffLine struct {
	LineNumber uint
	Indents IndentKind
	Content string
}

type IndentKind int
const (
	IndentUnknown IndentKind = iota
	IndentTab
	IndentSpace
	IndentMixedLine
)

func (ik IndentKind) String() string {
	switch ik {
		case IndentUnknown: return "IndentUnknown"
		case IndentTab: return "IndentTab"
		case IndentSpace: return "IndentSpace"
		case IndentMixedLine: return "IndentMixedLine"
	}
	platform.Assert(false, fmt.Sprintf("Invalid IndentKind value provided: %d", ik))
	panic("INVALID STATE: INVALID IndentKind VALUE PROVIDED")
}

func (ik IndentKind) incompatibleWithLine(lineIndent IndentKind) bool {
	if ik == IndentUnknown || lineIndent == IndentUnknown {
		return false
	}

	return ik != lineIndent
}

type stashEntry struct {
	Number uint
	Branch string
	RawString string
}

func gatherState() (checkData, error) {
		return checkData{}, err
	checkData := checkData{}
func parseStashEntries(rawEntries []string) ([]stashEntry, error) {
	entries := make([]stashEntry, 0, len(rawEntries))
func parseStashEntry(rawEntry string) (stashEntry, error) {
		return stashEntry{}, errors.Join(stashParseError, err)
	return stashEntry{

func parseDiffLines(rawLines []string) []diffFile {
	files := make(map[string]*diffFile, 0)
			files[fileName] = &diffFile{
				ChangedLines: make([]diffLine, 0),
		line := diffLine{}
	fileSlice := make([]diffFile, 0, len(files))
// do filesystem-related things with a diffFile
func populateFileInfo(diffFile *diffFile, file io.Reader) {
var errorKeywords = map[string]struct{} {
	"NOCHECKIN": struct{}{},
var warnKeywords = map[string]struct{} {
	"TODO": struct{}{},
func initKeywordRegex(wordSets... map[string]struct{}) *regexp.Regexp {
	var buf strings.Builder
	_, _ = buf.WriteString(`\b(`)

	isFirst := true
	for _, words := range wordSets {
		for word, _ := range words {
			if !isFirst { _, _ = buf.WriteString(`|`) }
			isFirst = false
			_, _ = buf.WriteString(word)
		}
	}

	_, _ = buf.WriteString(`)\b`)
	return regexp.MustCompile(buf.String())
var keywordRegex = initKeywordRegex(warnKeywords, errorKeywords)
func reportChecks(data checkData) CheckReport {
	result := CheckReport{
		Errors: make([]CheckFlag, 0),
		Warnings: make([]CheckFlag, 0),
	for _, entry := range data.StashEntries {
		if entry.Branch == data.CurrentBranch {
			result.Warnings = append(result.Warnings, StashEntryFlag{ Number: entry.Number })
		}
	}

	for _, file := range data.Files {
		for _, line := range file.ChangedLines {
			if file.Indents.incompatibleWithLine(line.Indents) {
				result.Errors = append(
					result.Errors,
					LineIndentFlag {
						FileName: file.FileName,
						LineNumber: line.LineNumber,
						FileIndents: file.Indents,
						LineIndents: line.Indents,
					},
				)
			}

			keyword := keywordRegex.FindString(line.Content)
			keywordFlag := KeywordPresenceFlag{
				FileName: file.FileName,
				LineNumber: line.LineNumber,
				Keyword: keyword,
				LineContent: line.Content,
			}
			if _, ok := errorKeywords[keyword]; ok {
				result.Errors = append(result.Errors, keywordFlag)
			}
			if _, ok := warnKeywords[keyword]; ok {
				result.Warnings = append(result.Errors, keywordFlag)
			}
		}
	}

	return result