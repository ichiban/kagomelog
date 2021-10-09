package main

import (
	"errors"
	"fmt"

	"github.com/ichiban/prolog"
	"github.com/ichiban/prolog/engine"
	"github.com/ichiban/prolog/nondet"
	"github.com/ichiban/prolog/term"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func main() {
	// prologインタプリタを作成
	i := prolog.New(nil, nil)

	// 形態素解析の述語 analyze/3 を登録
	i.Register3("analyze", Analyze)

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
		Tokens []term.Interface
		Nouns  []string
	}
	for sols.Next() {
		if err := sols.Scan(&s); err != nil {
			panic(err)
		}
		fmt.Printf("Tokens: %+v\n", s.Tokens)
		fmt.Printf("Nouns: %s\n", s.Nouns)
	}
}

// analyze/3の定義
func Analyze(input, mode, tokens term.Interface, k func(*term.Env) *nondet.Promise, env *term.Env) *nondet.Promise {
	// 最初の引数 input が（変数かもしれないので、それをたどっていった先で）アトム（文字列みたいなもの）であるか確認
	i, ok := env.Resolve(input).(term.Atom)
	if !ok {
		return nondet.Error(errors.New("not an atom"))
	}

	// 形態素解析のモードを与えられたアトムから特定
	var tm tokenizer.TokenizeMode
	m, ok := env.Resolve(mode).(term.Atom)
	if !ok {
		return nondet.Error(errors.New("not an atom"))
	}
	switch m {
	case "normal":
		tm = tokenizer.Normal
	case "search":
		tm = tokenizer.Search
	case "extended":
		tm = tokenizer.Extended
	default:
		return nondet.Error(errors.New("unknown mode"))
	}

	// 形態素解析器を準備
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nondet.Error(err)
	}

	// 形態素解析の結果得られたトークンを token(ID, Surface, [Feature, ...]) の形式の複合項に変換
	var ret []term.Interface
	for _, t := range t.Analyze(string(i), tm) {
		features := t.Features()
		fs := make([]term.Interface, len(features))
		for i, f := range features {
			fs[i] = term.Atom(f)
		}

		ret = append(ret, &term.Compound{
			Functor: "token",
			Args: []term.Interface{
				term.Integer(t.ID),
				term.Atom(t.Surface),
				term.List(fs...),
			},
		})
	}

	// トークンを表す複合項の入ったスライスをリストに変換して出力用の引数と単一化
	// 継続 k はそのまま渡した先で処理してもらう
	return engine.Unify(tokens, term.List(ret...), k, env)
}
