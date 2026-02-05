package openmetrics

import (
	"bytes"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

func Parse(raw []byte) (MetricFamilySet, error) {
	lines := bytes.Lines(raw)

	sets := make(MetricFamilySet)

	for lineRaw := range lines {
		line := strings.TrimSpace(string(lineRaw))
		if line == "" || line == "# EOF" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}

			action := parts[1] // HELP or TYPE
			name := parts[2]

			if sets[name] == nil {
				sets[name] = &MetricFamily{Name: name}
			}

			if action == "HELP" {
				sets[name].Description = strings.Join(parts[3:], " ")
			} else if action == "TYPE" {
				sets[name].Type = parts[3]
			}

			continue
		}

		metric, baseName := parseMetric(line)

		if metric != nil {
			if sets[baseName] == nil {
				sets[baseName] = &MetricFamily{Name: baseName}
			}
			sets[baseName].Metrics = append(sets[baseName].Metrics, metric)
		}
	}

	return sets, nil
}

func parseMetric(line string) (*Metric, string) {
	spaceStart := strings.LastIndex(line, " ")
	if spaceStart == -1 {
		return nil, ""
	}

	m := &Metric{
		Value:  line[spaceStart+1:],
		Labels: make(map[string]string),
	}

	bracketStart := strings.Index(line, "{")

	var fullName string
	if bracketStart != -1 && bracketStart < spaceStart {
		fullName = line[:bracketStart]
		labelPart := line[bracketStart+1 : spaceStart-1]

		s := &scanner.Scanner{}
		s.Init(bytes.NewBufferString(labelPart))
		s.IsIdentRune = func(ch rune, i int) bool {
			return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
		}

		kv := [2]string{}
		off := 0

		reset := func() {
			off = 0
		}

		commitIfNeed := func() {
			if off == 1 {
				k := kv[0]
				v := kv[1]
				if len(v) > 0 {
					if v[0] == '"' {
						v, _ = strconv.Unquote(v)
					}
				}

				m.Labels[k] = v

				reset()
			}
		}

		for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
			switch tok {
			case ',':
				commitIfNeed()
			case '=':
				off++
			default:
				kv[off] = s.TokenText()
			}
		}

		commitIfNeed()
	} else {
		fullName = line[:spaceStart]
	}

	m.Name = fullName

	baseName := fullName
	for _, suffix := range []string{"_bucket", "_sum", "_count", "_min", "_max"} {
		if strings.HasSuffix(fullName, suffix) {
			baseName = strings.TrimSuffix(fullName, suffix)
			break
		}
	}

	return m, baseName
}
