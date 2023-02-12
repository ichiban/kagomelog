package kagomelog

import (
	"context"
	"github.com/ichiban/prolog/engine"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		title               string
		input, mode, tokens engine.Term
		ok                  bool
		err                 error
		env                 map[engine.Variable]engine.Term
	}{
		{title: "ok", input: engine.NewAtom("すもももももももものうち"), mode: atomNormal, tokens: engine.List(
			atomToken.Apply(engine.NewVariable(), engine.NewAtom("すもも"), engine.PartialList(engine.NewVariable(), engine.NewAtom("名詞"))),
			atomToken.Apply(engine.NewVariable(), engine.NewAtom("も"), engine.PartialList(engine.NewVariable(), engine.NewAtom("助詞"))),
			atomToken.Apply(engine.NewVariable(), engine.NewAtom("もも"), engine.PartialList(engine.NewVariable(), engine.NewAtom("名詞"))),
			atomToken.Apply(engine.NewVariable(), engine.NewAtom("も"), engine.PartialList(engine.NewVariable(), engine.NewAtom("助詞"))),
			atomToken.Apply(engine.NewVariable(), engine.NewAtom("もも"), engine.PartialList(engine.NewVariable(), engine.NewAtom("名詞"))),
			atomToken.Apply(engine.NewVariable(), engine.NewAtom("の"), engine.PartialList(engine.NewVariable(), engine.NewAtom("助詞"))),
			atomToken.Apply(engine.NewVariable(), engine.NewAtom("うち"), engine.PartialList(engine.NewVariable(), engine.NewAtom("名詞"))),
		), ok: true},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			var vm engine.VM
			ok, err := Analyze(&vm, tt.input, tt.mode, tt.tokens, func(env *engine.Env) *engine.Promise {
				for k, v := range tt.env {
					_, ok := env.Unify(k, v)
					assert.True(t, ok)
				}
				return engine.Bool(true)
			}, nil).Force(context.Background())
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.err, err)
		})
	}
}
