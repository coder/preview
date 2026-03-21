package preview

import (
	"io"
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
		// trim because BuildTokens preserves the leading whitespace.
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

	t.Run("EmptyOverride", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `resource "a" "b" {
  x = 1
  nested {
    y = 2
  }
}`)
		override := parseBlock(t, `resource "a" "b" {}`)
		mergeBlock(primary, override)

		// Nothing should change — attributes and blocks preserved.
		assert.Equal(t, "1", attrValue(t, primary.Body().Attributes(), "x"))
		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "nested", blocks[0].Type())
	})

	t.Run("EmptyInlineBlock", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `variable "sizes" {}`)
		override := parseBlock(t, `variable "sizes" {
  type    = set(string)
  default = ["a", "b"]
}`)
		mergeBlock(primary, override)

		attrs := primary.Body().Attributes()
		require.Contains(t, attrs, "type")
		require.Contains(t, attrs, "default")
		// Verify the output is valid HCL by re-parsing it.
		_, diags := hclwrite.ParseConfig(primary.BuildTokens(nil).Bytes(), "test.tf", hcl.Pos{Line: 1, Column: 1})
		require.False(t, diags.HasErrors(), diags.Error())
	})

	t.Run("NewNestedBlockType", func(t *testing.T) {
		t.Parallel()
		primary := parseBlock(t, `data "coder_parameter" "foo" {
  name = "foo"
  type = "number"
}`)
		override := parseBlock(t, `data "coder_parameter" "foo" {
  validation {
    monotonic = "increasing"
  }
}`)
		mergeBlock(primary, override)

		attrs := primary.Body().Attributes()
		assert.Equal(t, `"foo"`, attrValue(t, attrs, "name"))
		blocks := primary.Body().Blocks()
		require.Len(t, blocks, 1)
		assert.Equal(t, "validation", blocks[0].Type())
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

// testMergeOverrides wraps mergeOverrides and asserts the contract:
// error-severity diagnostics are always accompanied by a non-nil error.
func testMergeOverrides(t *testing.T, fsys fs.FS) (fs.FS, hcl.Diagnostics, error) {
	t.Helper()
	result, diags, err := mergeOverrides(fsys)
	if err == nil {
		for _, d := range diags {
			if d.Severity == hcl.DiagError {
				t.Fatal("mergeOverrides returned error diagnostic without non-nil error")
			}
		}
	}
	return result, diags, err
}

func TestMergeOverrideFiles(t *testing.T) {
	t.Parallel()

	t.Run("EmptyOverrideFile", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf":     &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
			"override.tf": &fstest.MapFile{Data: []byte(``)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)
		assert.Empty(t, diags)

		content := string(readFile(t, result, "main.tf"))
		assert.Contains(t, content, "x = 1")
	})

	t.Run("NoOverrideFiles", func(t *testing.T) {
		t.Parallel()
		original := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
		}
		result, diags, err := testMergeOverrides(t, original)
		require.NoError(t, err)
		assert.Empty(t, diags)
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
		_, _, err := testMergeOverrides(t, fsys)
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
		result, _, err := testMergeOverrides(t, fsys)
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
		result, _, err := testMergeOverrides(t, fsys)
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
		result, _, err := testMergeOverrides(t, fsys)
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
		result, _, err := testMergeOverrides(t, fsys)
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
		result, _, err := testMergeOverrides(t, fsys)
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

	t.Run("TfJsonOverrideWarning", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf":          &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
			"override.tf.json": &fstest.MapFile{Data: []byte(`{}`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)

		// Should warn about the .tf.json override file.
		require.Len(t, diags, 1)
		assert.Equal(t, hcl.DiagWarning, diags[0].Severity)
		assert.Contains(t, diags[0].Detail, "override.tf.json")

		// Original FS returned since no .tf overrides exist.
		assert.Equal(t, fsys, result)
	})

	t.Run("TfJsonPrimaryNoWarning", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf":       &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
			"other.tf.json": &fstest.MapFile{Data: []byte(`{}`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)

		// No warnings for primary .tf.json files because there are no overrides.
		assert.Empty(t, diags)

		// Original FS returned since no overrides exist.
		assert.Equal(t, fsys, result)
	})

	t.Run("TfJsonFilesPassThrough", func(t *testing.T) {
		t.Parallel()
		jsonContent := []byte(`{"resource": {"a": {"b": {"x": 1}}}}`)
		fsys := fstest.MapFS{
			"main.tf":       &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
			"extra.tf.json": &fstest.MapFile{Data: jsonContent},
			"override.tf":   &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 2 }`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)

		require.Len(t, diags, 1)
		assert.Contains(t, diags[0].Summary, "Primary file uses .tf.json format")

		// .tf.json file should still be accessible in the result FS.
		content := readFile(t, result, "extra.tf.json")
		assert.Equal(t, jsonContent, content)
	})

	t.Run("MixedTfAndTfJsonOverrides", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" {
  x = 1
}`)},
			"a_override.tf":      &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 99 }`)},
			"b_override.tf.json": &fstest.MapFile{Data: []byte(`{}`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)

		// Warning for the .tf.json override only.
		require.Len(t, diags, 1)
		assert.Equal(t, hcl.DiagWarning, diags[0].Severity)
		assert.Contains(t, diags[0].Detail, "b_override.tf.json")

		// .tf override should still be merged.
		content := readFile(t, result, "main.tf")
		assert.Contains(t, string(content), "99")
	})

	t.Run("LocalsAttributeLevelMerge", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`locals {
  foo = "original"
  bar = "keep"
}`)},
			"override.tf": &fstest.MapFile{Data: []byte(`locals {
  foo = "overridden"
}`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)
		assert.Empty(t, diags)

		content := string(readFile(t, result, "main.tf"))
		assert.Contains(t, content, `"overridden"`)
		assert.Contains(t, content, `"keep"`)
		assert.NotContains(t, content, `"original"`)
	})

	t.Run("LocalsAcrossMultipleFiles", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"a.tf": &fstest.MapFile{Data: []byte(`locals {
  from_a = "aaa"
}`)},
			"b.tf": &fstest.MapFile{Data: []byte(`locals {
  from_b = "bbb"
}`)},
			"override.tf": &fstest.MapFile{Data: []byte(`locals {
  from_a = "overridden_a"
  from_b = "overridden_b"
}`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)
		assert.Empty(t, diags)

		contentA := string(readFile(t, result, "a.tf"))
		assert.Contains(t, contentA, `"overridden_a"`)

		contentB := string(readFile(t, result, "b.tf"))
		assert.Contains(t, contentB, `"overridden_b"`)
	})

	t.Run("LocalsMultiplePrimaryAndOverrideBlocks", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`
locals {
  x = "x_orig"
}

locals {
  y = "y_orig"
}`)},
			"other.tf": &fstest.MapFile{Data: []byte(`locals {
  z = "z_orig"
}`)},
			"a_override.tf": &fstest.MapFile{Data: []byte(`locals {
  x = "x_from_a"
  z = "z_from_a"
}`)},
			"b_override.tf": &fstest.MapFile{Data: []byte(`locals {
  y = "y_from_b"
}`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)
		assert.Empty(t, diags)

		mainContent := string(readFile(t, result, "main.tf"))
		assert.Contains(t, mainContent, `"x_from_a"`)
		assert.Contains(t, mainContent, `"y_from_b"`)

		otherContent := string(readFile(t, result, "other.tf"))
		assert.Contains(t, otherContent, `"z_from_a"`)
	})

	t.Run("LocalsNewAttributeErrors", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`locals {
  existing = "yes"
}`)},
			"override.tf": &fstest.MapFile{Data: []byte(`locals {
  nonexistent = "nope"
}`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.Error(t, err)
		assert.Nil(t, result)
		require.Len(t, diags, 1)
		assert.Equal(t, hcl.DiagError, diags[0].Severity)
		assert.Contains(t, diags[0].Summary, "Missing base local")
	})

	t.Run("TerraformBlockSkipped", func(t *testing.T) {
		t.Parallel()
		fsys := fstest.MapFS{
			"main.tf": &fstest.MapFile{Data: []byte(`terraform {
  required_version = ">= 1.0"
}

resource "a" "b" { x = 1 }`)},
			"override.tf": &fstest.MapFile{Data: []byte(`terraform {
  required_version = ">= 2.0"
}

resource "a" "b" { x = 2 }`)},
		}
		result, diags, err := testMergeOverrides(t, fsys)
		require.NoError(t, err)

		// Warning about skipped terraform block.
		require.Len(t, diags, 1)
		assert.Equal(t, hcl.DiagWarning, diags[0].Severity)
		assert.Contains(t, diags[0].Summary, "unsupported 'terraform' block")

		// Resource override still applied.
		content := string(readFile(t, result, "main.tf"))
		assert.Contains(t, content, "x = 2")
		// Terraform block unchanged.
		assert.Contains(t, content, ">= 1.0")
		assert.NotContains(t, content, ">= 2.0")
	})

}

// TestFilteredReadDir verifies that filteredDir.ReadDir(n) with n > 0
// never returns an empty slice with nil error, which would violate
// the fs.ReadDirFile contract.
func TestFilteredReadDir(t *testing.T) {
	t.Parallel()
	// Create an FS where every other file is hidden (override files).
	fsys := fstest.MapFS{
		"main.tf":       &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 1 }`)},
		"a_override.tf": &fstest.MapFile{Data: []byte(`resource "a" "b" { x = 2 }`)},
		"other.tf":      &fstest.MapFile{Data: []byte(`resource "c" "d" { y = 1 }`)},
		"b_override.tf": &fstest.MapFile{Data: []byte(`resource "c" "d" { y = 2 }`)},
	}
	result, _, err := testMergeOverrides(t, fsys)
	require.NoError(t, err)

	f, err := result.Open(".")
	require.NoError(t, err)
	defer f.Close()

	rdf, ok := f.(fs.ReadDirFile)
	require.True(t, ok)

	// Read one entry at a time - should never get empty slice with
	// nil error.
	var names []string
	for {
		entries, err := rdf.ReadDir(1)
		for _, e := range entries {
			names = append(names, e.Name())
		}
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		assert.NotEmpty(t, entries, "ReadDir(1) returned empty slice with nil error")
	}

	assert.Contains(t, names, "main.tf")
	assert.Contains(t, names, "other.tf")
	assert.NotContains(t, names, "a_override.tf")
	assert.NotContains(t, names, "b_override.tf")
}
