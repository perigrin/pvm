// ABOUTME: Reads and writes the committed oracle facts (testdata/oracle/): perl's per-file parse facts, produced once by the pinned perl.
// ABOUTME: The facts are the fields parseoracle.CompareFacts reads; nothing here runs perl.

package chart

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"tamarou.com/pvm/internal/parseoracle"
)

// OracleRecord is one file's committed oracle facts. A file perl produced no
// facts for (a runner timeout) has Missing set and zero Facts, so it reads as
// not accepted.
type OracleRecord struct {
	Key     string
	Missing string
	Facts   parseoracle.Facts
}

// OracleFactsFile is an oracle-facts file: its provenance header and records.
type OracleFactsFile struct {
	Provenance map[string]string
	Records    []OracleRecord
}

// oracleWire is the committed form of the facts CompareFacts reads. The op
// sequence is not kept: only whether an op was built.
type oracleWire struct {
	OK         bool              `json:"ok"`
	Walked     bool              `json:"walked"`
	OpCount    int               `json:"op_count"`
	Sites      map[string][]int  `json:"sites"`
	Prototypes map[string]string `json:"prototypes"`
}

// FormatOracleRecord renders one record line: "key<TAB>facts<TAB>json", or
// "key<TAB>none<TAB>reason" for a file perl produced no facts for.
func FormatOracleRecord(r OracleRecord) (string, error) {
	if r.Missing != "" {
		return r.Key + "\tnone\t" + strings.Join(strings.Fields(r.Missing), " "), nil
	}
	w := oracleWire{OK: r.Facts.OK, Walked: r.Facts.Walked, OpCount: r.Facts.OpCount,
		Sites: map[string][]int{}, Prototypes: r.Facts.Prototypes}
	if w.Prototypes == nil {
		w.Prototypes = map[string]string{}
	}
	for m, lines := range r.Facts.Sites {
		w.Sites[string(m)] = lines
	}
	b, err := json.Marshal(w)
	return r.Key + "\tfacts\t" + string(b), err
}

// ReadOracleFacts reads an oracle-facts file: "# key: value" provenance lines,
// then one record per line. A truncated or malformed record is an error.
func ReadOracleFacts(path string) (*OracleFactsFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	of := &OracleFactsFile{Provenance: map[string]string{}}
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<24)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		if rest, ok := strings.CutPrefix(line, "# "); ok {
			k, v, _ := strings.Cut(rest, ": ")
			of.Provenance[k] = v
			continue
		}
		r, err := parseOracleRecord(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, n, err)
		}
		of.Records = append(of.Records, r)
	}
	return of, sc.Err()
}

func parseOracleRecord(line string) (OracleRecord, error) {
	c := strings.Split(line, "\t")
	if len(c) != 3 || c[0] == "" || c[2] == "" {
		return OracleRecord{}, fmt.Errorf("malformed record %q", line)
	}
	switch c[1] {
	case "none":
		return OracleRecord{Key: c[0], Missing: c[2]}, nil
	case "facts":
		var w oracleWire
		dec := json.NewDecoder(bytes.NewReader([]byte(c[2])))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&w); err != nil || dec.More() {
			return OracleRecord{}, fmt.Errorf("%s: malformed facts %q (%v)", c[0], c[2], err)
		}
		sites := make(map[parseoracle.Marker][]int, len(w.Sites))
		for m, lines := range w.Sites {
			sites[parseoracle.Marker(m)] = lines
		}
		return OracleRecord{Key: c[0], Facts: parseoracle.Facts{OK: w.OK, Walked: w.Walked,
			OpCount: w.OpCount, Sites: sites, Prototypes: w.Prototypes}}, nil
	}
	return OracleRecord{}, fmt.Errorf("%s: record kind %q is neither facts nor none", c[0], c[1])
}
