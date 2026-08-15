package ownership

import (
	"bytes"
	"errors"
	"sort"
	"unicode/utf8"
)

const (
	mechanicalMergeAlgorithm = "mbc-diff3-lines-v1"
	maxMechanicalMergeBytes  = 1 << 20
	maxMechanicalMergeCells  = 4_000_000
)

type lineHunk struct {
	start       int
	end         int
	replacement [][]byte
}

func mechanicalMerge(base, local, upstream []byte, baseMode, localMode, upstreamMode string) ([]byte, string, error) {
	for _, data := range [][]byte{base, local, upstream} {
		if len(data) > maxMechanicalMergeBytes {
			return nil, "", errors.New("merge input exceeds 1 MiB limit")
		}
		if bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
			return nil, "", errors.New("merge input is not NUL-free UTF-8 text")
		}
	}
	baseLines, localLines, upstreamLines := splitExactLines(base), splitExactLines(local), splitExactLines(upstream)
	if len(baseLines)*len(localLines) > maxMechanicalMergeCells || len(baseLines)*len(upstreamLines) > maxMechanicalMergeCells {
		return nil, "", errors.New("merge line comparison exceeds work limit")
	}
	localHunks := diffLineHunks(baseLines, localLines)
	upstreamHunks := diffLineHunks(baseLines, upstreamLines)
	hunks, err := combineLineHunks(localHunks, upstreamHunks)
	if err != nil {
		return nil, "", err
	}
	mode, err := mergeMode(baseMode, localMode, upstreamMode)
	if err != nil {
		return nil, "", err
	}
	var result bytes.Buffer
	cursor := 0
	for _, hunk := range hunks {
		for _, line := range baseLines[cursor:hunk.start] {
			result.Write(line)
		}
		for _, line := range hunk.replacement {
			result.Write(line)
		}
		if hunk.end > cursor {
			cursor = hunk.end
		}
	}
	for _, line := range baseLines[cursor:] {
		result.Write(line)
	}
	return result.Bytes(), mode, nil
}

func splitExactLines(data []byte) [][]byte {
	if len(data) == 0 {
		return nil
	}
	parts := bytes.SplitAfter(data, []byte{'\n'})
	if len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}
	result := make([][]byte, len(parts))
	for index := range parts {
		result[index] = append([]byte(nil), parts[index]...)
	}
	return result
}

func diffLineHunks(base, side [][]byte) []lineHunk {
	rows, columns := len(base)+1, len(side)+1
	dp := make([]int, rows*columns)
	cell := func(i, j int) int { return i*columns + j }
	for i := len(base) - 1; i >= 0; i-- {
		for j := len(side) - 1; j >= 0; j-- {
			if bytes.Equal(base[i], side[j]) {
				dp[cell(i, j)] = 1 + dp[cell(i+1, j+1)]
			} else if dp[cell(i, j+1)] >= dp[cell(i+1, j)] {
				dp[cell(i, j)] = dp[cell(i, j+1)]
			} else {
				dp[cell(i, j)] = dp[cell(i+1, j)]
			}
		}
	}
	type match struct{ base, side int }
	matches := make([]match, 0, dp[0])
	for i, j := 0, 0; i < len(base) && j < len(side); {
		if bytes.Equal(base[i], side[j]) && dp[cell(i, j)] == 1+dp[cell(i+1, j+1)] {
			matches = append(matches, match{i, j})
			i, j = i+1, j+1
		} else if dp[cell(i, j+1)] == dp[cell(i, j)] {
			j++
		} else {
			i++
		}
	}
	hunks := make([]lineHunk, 0)
	baseStart, sideStart := 0, 0
	for _, current := range append(matches, match{len(base), len(side)}) {
		if baseStart != current.base || sideStart != current.side {
			replacement := make([][]byte, current.side-sideStart)
			for index := range replacement {
				replacement[index] = append([]byte(nil), side[sideStart+index]...)
			}
			hunks = append(hunks, lineHunk{start: baseStart, end: current.base, replacement: replacement})
		}
		baseStart, sideStart = current.base+1, current.side+1
	}
	return hunks
}

func combineLineHunks(left, right []lineHunk) ([]lineHunk, error) {
	combined := append(append([]lineHunk(nil), left...), right...)
	for i := 0; i < len(left); i++ {
		for j := 0; j < len(right); j++ {
			if identicalLineHunk(left[i], right[j]) {
				continue
			}
			if lineHunksCollide(left[i], right[j]) {
				return nil, errors.New("local and upstream line changes overlap")
			}
		}
	}
	unique := make([]lineHunk, 0, len(combined))
	for _, candidate := range combined {
		duplicate := false
		for _, existing := range unique {
			if identicalLineHunk(candidate, existing) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			unique = append(unique, candidate)
		}
	}
	sort.SliceStable(unique, func(i, j int) bool {
		if unique[i].start != unique[j].start {
			return unique[i].start < unique[j].start
		}
		return unique[i].start == unique[i].end && unique[j].start != unique[j].end
	})
	return unique, nil
}

func identicalLineHunk(left, right lineHunk) bool {
	if left.start != right.start || left.end != right.end || len(left.replacement) != len(right.replacement) {
		return false
	}
	for index := range left.replacement {
		if !bytes.Equal(left.replacement[index], right.replacement[index]) {
			return false
		}
	}
	return true
}

func lineHunksCollide(left, right lineHunk) bool {
	leftInsertion, rightInsertion := left.start == left.end, right.start == right.end
	switch {
	case leftInsertion && rightInsertion:
		return left.start == right.start
	case leftInsertion:
		return right.start < left.start && left.start < right.end
	case rightInsertion:
		return left.start < right.start && right.start < left.end
	default:
		return left.start < right.end && right.start < left.end
	}
}

func mergeMode(base, local, upstream string) (string, error) {
	if local == upstream {
		return local, nil
	}
	if local == base {
		return upstream, nil
	}
	if upstream == base {
		return local, nil
	}
	return "", errors.New("local and upstream executable-mode changes conflict")
}
