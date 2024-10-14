// Package cmdsim provides functionality for finding the best matching command
// based on the Jaro-Winkler distance algorithm.
package stringutil

import (
	"iter"
	"slices"
	"unicode/utf8"
)

// DefaultSimilarityThreshold is the default similarity threshold for the Jaro-Winkler distance.
const DefaultSimilarityThreshold = 0.7

// JaroWinklerDistance calculates the Jaro-Winkler distance between two strings.
// The result is a value between 0 and 1, where 1 indicates a perfect match.
// This function is case-sensitive.
func JaroWinklerDistance(s1, s2 string) float64 {
	// If strings are identical, return 1
	if s1 == s2 {
		return 1
	}

	// Get string lengths
	len1 := utf8.RuneCountInString(s1)
	len2 := utf8.RuneCountInString(s2)

	// Calculate match window
	maxDist := max(len1, len2)/2 - 1

	// Initialize match and transposition counts
	matches := 0
	transpositions := 0
	matched1 := make([]bool, len1)
	matched2 := make([]bool, len2)

	// Count matching characters
	for i, r1 := range s1 {
		start := max(0, i-maxDist)
		end := min(i+maxDist+1, len2)
		for j := start; j < end; j++ {
			if !matched2[j] && r1 == rune(s2[j]) {
				matched1[i] = true
				matched2[j] = true
				matches++
				break
			}
		}
	}

	// If no matches, return 0
	if matches == 0 {
		return 0
	}

	// Count transpositions
	j := 0
	for i, r1 := range s1 {
		if matched1[i] {
			for !matched2[j] {
				j++
			}
			if r1 != rune(s2[j]) {
				transpositions++
			}
			j++
		}
	}

	// Calculate Jaro distance
	jaro := (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0

	// Calculate common prefix length (up to 4 characters)
	prefixLen := 0
	for i := 0; i < min(min(len1, len2), 4); i++ {
		if s1[i] == s2[i] {
			prefixLen++
		} else {
			break
		}
	}

	// Calculate and return Jaro-Winkler distance
	return jaro + float64(prefixLen)*0.1*(1-jaro)
}

func identity[T any](_ T, s string) string {
	return s
}

// FindBestMatch finds the best matching string to the input
func FindBestMatch(source []string, input string, threshold float64) string {
	return FindBestMatchFunc(slices.All(source), input, threshold, identity)
}

// FindBestMatchFunc finds the best matching value and its similarity score
// from the given list of values. It returns the best match and its similarity score.
// If no command meets the threshold, it returns an zero key and zero similarity score.
func FindBestMatchFunc[F ~func(K, V) string, K, V any](seq iter.Seq2[K, V], input string, threshold float64, fn F) string {
	var bestMatch string
	var highestSimilarity float64

	for k, v := range seq {
		s := fn(k, v)
		sim := JaroWinklerDistance(input, s)
		if sim > highestSimilarity {
			highestSimilarity = sim
			bestMatch = s
		}
	}

	if highestSimilarity >= threshold {
		return bestMatch
	}
	return ""
}

// IsASCIIIdentifier checks if a string is a valid ascii identifier.
func IsASCIIIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_') {
				return false
			}
		} else {
			if !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9' || r == '_') {
				return false
			}
		}
	}
	return true
}
