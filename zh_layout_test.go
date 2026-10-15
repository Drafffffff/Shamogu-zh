package main

import (
	"codeberg.org/anaseto/gruid/ui"
	"strings"
	"testing"
)

func TestChineseClosingPunctuationWrap(t *testing.T) {
	for _, tail := range []string{"汉。", "汉。”", "汉@r。@N"} {
		text := strings.Repeat("界", 45) + tail
		wrapped := ui.Text(text).WithMarkups(Markups).Format(46)
		if wrapped.Size().X > 46 {
			t.Fatalf("line exceeds reading width: %v", wrapped.Size())
		}
		if strings.ReplaceAll(wrapped.Text(), "\n", "") != text {
			t.Fatal("wrapping changed text or markup")
		}
		for _, line := range strings.Split(wrapped.Text(), "\n") {
			if strings.HasPrefix(line, "。") || strings.HasPrefix(line, "”") {
				t.Fatal("closing punctuation begins a line")
			}
		}
	}
}

func TestChineseDynamicNouns(t *testing.T) {
	for _, n := range []int{0, 1, 2, 12} {
		if Countable("香蕉", n) != "香蕉" || strings.Contains(CountNoun("香蕉", n), "s") {
			t.Fatal("English plural leaked into Chinese noun")
		}
	}
	if One("香蕉") != "香蕉" {
		t.Fatal("English article leaked into Chinese noun")
	}
}
