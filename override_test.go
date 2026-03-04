package preview

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsOverrideFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		filename string
		expected bool
	}{
		{"override.tf", "override.tf", true},
		{"foo_override.tf", "foo_override.tf", true},
		{"override.tf.json", "override.tf.json", true},
		{"foo_override.tf.json", "foo_override.tf.json", true},
		{"main.tf", "main.tf", false},
		{"overrides.tf", "overrides.tf", false},
		{"my_override_file.tf", "my_override_file.tf", false},
		{"no extension", "override", false},
		{"go file", "override.go", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, isOverrideFile(tc.filename))
		})
	}
}

func TestBlockKey(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		blockType string
		labels    []string
		expected  string
	}{
		{"no labels", "terraform", nil, "terraform"},
		{"one label", "variable", []string{"env"}, "variable.env"},
		{"two labels", "data", []string{"coder_parameter", "region"}, "data.coder_parameter.region"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, blockKey(tc.blockType, tc.labels))
		})
	}
}

func TestMergeBlock(t *testing.T) {
	t.Parallel()

	// attrValue returns the text representation of an attribute's value.
	attrValue := func(t *testing.T, attrs map[string]*hclwrite.Attribute, name string) string {
		t.Helper()
		a, ok := attrs[name]
		require.True(t, ok, "attribute %q not found", name)
		// trim because BuilTokens preserves the leading whitespace.
		return strings.TrimSpace(string(a.Expr().BuildTokens(nil).Bytes()))
	}

	// parseBlock parses HCL source and returns the first block.
	parseBlock := func(t *testing.T, src string) *hclwrite.Block {
		t.Helper()
		f, diags := hclwrite.ParseConfig([]byte(src), "test.tf", hcl.Pos{Line: 1, Column: 1})
		require.False(t, diags.HasErrors(), diags.Error())
		return f.Body().Blocks()[0]
	}

	t.Run("AttributeClobber", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  x = 1
  y = 2
}`)
		override := parseBlock(t, `resource "a" "b" {
  y = 3
}`)
		mergeBlock(primary, override)

		attrs := primary.Body().Attributes()
		assert.Equal(t, "1", attrValue(t, attrs, "x"))
		assert.Equal(t, "3", attrValue(t, attrs, "y"))
	})

	t.Run("AttributeInsertion", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  x = 1
}`)
		override := parseBlock(t, `resource "a" "b" {
  z = "new"
}`)
		mergeBlock(primary, override)

		attrs := primary.Body().Attributes()
		require.Contains(t, attrs, "x")
		require.Contains(t, attrs, "z")
	})

	t.Run("NestedBlockSuppression", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `data "coder_parameter" "disk" {
  name = "disk"
  option {
    name  = "10GB"
    value = 10
  }
  option {
    name  = "20GB"
    value = 20
  }
}`)
		override := parseBlock(t, `data "coder_parameter" "disk" {
  option {
    name  = "30GB"
    value = 30
  }
}`)
		mergeBlock(primary, override)

		// Primary's two options should be replaced by override's single option.
		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "option", blocks[0].Type())
		assert.Equal(t, "30", attrValue(t, blocks[0].Body().Attributes(), "value"))
	})

	t.Run("DynamicStaticSuppression", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  option {
    name = "static"
  }
}`)
		override := parseBlock(t, `resource "a" "b" {
  dynamic "option" {
    for_each = var.options
    content {
      name = option.value
    }
  }
}`)
		mergeBlock(primary, override)

		// Static "option" should be removed and replaced by dynamic "option".
		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "dynamic", blocks[0].Type())
		require.Len(t, blocks[0].Labels(), 1)
		assert.Equal(t, "option", blocks[0].Labels()[0])
	})

	t.Run("StaticDynamicSuppression", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  dynamic "option" {
    for_each = var.options
    content {
      name = option.value
    }
  }
}`)
		override := parseBlock(t, `resource "a" "b" {
  option {
    name = "static"
  }
}`)
		mergeBlock(primary, override)

		// Dynamic "option" should be removed and replaced by static "option".
		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "option", blocks[0].Type())
		assert.Empty(t, blocks[0].Labels())
	})

	t.Run("MixedStaticDynamicSuppression", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  option {
    name = "static"
  }
  dynamic "option" {
    for_each = var.list
    content {
      name = option.value
    }
  }
}`)
		override := parseBlock(t, `resource "a" "b" {
  option {
    name = "replaced"
  }
}`)
		mergeBlock(primary, override)

		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "option", blocks[0].Type())
		assert.Empty(t, blocks[0].Labels())
	})

	t.Run("MixedStaticDynamicSuppressionByDynamic", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  option {
    name = "static"
  }
  dynamic "option" {
    for_each = var.list
    content {
      name = option.value
    }
  }
}`)
		override := parseBlock(t, `resource "a" "b" {
  dynamic "option" {
    for_each = var.other
    content {
      name = option.value
    }
  }
}`)
		mergeBlock(primary, override)

		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "dynamic", blocks[0].Type())
		require.Len(t, blocks[0].Labels(), 1)
		assert.Equal(t, "option", blocks[0].Labels()[0])
	})

	t.Run("StaticSuppressionByMixedOverride", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  option {
    name = "old"
  }
}`)
		override := parseBlock(t, `resource "a" "b" {
  option {
    name = "static"
  }
  dynamic "option" {
    for_each = var.list
    content {
      name = option.value
    }
  }
}`)
		mergeBlock(primary, override)

		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 2)
		assert.Equal(t, "option", blocks[0].Type())
		assert.Equal(t, "dynamic", blocks[1].Type())
		assert.Equal(t, "option", blocks[1].Labels()[0])
	})

	t.Run("NoNestedBlocksInOverride", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  x = 1
  nested {
    y = 2
  }
}`)
		override := parseBlock(t, `resource "a" "b" {
  x = 99
}`)
		mergeBlock(primary, override)

		// Primary's nested blocks should be preserved when override has none.
		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "nested", blocks[0].Type())
		// Attribute should still be overridden.
		assert.Equal(t, "99", attrValue(t, primary.Body().Attributes(), "x"))
	})
}

// readFile reads a file from an fs.FS using Open+Read (since overrideFS
// doesn't implement fs.ReadFileFS).
func readFile(t *testing.T, fsys fs.FS, name string) []byte {
	t.Helper()
	f, err := fsys.Open(name)
	require.NoError(t, err)
	defer f.Close()
	info, err := f.Stat()
	require.NoError(t, err)
	buf := make([]byte, info.Size())
	_, err = f.Read(buf)
	require.NoError(t, err)
	return buf
}

func TestMergeOverrideFiles(t *testing.T) {
	t.Parallel()

	t.Run("NoOverrideFiles", func(t *testing.T) {
		t.Parallel()
		original := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
		}
		result, err := mergeOverrideFiles(original)
		require.NoError(t, err)
		// Should return the exact same FS when there are no overrides.
		assert.Equal(t, original, result)
	})

	t.Run("UnmatchedOverrideBlock", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
			"override.tf": &fstest.MapFile{Data: []byte(`resource "c" "d" {
  y = 2
}`)},
		}
		_, err := mergeOverrideFiles(fsys)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no matching block")
	})

	t.Run("BasicAttributeMerge", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" {
  x = 1
  y = 2
}`)},
			"override.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" {
  y = 99
}`)},
		}
		result, err := mergeOverrideFiles(fsys)
		require.NoError(t, err)

		// Read the merged primary file.
		content := readFile(t, result, "main.tf")
		assert.Contains(t, string(content), "99")
	})

	t.Run("OverrideFileHidden", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf":     &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
			"override.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 2 }`)},
		}
		result, err := mergeOverrideFiles(fsys)
		require.NoError(t, err)

		_, err = result.Open("override.tf")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("DirectoryListingFiltersHidden", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf":         &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
			"foo_override.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 2 }`)},
			"other.txt":       &fstest.MapFile{Data: []byte("hello")},
		}
		result, err := mergeOverrideFiles(fsys)
		require.NoError(t, err)

		f, err := result.Open(".")
		require.NoError(t, err)
		defer f.Close()

		rdf, ok := f.(fs.ReadDirFile)
		require.True(t, ok)

		entries, err := rdf.ReadDir(-1)
		require.NoError(t, err)

		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		assert.Contains(t, names, "main.tf")
		assert.Contains(t, names, "other.txt")
		assert.NotContains(t, names, "foo_override.tf")
	})

	t.Run("SequentialOverrideMerge", func(t *testing.T) {
		t.Parallel()
		// Two override files modify the same block. Because WalkDir processes
		// them in lexical order (a_ before b_), both attributes should be
		// present in the merged result.
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" {
  original = "yes"
}`)},
			"a_override.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" {
  from_a = "aaa"
}`)},
			"b_override.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" {
  from_b = "bbb"
}`)},
		}
		result, err := mergeOverrideFiles(fsys)
		require.NoError(t, err)

		content := readFile(t, result, "main.tf")
		merged := string(content)
		assert.Contains(t, merged, "original")
		assert.Contains(t, merged, "from_a")
		assert.Contains(t, merged, "from_b")
	})

	t.Run("SubdirectoryOverrides", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`resource "root" "r" { x = 1 }`)},
			"modules/child/main.tf": &fstest.MapFile{Data: []byte(`resource "child" "c" {
  y = 1
}`)},
			"modules/child/override.tf": &fstest.MapFile{Data: []byte(`resource "child" "c" {
  y = 42
}`)},
		}
		result, err := mergeOverrideFiles(fsys)
		require.NoError(t, err)

		// Root file should be unchanged (no overrides in root).
		rootContent := readFile(t, result, "main.tf")
		assert.Contains(t, string(rootContent), "x = 1")

		// Child module should have merged content.
		childContent := readFile(t, result, "modules/child/main.tf")
		assert.Contains(t, string(childContent), "y = 42")

		// Override file in subdirectory should be hidden.
		_, err = result.Open("modules/child/override.tf")
		require.Error(t, err)
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
}
