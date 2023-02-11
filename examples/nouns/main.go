package main

import (
	"fmt"

	"github.com/ichiban/kagomelog"
	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"
)

func main() {
	// prologインタプリタを作成
	i := prolog.New(nil, nil)

	// 形態素解析の述語 analyze/3 を登録
	i.Register3(engine.NewAtom("analyze"), kagomelog.Analyze)

	// 名詞だけを取りだす述語 nouns/2 を定義
	if err := i.Exec(`
nouns([token(_, Noun, ['名詞'|_])|Tokens], [Noun|Nouns]) :- nouns(Tokens, Nouns), !.
nouns([_|Tokens], Nouns) :- nouns(Tokens, Nouns).
nouns([], []).
`); err != nil {
		panic(err)
	}

	// 与えられた文を形態素解析して名詞だけ取り出す
	sols, err := i.Query("analyze(?, normal, Tokens), nouns(Tokens, Nouns).", "すもももももももものうち")
	if err != nil {
		panic(err)
	}

	// Prologの変数と同名のフィールドを持つ構造体で受ける
	var s struct {
		Tokens []prolog.TermString
		Nouns  []string
	}
	for sols.Next() {
		if err := sols.Scan(&s); err != nil {
			panic(err)
		}
		fmt.Printf("Tokens: %s\n", s.Tokens)
		fmt.Printf("Nouns: %s\n", s.Nouns)
	}
}
