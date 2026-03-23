package preview

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// primaryState tracks a parsed primary .tf file during override
// merging.
type primaryState struct {
	path     string
	file     *hclwrite.File
	modified bool
}

// mergeOverrides scans the filesystem for .tf Terraform override files
// and returns a new FS where override content has been merged into primary
// files using Terraform's override semantics.
// If no override files are found, the original FS is returned unchanged.
// If an error is encountered, diagnostics are returned in addition to a
// non-nil error.
// Warning diagnostics may also be returned on success (e.g. for skipped
// .tf.json files).
//
// Override files are identified by Terraform's naming convention:
// "override.tf", "*_override.tf", and their .tf.json variants. We only support
// .tf files; .tf.json files get a diagnostic warning and are excluded from
// override merging.
//
// Ref: https://developer.hashicorp.com/terraform/language/files/override
func mergeOverrides(origFS fs.FS) (fs.FS, hcl.Diagnostics, error) {
	// Group files by directory, separating primary from override files.
	// Walk the entire tree, not just the root directory, because Trivy's
	// EvaluateAll processes all modules, so we need to pre-merge overrides at
	// every level before Trivy sees the FS.
	type dirFiles struct {
		primaries []string
		overrides []string
		// Used to generate warnings at merge stage.
		jsonPrimaries []string
	}
	dirs := make(map[string]*dirFiles)

	var warnings hcl.Diagnostics

	err := fs.WalkDir(origFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip dirs; we deal with them by acting on their files.
		if d.IsDir() {
			return nil
		}

		ext := tfFileExt(d.Name())
		if ext == "" {
			return nil
		}

		dir := path.Dir(p)
		if dirs[dir] == nil {
			dirs[dir] = &dirFiles{}
		}

		// We don't support parsing .tf.json files. They remain in the
		// FS for Trivy to parse directly but never participate in
		// override merging.
		if ext == ".tf.json" {
			if isOverrideFile(d.Name()) {
				warnings = warnings.Append(&hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  "Override file uses unsupported .tf.json format",
					Detail:   fmt.Sprintf("%s skipped for override merging", p),
				})
			} else {
				// Save the name of the .tf.json primary so we issue a
				// warning only if we do merging for the dir (less noise).
				dirs[dir].jsonPrimaries = append(dirs[dir].jsonPrimaries, p)
			}
			return nil
		}

		if isOverrideFile(d.Name()) {
			dirs[dir].overrides = append(dirs[dir].overrides, p)
		} else {
			dirs[dir].primaries = append(dirs[dir].primaries, p)
		}
		return nil
	})
	if err != nil {
		return nil, warnings, fmt.Errorf("error reading template files: %w", err)
	}

	hasOverrides := false
	for _, dir := range dirs {
		if len(dir.overrides) > 0 {
			hasOverrides = true
			break
		}
	}
	if !hasOverrides {
		// We are a no-op if there are no supported override files at
		// all. Include warnings so callers know about ignored
		// .tf.json files.
		return origFS, warnings, nil
	}

	replaced := make(map[string][]byte)
	hidden := make(map[string]bool)

	for _, dir := range dirs {
		if len(dir.overrides) == 0 {
			continue
		}

		for _, jp := range dir.jsonPrimaries {
			warnings = warnings.Append(&hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  "Primary file uses .tf.json format",
				Detail:   fmt.Sprintf("%s skipped for override merging", jp),
			})
		}

		// Parse all primary files upfront so override files can be applied
		// sequentially, each merging into the already-merged result.
		primaries := make([]*primaryState, 0, len(dir.primaries))
		for _, path := range dir.primaries {
			content, err := fs.ReadFile(origFS, path)
			if err != nil {
				return nil, warnings, fmt.Errorf("error reading file %s: %w", path, err)
			}
			f, diags := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return nil, warnings.Extend(diags), errors.New("error parsing file")
			}
			primaries = append(primaries, &primaryState{path: path, file: f})
		}

		// Process each override file sequentially. If multiple override files
		// define the same block, each merges into the already-merged primary,
		// matching Terraform's behavior.
		for _, path := range dir.overrides {
			content, err := fs.ReadFile(origFS, path)
			if err != nil {
				return nil, warnings, fmt.Errorf("error reading file %s: %w", path, err)
			}

			f, diags := hclwrite.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1})
			if diags.HasErrors() {
				return nil, warnings.Extend(diags), errors.New("error parsing file")
			}

			for _, oblock := range f.Body().Blocks() {
				// "locals" blocks are label-less and Terraform merges
				// them at the individual attribute level, not at the
				// block level.
				if oblock.Type() == "locals" {
					diags := mergeLocalsBlock(primaries, oblock, path)
					if diags.HasErrors() {
						return nil, warnings.Extend(diags), errors.New("error merging 'locals' block")
					}
					continue
				}
				// 'terraform' block override semantics are too nuanced
				// to implement right now. Hopefully they are rare in
				// practice.
				if oblock.Type() == "terraform" {
					warnings = warnings.Append(&hcl.Diagnostic{
						Severity: hcl.DiagWarning,
						Summary:  "Override file has unsupported 'terraform' block",
						Detail:   fmt.Sprintf("'terraform' block in %s skipped for override merging", path),
					})
					continue
				}

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
					return nil, warnings, fmt.Errorf("override block %q in %s has no matching block in a primary file", key, path)
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
		base:     origFS,
		replaced: replaced,
		hidden:   hidden,
	}, warnings, nil
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
func mergeBlock(primary, override *hclwrite.Block) {
	// hclwrite preserves the formatting of the original block. If the
	// primary body is empty and inline (e.g. `variable "x" {}`),
	// inserting attributes places them on the same line as the
	// opening brace, which HCL rejects. A newline defensively forces
	// multi-line formatting.
	if len(primary.Body().Attributes()) == 0 && len(primary.Body().Blocks()) == 0 {
		primary.Body().AppendNewline()
	}

	// Merge attributes: override clobbers base.
	for name, attr := range override.Body().Attributes() {
		primary.Body().SetAttributeRaw(name, attr.Expr().BuildTokens(nil))
	}

	// Merge blocks: determine which child (nested) block types are
	// overridden.
	overriddenBlockTypes := make(map[string]bool)
	for _, child := range override.Body().Blocks() {
		// E.g. `dynamic "option" {...}`
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

// mergeLocalsBlock merges an override locals block into the primaries
// at the individual attribute level. Each override attribute replaces
// the matching attribute in whichever primary locals block defines
// it. Attributes not found in any primary block produce an error,
// matching Terraform's "Missing base local value definition to
// override" behavior.
// Ref: https://github.com/hashicorp/terraform/blob/7960f60d2147d43f5cf675a898438f6a6693da1b/internal/configs/module.go#L772-L784
func mergeLocalsBlock(primaries []*primaryState, override *hclwrite.Block, overridePath string) hcl.Diagnostics {
	var diags hcl.Diagnostics
	for name, attr := range override.Body().Attributes() {
		found := false
		for _, primary := range primaries {
			for _, pblock := range primary.file.Body().Blocks() {
				if pblock.Type() != "locals" {
					continue
				}
				// NOTE: We don't insert new attrs into an empty body.
				// If that ever changes, empty inline blocks (e.g.
				// `locals {}`) would need the same AppendNewline fix
				// as mergeBlock to avoid same line usage that breaks
				// HCL.
				if _, exists := pblock.Body().Attributes()[name]; exists {
					pblock.Body().SetAttributeRaw(name, attr.Expr().BuildTokens(nil))
					primary.modified = true
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Missing base local value definition to override",
				Detail:   fmt.Sprintf("Local %q in %s has no base definition to override", name, overridePath),
			})
		}
	}
	return diags
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
