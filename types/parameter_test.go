package types_test

import (
	"crypto/sha256"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/preview/types"
)

// TestParameterEquality might be a bit pointless. It just ensures the
// Hash function returns a consistent value for the same input.
// TODO: Remove this, just wanted to create some random parameters
func TestParameterEquality(t *testing.T) {
	t.Parallel()

	for i := 0; i < 100; i++ {
		t.Run(fmt.Sprintf("EqualityCheck_%d", i), func(t *testing.T) {
			t.Parallel()
			seed := sha256.Sum256([]byte(t.Name()))
			src := rand.NewChaCha8(seed)

			param := randomParameter(src)
			a, err := param.Hash()
			require.NoError(t, err)

			b, err := param.Hash()
			require.NoError(t, err)

			require.Equal(t, a, b)
		})
	}
}

func randomParameter(src *rand.ChaCha8) *types.RichParameter {
	ty := randomElement(src, "string", "number", "bool", "list(string)")
	opts := make([]*types.ParameterOption, randomInt(src, 0, 5))
	for i := range opts {
		opts[i] = randomParameterOption(src, ty)
	}

	return &types.RichParameter{
		Name:         randomString(src, 20),
		Description:  randomString(src, 20),
		Type:         ty,
		Mutable:      randomElement(src, true, false),
		DefaultValue: randomValue(src, ty),
		Icon:         randomString(src, 10),
		Options:      opts,
		Validation:   randomValidation(src, ty),
		Required:     randomElement(src, true, false),
		DisplayName:  randomString(src, 10),
		Order:        int32(randomInt(src, 0, 10)),
		Ephemeral:    randomElement(src, true, false),
	}
}

func randomValidation(src *rand.ChaCha8, ty string) *types.ParameterValidation {
	var minVal, maxVal *int32
	mono := ""
	if ty == "number" {
		mv := int32(randomInt(src, 0, 10))
		mxv := int32(randomInt(src, uint64(mv), uint64(10+mv)))
		minVal = &mv
		maxVal = &mxv
		mono = randomElement(src, "", "increasing", "decreasing")
	}

	return &types.ParameterValidation{
		Regex:     randomString(src, 10),
		Error:     randomString(src, 10),
		Min:       minVal,
		Max:       maxVal,
		Monotonic: mono,
	}
}

func randomParameterOption(src *rand.ChaCha8, ty string) *types.ParameterOption {
	return &types.ParameterOption{
		Name:        randomString(src, 10),
		Description: randomString(src, 20),
		Value:       randomValue(src, ty),
		Icon:        randomString(src, 10),
	}
}

func randomValue(src *rand.ChaCha8, ty string) string {
	switch ty {
	case "string":
		return randomString(src, 20)
	case "number":
		return fmt.Sprintf("%d", randomInt(src, 0, 100))
	case "bool":
		return fmt.Sprintf("%t", randomElement(src, true, false))
	case "list(string)":
		elems := make([]string, randomInt(src, 0, 5))
		for i := range elems {
			elems[i] = fmt.Sprintf("%q", randomString(src, 7))
		}
		return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
	}
	panic(fmt.Sprintf("unsupported type: %s", ty))
}

func randomInt(src *rand.ChaCha8, min, max uint64) int64 {
	v := src.Uint64() % max
	return int64(v) + int64(min)
}

func randomElement[T any](src *rand.ChaCha8, elements ...T) T {
	if len(elements) > 255 {
		panic(fmt.Sprintf("randomElement only supports 255 elements, got %d", len(elements)))
	}
	b := make([]byte, 1)
	_, err := src.Read(b)
	if err != nil {
		panic(fmt.Errorf("rand.Read: %w", err))
	}
	return elements[int(b[0])%len(elements)]
}

func randomString(src *rand.ChaCha8, length int) string {
	out := make([]byte, length)
	_, err := src.Read(out)
	if err != nil {
		panic(fmt.Errorf("rand.Read: %w", err))
	}

	for i := range out {
		out[i] = out[i] % (26 * 2)
		if out[i] >= 26 {
			out[i] += 'A' - 26 // subtract off the 0-26 lowercase
		} else {
			out[i] += 'a'
		}
	}
	return string(out)
}
