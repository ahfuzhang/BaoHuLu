package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const manifestName = ".hulu-manifest.json"
const transactionName = ".hulu-transaction.json"
const lockName = ".hulu-lock"

type Manifest struct {
	Version int             `json:"version"`
	Units   map[string]Unit `json:"units"`
}

type Unit struct {
	Metadata UnitMetadata         `json:"metadata"`
	Files    map[string]OwnedFile `json:"files"`
}

type UnitMetadata struct {
	GeneratorVersion string `json:"generator_version,omitempty"`
	SourceHash       string `json:"source_hash,omitempty"`
	TemplateHash     string `json:"template_hash,omitempty"`
}

type OwnedFile struct {
	Hash       string `json:"hash"`
	Template   string `json:"template"`
	Part       string `json:"part"`
	Symbol     string `json:"symbol,omitempty"`
	CreateOnly bool   `json:"create_only,omitempty"`
}

type WriteResult struct {
	Written   []string
	Removed   []string
	Preserved []string
}

type snapshot struct {
	Path           string      `json:"path"`
	Existed        bool        `json:"existed"`
	Content        []byte      `json:"content,omitempty"`
	Mode           fs.FileMode `json:"mode"`
	ExpectedExists bool        `json:"expected_exists"`
	ExpectedHash   string      `json:"expected_hash,omitempty"`
}

type journal struct {
	Version int        `json:"version"`
	Files   []snapshot `json:"files"`
}

func digest(contents []byte) string {
	value := sha256.Sum256(contents)
	return hex.EncodeToString(value[:])
}

func checkedPath(root, relative string) (string, error) {
	if err := safeRelative(relative); err != nil {
		return "", err
	}
	current := root
	for _, part := range strings.Split(relative, "/") {
		entries, err := os.ReadDir(current)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), part) && entry.Name() != part {
				return "", fmt.Errorf("case-insensitive path collision: %s and %s", entry.Name(), relative)
			}
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return "", fmt.Errorf("output symlink is not allowed: %s", current)
		}
	}
	return current, nil
}

func readSnapshot(root, relative string) (snapshot, error) {
	entry := snapshot{Path: relative, Mode: 0644}
	absolute, err := checkedPath(root, relative)
	if err != nil {
		return entry, err
	}
	info, err := os.Lstat(absolute)
	if errors.Is(err, fs.ErrNotExist) {
		return entry, nil
	}
	if err != nil {
		return entry, err
	}
	if !info.Mode().IsRegular() {
		return entry, fmt.Errorf("output is not a regular file: %s", relative)
	}
	entry.Existed, entry.Mode = true, info.Mode().Perm()
	entry.Content, err = os.ReadFile(absolute)
	return entry, err
}

func atomicWrite(root, relative string, contents []byte, mode fs.FileMode) error {
	absolute, err := checkedPath(root, relative)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absolute), 0755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(absolute), ".hulu-stage-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	if _, err := temporary.Write(contents); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if _, err := checkedPath(root, relative); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), absolute)
}

func restore(root string, transaction journal) error {
	if transaction.Version != 1 {
		return fmt.Errorf("unsupported transaction version %d", transaction.Version)
	}
	for _, original := range transaction.Files {
		if original.Path != manifestName {
			if err := validateOutputPath(original.Path); err != nil {
				return err
			}
		}
		if _, err := checkedPath(root, original.Path); err != nil {
			return err
		}
		current, err := readSnapshot(root, original.Path)
		if err != nil {
			return err
		}
		unchanged := current.Existed == original.Existed && digest(current.Content) == digest(original.Content)
		expected := current.Existed == original.ExpectedExists && (!current.Existed || digest(current.Content) == original.ExpectedHash)
		if !unchanged && !expected {
			return fmt.Errorf("recovery would overwrite a concurrent modification: %s", original.Path)
		}
	}
	for index := len(transaction.Files) - 1; index >= 0; index-- {
		original := transaction.Files[index]
		if original.Existed {
			if err := atomicWrite(root, original.Path, original.Content, original.Mode); err != nil {
				return err
			}
		} else {
			absolute, err := checkedPath(root, original.Path)
			if err != nil {
				return err
			}
			if err := os.Remove(absolute); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
	}
	return nil
}

func Write(root, identity string, outputs []Output, metadata ...UnitMetadata) (*WriteResult, error) {
	if len(metadata) > 1 {
		return nil, fmt.Errorf("only one generation metadata value is allowed")
	}
	if identity == "" {
		return nil, fmt.Errorf("generation identity is required")
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absolute, 0755); err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	root = resolved
	lock, err := os.OpenFile(filepath.Join(root, lockName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("output directory is locked; inspect %s before recovering an interrupted run: %w", filepath.Join(root, lockName), err)
	}
	if _, err := fmt.Fprintf(lock, "%d\n", os.Getpid()); err != nil {
		lock.Close()
		os.Remove(lock.Name())
		return nil, err
	}
	lock.Close()
	defer os.Remove(lock.Name())
	previousTransaction, err := readSnapshot(root, transactionName)
	if err != nil {
		return nil, err
	}
	if previousTransaction.Existed {
		var transaction journal
		if err := json.Unmarshal(previousTransaction.Content, &transaction); err != nil {
			return nil, fmt.Errorf("invalid recovery journal: %w", err)
		}
		if err := restore(root, transaction); err != nil {
			return nil, fmt.Errorf("incomplete generation; recovery refused: %w", err)
		}
		if err := os.Remove(filepath.Join(root, transactionName)); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("recovered interrupted generation; inspect restored files and rerun")
	}
	manifestSnapshot, err := readSnapshot(root, manifestName)
	if err != nil {
		return nil, err
	}
	manifest := Manifest{Version: 1, Units: map[string]Unit{}}
	if manifestSnapshot.Existed {
		if err := json.Unmarshal(manifestSnapshot.Content, &manifest); err != nil {
			return nil, fmt.Errorf("invalid manifest: %w", err)
		}
		if manifest.Version != 1 || manifest.Units == nil {
			return nil, fmt.Errorf("unsupported or invalid output manifest")
		}
	}
	owners := map[string]string{}
	for unitID, unit := range manifest.Units {
		for name := range unit.Files {
			if err := validateOutputPath(name); err != nil {
				return nil, fmt.Errorf("invalid manifest path: %w", err)
			}
			folded := strings.ToLower(name)
			if previous := owners[folded]; previous != "" && previous != unitID {
				return nil, fmt.Errorf("manifest contains multiple owners for %s", name)
			}
			owners[folded] = unitID
		}
	}
	oldUnit := manifest.Units[identity]
	newUnit := Unit{Files: map[string]OwnedFile{}}
	if len(metadata) == 1 {
		newUnit.Metadata = metadata[0]
	}
	result := &WriteResult{}
	writes := map[string]Output{}
	backups := map[string]snapshot{}
	seen := map[string]bool{}
	for _, output := range outputs {
		if err := validateOutputPath(output.Path); err != nil {
			return nil, err
		}
		folded := strings.ToLower(output.Path)
		if seen[folded] {
			return nil, fmt.Errorf("duplicate output: %s", output.Path)
		}
		for existing := range seen {
			if strings.HasPrefix(folded, existing+"/") || strings.HasPrefix(existing, folded+"/") {
				return nil, fmt.Errorf("output file/directory collision: %s", output.Path)
			}
		}
		seen[folded] = true
		if owner := owners[folded]; owner != "" && owner != identity {
			return nil, fmt.Errorf("%s belongs to another generation unit", output.Path)
		}
		original, err := readSnapshot(root, output.Path)
		if err != nil {
			return nil, err
		}
		old, owned := oldUnit.Files[output.Path]
		if original.Existed && output.CreateOnly {
			result.Preserved = append(result.Preserved, output.Path)
			if owned {
				old.CreateOnly = true
				newUnit.Files[output.Path] = old
			}
			continue
		}
		if original.Existed && (!owned || digest(original.Content) != old.Hash) {
			return nil, fmt.Errorf("refusing to overwrite unowned or manually modified file: %s", output.Path)
		}
		symbol := ""
		switch {
		case output.Context.RPC != nil:
			symbol = output.Context.RPC.ID
		case output.Context.Service != nil:
			symbol = output.Context.Service.ID
		case output.Context.Message != nil:
			symbol = output.Context.Message.ID
		case output.Context.Enum != nil:
			symbol = output.Context.Enum.ID
		}
		newUnit.Files[output.Path] = OwnedFile{Hash: digest(output.Content), Template: output.Template, Part: output.Part, Symbol: symbol, CreateOnly: output.CreateOnly}
		if original.Existed && digest(original.Content) == digest(output.Content) {
			continue
		}
		original.ExpectedExists, original.ExpectedHash = true, digest(output.Content)
		writes[output.Path], backups[output.Path] = output, original
	}
	for name, old := range oldUnit.Files {
		if _, keep := newUnit.Files[name]; keep {
			continue
		}
		original, err := readSnapshot(root, name)
		if err != nil {
			return nil, err
		}
		if old.CreateOnly {
			if original.Existed {
				newUnit.Files[name] = old
				result.Preserved = append(result.Preserved, name)
			}
			continue
		}
		if !original.Existed {
			continue
		}
		if digest(original.Content) != old.Hash {
			return nil, fmt.Errorf("refusing to prune manually modified file: %s", name)
		}
		backups[name] = original
		result.Removed = append(result.Removed, name)
	}
	manifest.Units[identity] = newUnit
	contents, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	contents = append(contents, '\n')
	if len(backups) == 0 && digest(manifestSnapshot.Content) == digest(contents) {
		return result, nil
	}
	transaction := journal{Version: 1}
	names := make([]string, 0, len(backups))
	for name := range backups {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		transaction.Files = append(transaction.Files, backups[name])
	}
	manifestSnapshot.ExpectedExists, manifestSnapshot.ExpectedHash = true, digest(contents)
	transaction.Files = append(transaction.Files, manifestSnapshot)
	journalBytes, err := json.Marshal(transaction)
	if err != nil {
		return nil, err
	}
	if err := atomicWrite(root, transactionName, journalBytes, 0600); err != nil {
		return nil, err
	}
	rollback := func(cause error) (*WriteResult, error) {
		if err := restore(root, transaction); err != nil {
			return nil, fmt.Errorf("%w; rollback incomplete, retained %s: %v", cause, transactionName, err)
		}
		if err := os.Remove(filepath.Join(root, transactionName)); err != nil {
			return nil, fmt.Errorf("%w; remove transaction: %v", cause, err)
		}
		return nil, cause
	}
	for _, name := range names {
		current, err := readSnapshot(root, name)
		if err != nil {
			return rollback(err)
		}
		original := backups[name]
		if current.Existed != original.Existed || digest(current.Content) != digest(original.Content) {
			return rollback(fmt.Errorf("output changed concurrently: %s", name))
		}
		if output, writing := writes[name]; writing {
			if err := atomicWrite(root, name, output.Content, original.Mode); err != nil {
				return rollback(err)
			}
			result.Written = append(result.Written, name)
		} else if err := os.Remove(filepath.Join(root, filepath.FromSlash(name))); err != nil {
			return rollback(err)
		}
	}
	currentManifest, err := readSnapshot(root, manifestName)
	if err != nil {
		return rollback(err)
	}
	if currentManifest.Existed != manifestSnapshot.Existed || digest(currentManifest.Content) != digest(manifestSnapshot.Content) {
		return rollback(fmt.Errorf("manifest changed concurrently"))
	}
	if err := atomicWrite(root, manifestName, contents, 0644); err != nil {
		return rollback(err)
	}
	if err := os.Remove(filepath.Join(root, transactionName)); err != nil {
		return nil, fmt.Errorf("generation committed but journal cleanup failed: %w", err)
	}
	sort.Strings(result.Written)
	sort.Strings(result.Removed)
	sort.Strings(result.Preserved)
	return result, nil
}

func Recover(root string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(root, lockName), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("cannot recover locked directory: %w", err)
	}
	lock.Close()
	defer os.Remove(lock.Name())
	original, err := readSnapshot(root, transactionName)
	if err != nil || !original.Existed {
		return err
	}
	var transaction journal
	if err := json.Unmarshal(original.Content, &transaction); err != nil {
		return err
	}
	if err := restore(root, transaction); err != nil {
		return err
	}
	return os.Remove(filepath.Join(root, transactionName))
}
