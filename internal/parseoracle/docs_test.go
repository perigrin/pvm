// ABOUTME: Tests that the M3 gate's numbers in the plan and the perl-parser README are the committed ones.
// ABOUTME: Each number is read from its file -- sites.txt, m3start.txt, the faithful and clean ratchets, the opcount fixtures.

package parseoracle_test

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const (
	planPath         = "../../docs/plans/2026-09-05-parser-conformance-plan.md"
	parserREADMEPath = "../../docs/specs/perl-parser/README.md"
	oracleREADMEPath = "README.md"
	faithfulPath     = "../parse/testdata/perlgittfaithful.ratchet"
	compiledClean    = "../parse/testdata/compiledclean.txt"
	m3StartPath      = "testdata/ratchet/m3start.txt"
)

// m3Dirs are the perl.git t/ directories the M3 gate measures.
var m3Dirs = []string{"op", "mro", "uni", "lib", "test_pl"}

// siteTotals is sites.txt's `total <scope> <marker> <bucket> <count>` rows,
// summed per scope and bucket, with rv2hv's wider count apart.
type siteTotals struct {
	sites, exact, wider, noAnswer, wrong int
}

func readSiteTotals(t *testing.T) (all, m3 siteTotals, rv2hvWider int) {
	t.Helper()
	f, err := os.Open(corpusSitesPath())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	rows := 0
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 5 || fields[0] != "total" {
			continue
		}
		n, err := strconv.Atoi(fields[4])
		if err != nil {
			t.Fatalf("sites.txt: %q: %v", sc.Text(), err)
		}
		rows++
		var s *siteTotals
		switch fields[1] {
		case "all":
			s = &all
			if fields[2] == "rv2hv" && fields[3] == "wider" {
				rv2hvWider = n
			}
		case "m3":
			s = &m3
		default:
			t.Fatalf("sites.txt: unknown scope in %q", sc.Text())
		}
		s.sites += n
		switch fields[3] {
		case "exact":
			s.exact += n
		case "wider":
			s.wider += n
		case "no-answer":
			s.noAnswer += n
		case "WRONG":
			s.wrong += n
		default:
			t.Fatalf("sites.txt: unknown bucket in %q", sc.Text())
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if rows == 0 {
		t.Fatal("sites.txt holds no total rows")
	}
	return all, m3, rv2hvWider
}

// readNumber is the integer after `key = ` in path.
func readNumber(t *testing.T, path, key string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = ([0-9]+)$`).FindSubmatch(data)
	if m == nil {
		t.Fatalf("%s: no %q line", path, key)
	}
	n, _ := strconv.Atoi(string(m[1]))
	return n
}

// readHeader is the N and M of a `# N/M word` header line in path.
func readHeader(t *testing.T, path, word string) (int, int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`(?m)^# ([0-9]+)/([0-9]+) ` + word + `$`).FindSubmatch(data)
	if m == nil {
		t.Fatalf("%s: no %q header", path, "# N/M "+word)
	}
	n, _ := strconv.Atoi(string(m[1]))
	total, _ := strconv.Atoi(string(m[2]))
	return n, total
}

// m3Faithful is the faithful ratchet's M3-dir rows: how many hold 0, and
// how many there are.
func m3Faithful(t *testing.T) (faithful, total int) {
	t.Helper()
	f, err := os.Open(faithfulPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		dir, _, _ := strings.Cut(fields[1], "/")
		for _, d := range m3Dirs {
			if d == dir {
				total++
				if fields[0] == "0" {
					faithful++
				}
			}
		}
	}
	if total == 0 {
		t.Fatal("the faithful ratchet holds no M3-dir row")
	}
	return faithful, total
}

// planM3Section is the plan's `#### Measured at the M3 gate` subsection,
// up to the next heading.
func planM3Section(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, "\n#### Measured at the M3 gate\n")
	if start < 0 {
		t.Fatal("the plan has no `#### Measured at the M3 gate` subsection")
	}
	rest := text[start+1:]
	if end := regexp.MustCompile(`(?m)^#{2,4} `).FindAllStringIndex(rest, 2); len(end) == 2 {
		rest = rest[:end[1][0]]
	}
	return rest
}

// readmeM3Paragraph is the perl-parser README's paragraph starting `The
// M3 gate`, up to the blank line after it.
func readmeM3Paragraph(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(parserREADMEPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, "\nThe M3 gate")
	if start < 0 {
		t.Fatal("the perl-parser README has no paragraph starting `The M3 gate`")
	}
	rest := text[start+1:]
	if end := strings.Index(rest, "\n\n"); end >= 0 {
		rest = rest[:end]
	}
	return rest
}

// requireQuotes fails t for each want that doc, whitespace normalised so
// hard wrapping cannot fail it, does not contain.
func requireQuotes(t *testing.T, name, doc string, wants []string) {
	t.Helper()
	norm := strings.Join(strings.Fields(doc), " ")
	for _, want := range wants {
		if !strings.Contains(norm, want) {
			t.Errorf("%s does not state %q", name, want)
		}
	}
}

func pct(n, d int) string { return fmt.Sprintf("%.2f%%", 100*float64(n)/float64(d)) }

// TestPlanM3SectionQuotesSitesRatchet: the plan's measured subsection
// states the whole-corpus and M3-dir site totals and the rv2hv wider count
// exactly as sites.txt records them, so editing either side alone fails.
func TestPlanM3SectionQuotesSitesRatchet(t *testing.T) {
	all, m3, rv2hv := readSiteTotals(t)
	requireQuotes(t, "the plan's M3 subsection", planM3Section(t), []string{
		fmt.Sprintf("**%d/%d** M3-dir marker sites exact (%s)", m3.exact, m3.sites, pct(m3.exact, m3.sites)),
		fmt.Sprintf("%d wider, %d no-answer, %d WRONG", m3.wider, m3.noAnswer, m3.wrong),
		fmt.Sprintf("**%d/%d** whole-corpus marker sites exact (%s)", all.exact, all.sites, pct(all.exact, all.sites)),
		fmt.Sprintf("%d wider, %d no-answer, %d WRONG", all.wider, all.noAnswer, all.wrong),
		fmt.Sprintf("rv2hv wider **%d**", rv2hv),
	})
}

// TestM3DocsQuoteCommittedNumbers: the plan's measured subsection and the
// README's `The M3 gate` paragraph state the compiled count, the clean
// count, the faithful counts and the opcount assertion count, each equal
// to its file, and the parseoracle README's ceiling states m3start.txt's
// compiled count.
func TestM3DocsQuoteCommittedNumbers(t *testing.T) {
	compiled := readNumber(t, m3StartPath, "compiled")
	clean, cleanOf := readHeader(t, compiledClean, "clean")
	if cleanOf != compiled {
		t.Fatalf("compiledclean.txt counts %d compiled files, m3start.txt %d", cleanOf, compiled)
	}
	faithful, faithfulOf := readHeader(t, faithfulPath, "faithful")
	m3f, m3of := m3Faithful(t)
	total, _ := opcountPairs(mustLoadOpcountFixtures(t))
	wants := []string{
		fmt.Sprintf("**%d/%d** of the files perl compiles parse without error", clean, compiled),
		fmt.Sprintf("**%d/%d** M3-dir files re-emit faithfully (100%%)", m3f, m3of),
		fmt.Sprintf("perl.git t/ **%d/%d** faithful", faithful, faithfulOf),
		fmt.Sprintf("**%d** op=>count assertions", total),
	}
	if m3f != m3of {
		t.Errorf("re-emission is %d/%d on the M3 dirs, not 100%%", m3f, m3of)
	}
	requireQuotes(t, "the plan's M3 subsection", planM3Section(t), wants)
	requireQuotes(t, "the README's M3 gate paragraph", readmeM3Paragraph(t), wants)

	data, err := os.ReadFile(filepath.Clean(oracleREADMEPath))
	if err != nil {
		t.Fatal(err)
	}
	if want := fmt.Sprintf("**%d** of them are the compiled set", compiled); !strings.Contains(strings.Join(strings.Fields(string(data)), " "), want) {
		t.Errorf("internal/parseoracle/README.md's ceiling does not state %q", want)
	}
}
