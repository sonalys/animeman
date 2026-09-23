package nyaaquerier

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder(t *testing.T) {
	t.Run("term", func(t *testing.T) {
		require.Equal(t, "1080", Term("1080").String())
	})

	t.Run("phrase", func(t *testing.T) {
		require.Equal(t, `"1080 HEVC"`, Phrase("1080 HEVC").String())
	})

	t.Run("not", func(t *testing.T) {
		require.Equal(t, `-"dub"`, Not{Phrase("dub")}.String())
	})

	t.Run("and", func(t *testing.T) {
		got := And{Term("1080"), Or{Term("HEVC"), Term("AV1")}}
		require.Equal(t, "1080 HEVC|AV1", got.String())
	})

	t.Run("or wraps alternatives in groups", func(t *testing.T) {
		got := Or{And{Term("1080"), Term("HEVC")}, And{Term("1080"), Term("AV1")}}
		require.Equal(t, "(1080 HEVC)|(1080 AV1)", got.String())
	})

	t.Run("group", func(t *testing.T) {
		got := Group{Term("1080"), Term("HEVC")}
		require.Equal(t, "(1080 HEVC)", got.String())
	})

	t.Run("sanitize strips syntax chars", func(t *testing.T) {
		require.Equal(t, "WEB-DL x", Sanitize(`WEB-DL "x"`))
		require.Equal(t, "grand blue", Sanitize("grand blue:"))
		require.Equal(t, "Erai-raws", Sanitize("Erai-raws"))
	})

	t.Run("phraseOf single word is term", func(t *testing.T) {
		require.Equal(t, Term("1080"), PhraseOf("1080"))
	})

	t.Run("phraseOf multi word is phrase", func(t *testing.T) {
		require.Equal(t, Phrase("1080 HEVC"), PhraseOf("1080 HEVC"))
	})
	t.Run("raw passes through verbatim", func(t *testing.T) {
		require.Equal(t, `"-dub"`, Raw(`"-dub"`).String())
	})
	t.Run("phraseOf with dash", func(t *testing.T) {
		require.Equal(t, Term("WEB-DL"), PhraseOf("WEB-DL"))
	})
}
