package kagomelog

import (
	"github.com/ichiban/prolog/engine"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

var (
	atomAtom     = engine.NewAtom("atom")
	atomMode     = engine.NewAtom("mode")
	atomNormal   = engine.NewAtom("normal")
	atomSearch   = engine.NewAtom("search")
	atomExtended = engine.NewAtom("extended")
	atomCompound = engine.NewAtom("compound")
	atomToken    = engine.NewAtom("token")
	atomInteger  = engine.NewAtom("integer")
)

// Analyze は形態素解析を行う
func Analyze(vm *engine.VM, input, mode, tokens engine.Term, k engine.Cont, env *engine.Env) *engine.Promise {
	// 最初の引数 input が（変数かもしれないので、それをたどっていった先で）アトムであるか確認
	var i engine.Atom
	switch input := env.Resolve(input).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case engine.Atom:
		i = input
	default:
		return engine.Error(engine.TypeError(atomAtom, input, env))
	}

	// 形態素解析のモードを与えられたアトムから特定
	var tm tokenizer.TokenizeMode
	switch mode := env.Resolve(mode).(type) {
	case engine.Variable:
		return engine.Error(engine.InstantiationError(env))
	case engine.Atom:
		switch mode {
		case atomNormal:
			tm = tokenizer.Normal
		case atomSearch:
			tm = tokenizer.Search
		case atomExtended:
			tm = tokenizer.Extended
		default:
			return engine.Error(engine.DomainError(atomMode, mode, env))
		}
	default:
		return engine.Error(engine.TypeError(atomAtom, mode, env))
	}

	// 出力用の引数 tokens が出力にマッチしうるかチェック
	iter := engine.ListIterator{List: tokens, Env: env, AllowPartial: true}
	for iter.Next() {
		switch token := env.Resolve(iter.Current()).(type) {
		case engine.Variable:
			break
		case engine.Compound:
			if token.Functor() != atomToken || token.Arity() != 3 {
				return engine.Error(engine.DomainError(atomToken, token, env))
			}

			switch id := env.Resolve(token.Arg(0)).(type) {
			case engine.Variable, engine.Integer:
				break
			default:
				return engine.Error(engine.TypeError(atomInteger, id, env))
			}

			switch surface := env.Resolve(token.Arg(1)).(type) {
			case engine.Variable, engine.Atom:
				break
			default:
				return engine.Error(engine.TypeError(atomAtom, surface, env))
			}

			iter := engine.ListIterator{List: token.Arg(2), Env: env, AllowPartial: true}
			for iter.Next() {
				switch feature := env.Resolve(iter.Current()).(type) {
				case engine.Variable, engine.Atom:
					break
				default:
					return engine.Error(engine.TypeError(atomAtom, feature, env))
				}
			}
			if err := iter.Err(); err != nil {
				return engine.Error(err)
			}
		default:
			return engine.Error(engine.TypeError(atomCompound, token, env))
		}
	}
	if err := iter.Err(); err != nil {
		return engine.Error(err)
	}

	// 形態素解析器を準備
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return engine.Error(err)
	}

	// 形態素解析の結果得られたトークンを token(ID, Surface, [Feature, ...]) の形式の複合項に変換
	var ret []engine.Term
	for _, t := range t.Analyze(i.String(), tm) {
		features := t.Features()
		fs := make([]engine.Term, len(features))
		for i, f := range features {
			fs[i] = engine.NewAtom(f)
		}

		ret = append(ret, atomToken.Apply(
			engine.Integer(t.ID),
			engine.NewAtom(t.Surface),
			engine.List(fs...),
		))
	}

	// トークンを表す複合項の入ったスライスをリストに変換して出力用の引数と単一化
	// 継続 k はそのまま渡した先で処理してもらう
	return engine.Unify(vm, tokens, engine.List(ret...), k, env)
}
