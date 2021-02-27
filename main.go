package main

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/ichiban/prolog"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func main() {
	logrus.SetLevel(logrus.WarnLevel) // ちょっとログがうるさいので黙らせる

	// prologエンジンを作成
	e, err := prolog.NewEngine(nil, nil)
	if err != nil {
		panic(err)
	}

	// 形態素解析の述語 analyze/3 を登録
	e.Register3("analyze", Analyze)

	// 名詞だけを取りだす述語 nouns/2 を定義
	if err := e.Exec(`
nouns([token(_, Noun, ['名詞'|_])|Tokens], [Noun|Nouns]) :- nouns(Tokens, Nouns), !.
nouns([_|Tokens], Nouns) :- nouns(Tokens, Nouns).
nouns([], []).
`); err != nil {
		panic(err)
	}

	// 与えられた文を形態素解析して名詞だけ取り出す
	if _, err := e.Query("analyze('すもももももももものうち', normal, Tokens), nouns(Tokens, Nouns).", func(vars []*prolog.Variable) bool {
		for _, v := range vars {
			fmt.Printf("%s\n", e.Describe(v))
		}
		return true
	}); err != nil {
		panic(err)
	}
}

// analyze/3の定義
func Analyze(input, mode, tokens prolog.Term, k func() (bool, error)) (bool, error) {
	// 最初の引数 input が（変数かもしれないので、それをたどっていった先で）アトム（文字列みたいなもの）であるか確認
	i, ok := prolog.Resolve(input).(prolog.Atom)
	if !ok {
		return false, errors.New("not an atom")
	}

	// 形態素解析のモードを与えられたアトムから特定
	var tm tokenizer.TokenizeMode
	m, ok := prolog.Resolve(mode).(prolog.Atom)
	if !ok {
		return false, errors.New("not an atom")
	}
	switch m {
	case "normal":
		tm = tokenizer.Normal
	case "search":
		tm = tokenizer.Search
	case "extended":
		tm = tokenizer.Extended
	default:
		return false, errors.New("unknown mode")
	}

	// 形態素解析器を準備
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return false, err
	}

	// 形態素解析の結果得られたトークンを token(ID, Surface, [Feature, ...]) の形式の複合項に変換
	var ret []prolog.Term
	for _, t := range t.Analyze(string(i), tm) {
		features := t.Features()
		fs := make([]prolog.Term, len(features))
		for i, f := range features {
			fs[i] = prolog.Atom(f)
		}

		ret = append(ret, &prolog.Compound{
			Functor: "token",
			Args: []prolog.Term{
				prolog.Integer(t.ID),
				prolog.Atom(t.Surface),
				prolog.List(fs...),
			},
		})
	}

	// トークンを表す複合項の入ったスライスをリストに変換して出力用の引数と単一化
	// 継続 k はそのまま渡した先で処理してもらう
	return prolog.Unify(tokens, prolog.List(ret...), k)
}
