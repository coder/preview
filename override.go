package preview

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// mergeOverrideFiles scans the filesystem for Terraform override files and
// returns a new FS where override content has been merged into primary files
// using Terraform's override semantics. If no override files exist, the
// original FS is returned unchanged.
//
// Override files are identified by Terraform's naming convention:
// "override.tf", "*_override.tf", and their .tf.json variants.
//
// https://developer.hashicorp.com/terraform/language/files/override
//
// Note: Terraform validates primary blocks before merging overrides, so it
// rejects a primary block that is missing a required attribute even if the
// override would supply it. We merge first, so this edge case passes
// validation. This is immaterial in practice: Coder runs terraform plan
// during template import, which would catch it before the template is saved.
func mergeOverrideFiles(original fs.FS) (fs.FS, error) {
	// Group files by directory, separating primary from override files.
	// Walk the entire tree, not just the root directory, because Trivy's
	// EvaluateAll processes all modules, so we need to pre-merge overrides at
	// every level before Trivy sees the FS.
	type dirFiles struct {
		primary  []string
		override []string
	}
	dirs := make(map[string]*dirFiles)

	err := fs.WalkDir(original, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip dirs; we deal with them by acting on their files.
		if d.IsDir() {
			return nil
		}
		if tfFileExt(d.Name()) == "" {
			return nil
		}

		dir := path.Dir(p)
		if dirs[dir] == nil {
			dirs[dir] = &dirFiles{}
		}

		if isOverrideFile(d.Name()) {
			dirs[dir].override = append(dirs[dir].override, p)
		} else {
			dirs[dir].primary = append(dirs[dir].primary, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk filesystem: %w", err)
	}

	// We are a no-op if there are no override files at all.
	hasOverrides := false
	for _, dir := range dirs {
		if len(dir.override) > 0 {
			hasOverrides = true
			break
		}
	}
	if !hasOverrides {
		return original, nil
	}

	replaced := make(map[string][]byte)
	hidden := make(map[string]bool)

	for _, dir := range dirs {
		if len(dir.override) == 0 {
			continue
		}

		// Parse all primary files upfront so override files can be applied
		// sequentially, each merging into the already-merged result.
		type primaryState struct {
			path     string
			file     *hclwrite.File
			modified bool
		}
		primaries := make([]*primaryState, 0, len(dir.primary))
		for _, path := range dir.primary {
			content, err := fs.ReadFile(original, path)
			if err != nil {
				return nil, fmt.Errorf("read file %s: %w", path, err)
			}
			f, diags := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return nil, fmt.Errorf("parse file %s: %s", path, diags.Error())
			}
			primaries = append(primaries, &primaryState{path: path, file: f})
		}

		// Process each override file sequentially. If multiple override files
		// define the same block, each merges into the already-merged primary,
		// matching Terraform's behavior.
		for _, path := range dir.override {
			content, err := fs.ReadFile(original, path)
			if err != nil {
				return nil, fmt.Errorf("read file %s: %w", path, err)
			}

			f, diags := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return nil, fmt.Errorf("parse file %s: %s", path, diags.Error())
			}

			for _, oblock := range f.Body().Blocks() {
				key := blockKey(oblock.Type(), oblock.Labels())
				matched := false
				for _, primary := range primaries {
					for _, pblock := range primary.file.Body().Blocks() {
						if blockKey(pblock.Type(), pblock.Labels()) == key {
							mergeBlock(pblock, oblock)
							primary.modified = true
							matched = true
							break
						}
					}
					if matched {
						break
					}
				}
				if !matched {
					// Terraform requires every override block to have a corresponding
					// primary block — override files can only modify, not create.
					return nil, fmt.Errorf("override block %s in %s has no matching block in a primary configuration file", key, path)
				}
			}

			hidden[path] = true
		}

		// Collect modified primary files.
		for _, p := range primaries {
			if p.modified {
				replaced[p.path] = p.file.Bytes()
			}
		}
	}

	return &overrideFS{
		base:     original,
		replaced: replaced,
		hidden:   hidden,
	}, nil
}

// mergeBlock applies override attributes and child blocks to a primary block
// using Terraform's prepareContent semantics.
//
//   - Attributes: each override attribute replaces the corresponding primary
//     attribute, or is inserted if it does not exist in the primary block.
//
//   - Child blocks: if override has any block of type X (including dynamic "X"),
//     all blocks of type X and dynamic "X" are removed from primary. Then all
//     override child blocks are appended — both replacing suppressed types and
//     introducing entirely new block types not present in the primary.
//
// Ref: https://github.com/hashicorp/terraform/blob/7960f60d2147d43f5cf675a898438f6a6693da1b/internal/configs/module_merge_body.go#L76-L121
//
// Note: Terraform re-validates type/default compatibility after variable block
// merge.  We rely on Trivy's evaluator to do that during evaluation of the
// merged HCL.
func mergeBlock(primary, override *hclwrite.Block) {
	// Merge attributes: override clobbers base.
	for name, attr := range override.Body().Attributes() {
		primary.Body().SetAttributeRaw(name, attr.Expr().BuildTokens(nil))
	}

	// Determine which child (nested) block types are overridden.
	overriddenBlockTypes := make(map[string]bool)
	for _, child := range override.Body().Blocks() {
		if child.Type() == "dynamic" && len(child.Labels()) > 0 {
			overriddenBlockTypes[child.Labels()[0]] = true
		} else {
			overriddenBlockTypes[child.Type()] = true
		}
	}

	if len(overriddenBlockTypes) == 0 {
		return
	}

	// Remove overridden block types from primary.
	// Collect blocks to remove first to avoid modifying during iteration.
	var toRemove []*hclwrite.Block
	for _, child := range primary.Body().Blocks() {
		shouldRemove := false
		if child.Type() == "dynamic" && len(child.Labels()) > 0 {
			shouldRemove = overriddenBlockTypes[child.Labels()[0]]
		} else {
			shouldRemove = overriddenBlockTypes[child.Type()]
		}
		if shouldRemove {
			toRemove = append(toRemove, child)
		}
	}
	for _, block := range toRemove {
		primary.Body().RemoveBlock(block)
	}

	// Append all override child blocks.
	for _, child := range override.Body().Blocks() {
		primary.Body().AppendBlock(child)
	}
}

// isOverrideFile returns true if the filename matches Terraform's override
// file naming convention: "override.tf", "*_override.tf", and .tf.json variants.
//
// Ref: https://github.com/hashicorp/terraform/blob/7960f60d2147d43f5cf675a898438f6a6693da1b/internal/configs/parser_file_matcher.go#L161-L170
func isOverrideFile(filename string) bool {
	name := path.Base(filename)
	ext := tfFileExt(name)
	if ext == "" {
		return false
	}
	baseName := name[:len(name)-len(ext)]
	return baseName == "override" || strings.HasSuffix(baseName, "_override")
}

// tfFileExt returns the Terraform file extension (".tf" or ".tf.json") if
// present, or "" otherwise.
func tfFileExt(name string) string {
	if strings.HasSuffix(name, ".tf.json") {
		return ".tf.json"
	}
	if strings.HasSuffix(name, ".tf") {
		return ".tf"
	}
	return ""
}

// blockKey returns a string that uniquely identifies a block for override
// matching purposes. Two blocks with the same key represent the same logical
// entity (one primary, one override).
func blockKey(blockType string, labels []string) string {
	if len(labels) == 0 {
		return blockType
	}
	return blockType + "." + strings.Join(labels, ".")
}
