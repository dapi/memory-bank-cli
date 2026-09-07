package ownership

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/dapi/memory-bank-cli/internal/contracts"
)

type componentJournal struct {
	SchemaVersion int                       `json:"schema_version"`
	State         string                    `json:"state"`
	Before        map[string]observation    `json:"before"`
	After         map[string]observation    `json:"after"`
	Backups       map[string]string         `json:"backups"`
	Directories   map[string]directoryState `json:"directories"`
}

func prepareComponentJournal(repo pinnedRepo, options Options, mutations []mutation, staged []stagedMutation, staging string) (*componentJournal, error) {
	j := &componentJournal{1, "prepared", map[string]observation{}, map[string]observation{}, map[string]string{}, map[string]directoryState{}}
	for p, o := range options.componentObservations {
		j.Before[p] = o
		j.After[p] = o
	}
	for i, item := range mutations {
		if item.topology != nil {
			return nil, errors.New("component file/directory topology replacement requires an explicit migration")
		}
		p := item.decision.Path
		o, _, err := observeComponent(repo, p)
		if err != nil {
			return nil, err
		}
		if expected, ok := j.Before[p]; ok && !sameObservation(o, expected) {
			return nil, fmt.Errorf("component input changed before journal: %s", p)
		}
		j.Before[p] = o
		j.After[p] = observation{}
		if item.decision.Action != Delete {
			mode := staged[i].replacementInfo.Mode().Perm()
			j.After[p] = observation{true, digest(item.data), gitMode(mode), fmt.Sprintf("%04o", mode)}
		}
		if o.Exists {
			j.Backups[p] = fmt.Sprintf("old/%06d", i)
		}
	}
	for _, p := range contracts.Keys(j.Before) {
		actual, _, err := observeComponent(repo, p)
		if err != nil {
			return nil, err
		}
		if !sameObservation(actual, j.Before[p]) {
			return nil, fmt.Errorf("input changed before journal: %s", p)
		}
		if actual.Exists {
			if err = syncComponentFile(repo, p); err != nil {
				return nil, err
			}
		}
		for parent := path.Dir(p); parent != "."; parent = path.Dir(parent) {
			if _, ok := j.Directories[parent]; ok {
				continue
			}
			if err = checkComponentAncestors(repo, parent+"/journal-check"); err != nil {
				return nil, err
			}
			info, e := os.Lstat(filepath.Join(repo.root, filepath.FromSlash(parent)))
			state := directoryState{}
			if e == nil {
				if !info.IsDir() || info.Mode()&(os.ModeSymlink|os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
					return nil, fmt.Errorf("unsafe component directory %s", parent)
				}
				state = directoryState{true, fmt.Sprintf("%04o", info.Mode().Perm()), true, fmt.Sprintf("%04o", info.Mode().Perm())}
			} else if !errors.Is(e, os.ErrNotExist) {
				return nil, e
			}
			j.Directories[parent] = state
		}
	}
	for _, item := range mutations {
		if item.decision.Action == Delete {
			continue
		}
		for parent := path.Dir(item.decision.Path); parent != "."; parent = path.Dir(parent) {
			state := j.Directories[parent]
			if !state.BeforeExists {
				state.AfterExists = true
				state.AfterMode = "0755"
				j.Directories[parent] = state
			}
		}
	}
	if options.componentDirectories != nil && !reflect.DeepEqual(options.componentDirectories, j.Directories) {
		return nil, errors.New("planned directory state changed before journal")
	}
	// A caller-supplied draft is read-only, but recovery binds its original
	// bytes too. Keep an independently synced snapshot until cleanup succeeds.
	if p := options.componentDraftInput; p != "" {
		actual, data, err := observeComponent(repo, p)
		if err != nil {
			return nil, err
		}
		if !actual.Exists || !sameObservation(actual, j.Before[p]) {
			return nil, errors.New("draft changed before recovery snapshot")
		}
		inputDir := filepath.Join(staging, "inputs")
		if err = os.Mkdir(inputDir, 0700); err != nil {
			return nil, err
		}
		slot := "inputs/000000"
		if err = os.WriteFile(filepath.Join(staging, slot), data, 0600); err != nil {
			return nil, err
		}
		if err = syncLocalRegular(filepath.Join(staging, slot)); err != nil {
			return nil, err
		}
		if err = syncLocalDirectory(inputDir); err != nil {
			return nil, err
		}
		j.Backups[p] = slot
	}
	for _, item := range staged {
		if item.replacement != "" {
			if err := syncLocalRegular(item.replacement); err != nil {
				return nil, err
			}
		}
	}
	if err := verifyComponentInventory(repo, j.Before); err != nil {
		return nil, err
	}
	if err := writeComponentJournal(staging, j); err != nil {
		return nil, err
	}
	if err := syncLocalDirectory(repo.root); err != nil {
		return nil, err
	}
	return j, nil
}
func writeComponentJournal(staging string, j *componentJournal) error {
	b, err := contracts.Canonical(j)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	temporary := filepath.Join(staging, "recovery.json.next")
	f, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(b)
	syncErr := f.Sync()
	closeErr := f.Close()
	if err = errors.Join(writeErr, syncErr, closeErr); err != nil {
		return err
	}
	if err = os.Rename(temporary, filepath.Join(staging, "recovery.json")); err != nil {
		return err
	}
	return syncLocalDirectory(staging)
}
func commitComponentJournal(repo pinnedRepo, j *componentJournal, staging string) error {
	if err := verifyComponentInventory(repo, j.After); err != nil {
		return err
	}
	if err := verifyComponentDirectoryStates(repo, j, true); err != nil {
		return err
	}
	for _, p := range contracts.Keys(j.After) {
		o, _, err := observeComponent(repo, p)
		if err != nil {
			return err
		}
		if !sameObservation(o, j.After[p]) {
			return fmt.Errorf("component after state changed: %s", p)
		}
		if o.Exists {
			if err = syncComponentFile(repo, p); err != nil {
				return err
			}
		}
	}
	for _, p := range contracts.Keys(j.Directories) {
		s := j.Directories[p]
		if s.AfterExists {
			if err := syncComponentDirectory(repo, p); err != nil {
				return err
			}
		}
	}
	if err := syncLocalDirectory(repo.root); err != nil {
		return err
	}
	j.State = "committed"
	return writeComponentJournal(staging, j)
}
func syncComponentParents(repo pinnedRepo, p, backupParent string) error {
	if err := syncComponentDirectory(repo, path.Dir(p)); err != nil {
		return err
	}
	return syncLocalDirectory(backupParent)
}
func syncLocalRegular(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func syncLocalDirectory(p string) error { return syncLocalRegular(p) }

var recoveryBackupPattern = regexp.MustCompile(`^(?:old/[0-9]{6}|inputs/000000)$`)
var permissionPattern = regexp.MustCompile(`^0[0-7]{3}$`)

func validateJournal(j componentJournal) error {
	if j.SchemaVersion != 1 || (j.State != "prepared" && j.State != "committed") || len(j.Before) == 0 || !reflect.DeepEqual(contracts.Keys(j.Before), contracts.Keys(j.After)) {
		return errors.New("unknown recovery journal contract")
	}
	if err := contracts.CheckPortable(contracts.Keys(j.Before)); err != nil {
		return err
	}
	for p, o := range j.Before {
		if !contracts.ValidPath(p) || strings.HasPrefix(p, ".memory-bank-update-") {
			return errors.New("unsafe recovery target")
		}
		for parent := path.Dir(p); parent != "."; parent = path.Dir(parent) {
			if _, exists := j.Directories[parent]; !exists {
				return errors.New("recovery journal omits input ancestor")
			}
		}
		for _, o := range []observation{o, j.After[p]} {
			if o.Exists {
				if !contracts.ValidDigest(o.Digest) || !modePattern.MatchString(o.Mode) || !permissionPattern.MatchString(o.Permissions) {
					return errors.New("invalid recovery observation")
				}
				var bits uint32
				fmt.Sscanf(o.Permissions, "%o", &bits)
				if gitMode(os.FileMode(bits)) != o.Mode {
					return errors.New("inconsistent recovery permissions")
				}
			} else if o.Digest != "" || o.Mode != "" || o.Permissions != "" {
				return errors.New("invalid absent observation")
			}
		}
	}
	slots := map[string]bool{}
	for p, slot := range j.Backups {
		before, ok := j.Before[p]
		if !ok || !before.Exists || (strings.HasPrefix(slot, "old/") == sameObservation(before, j.After[p])) || !recoveryBackupPattern.MatchString(slot) || slots[slot] {
			return errors.New("invalid recovery backup mapping")
		}
		slots[slot] = true
	}
	for p, d := range j.Directories {
		if !contracts.ValidPath(p) || strings.HasPrefix(p, ".memory-bank-update-") {
			return errors.New("unsafe recovery directory")
		}
		if d.BeforeExists != permissionPattern.MatchString(d.BeforeMode) || d.AfterExists != permissionPattern.MatchString(d.AfterMode) || (!d.BeforeExists && d.BeforeMode != "") || (!d.AfterExists && d.AfterMode != "") {
			return errors.New("invalid recovery directory state")
		}
	}
	return nil
}

// Recovery only verifies a complete owner-restored before state, or a durably
// committed after state. It never replays changes into repository targets.
func checkComponentRecovery(repo pinnedRepo, cleanup bool) error {
	entries, err := os.ReadDir(repo.root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".memory-bank-update-") {
			continue
		}
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return errors.New("recovery_required: unsafe retained staging")
		}
		staging := filepath.Join(repo.root, entry.Name())
		journalPath := filepath.Join(staging, "recovery.json")
		info, e := os.Lstat(journalPath)
		if e != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("recovery_required: missing safe journal in %s", staging)
		}
		b, e := os.ReadFile(journalPath)
		if e != nil {
			return e
		}
		var j componentJournal
		if e = contracts.Decode(b, &j); e != nil {
			return fmt.Errorf("recovery_required: %w", e)
		}
		canonical, e := contracts.Canonical(j)
		if e != nil || string(append(canonical, '\n')) != string(b) {
			return errors.New("recovery_required: noncanonical journal")
		}
		if e = validateJournal(j); e != nil {
			return fmt.Errorf("recovery_required: %w", e)
		}
		expected := j.Before
		if j.State == "committed" {
			expected = j.After
		}
		for _, p := range contracts.Keys(expected) {
			o, _, e := observeComponent(repo, p)
			if e != nil || !sameObservation(o, expected[p]) {
				return fmt.Errorf("recovery_required: restore complete %s state for %s using %s", j.State, p, staging)
			}
		}
		for _, p := range contracts.Keys(j.Directories) {
			d := j.Directories[p]
			exists, mode := d.BeforeExists, d.BeforeMode
			if j.State == "committed" {
				exists, mode = d.AfterExists, d.AfterMode
			}
			if e = checkComponentAncestors(repo, p+"/recovery-check"); e != nil {
				return e
			}
			info, e := os.Lstat(filepath.Join(repo.root, filepath.FromSlash(p)))
			if !exists && errors.Is(e, os.ErrNotExist) {
				continue
			}
			if e != nil || !exists || !info.IsDir() || fmt.Sprintf("%04o", info.Mode().Perm()) != mode {
				return fmt.Errorf("recovery_required: directory %s", p)
			}
		}
		if j.State == "committed" {
			handled, _, _, e := ValidateComponents(repo.root, "")
			if !handled || e != nil {
				return fmt.Errorf("recovery_required: committed integrity: %v", e)
			}
		}
		if cleanup {
			if _, e = inspectRepoRoot(repo.root, repo.info); e != nil {
				return e
			}
			if e = os.RemoveAll(staging); e != nil {
				return e
			}
			if e = syncLocalDirectory(repo.root); e != nil {
				return e
			}
		}
	}
	return nil
}

func verifyComponentDirectoryStates(repo pinnedRepo, j *componentJournal, after bool) error {
	for _, p := range contracts.Keys(j.Directories) {
		d := j.Directories[p]
		exists, mode := d.BeforeExists, d.BeforeMode
		if after {
			exists, mode = d.AfterExists, d.AfterMode
		}
		if err := checkComponentAncestors(repo, p+"/directory-check"); err != nil {
			return err
		}
		info, err := os.Lstat(filepath.Join(repo.root, filepath.FromSlash(p)))
		if !exists && errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil || !exists || !info.IsDir() || info.Mode()&(os.ModeSymlink|os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 || fmt.Sprintf("%04o", info.Mode().Perm()) != mode {
			return fmt.Errorf("component directory state changed: %s", p)
		}
	}
	return nil
}
func verifyComponentInventory(repo pinnedRepo, expected map[string]observation) error {
	return filepath.WalkDir(filepath.Join(repo.root, "memory-bank"), func(full string, e os.DirEntry, err error) error {
		if errors.Is(err, os.ErrNotExist) && full == filepath.Join(repo.root, "memory-bank") {
			return nil
		}
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(repo.root, full)
		p := filepath.ToSlash(rel)
		if e.IsDir() {
			if p == "memory-bank/.repo" {
				return filepath.SkipDir
			}
			return nil
		}
		if e.Type()&os.ModeSymlink != 0 || !e.Type().IsRegular() {
			return fmt.Errorf("unsafe concurrent component entry: %s", p)
		}
		if o, ok := expected[p]; !ok || !o.Exists {
			return fmt.Errorf("unobserved component file appeared: %s", p)
		}
		return nil
	})
}

// Never discard the only recovery data until the complete restored state is
// verified and durable, including unchanged read inputs and ancestor modes.
func syncRestoredComponentState(repo pinnedRepo, j *componentJournal, staging string) error {
	if err := verifyComponentInventory(repo, j.Before); err != nil {
		return err
	}
	if err := verifyComponentDirectoryStates(repo, j, false); err != nil {
		return err
	}
	for _, p := range contracts.Keys(j.Before) {
		o, _, err := observeComponent(repo, p)
		if err != nil && strings.HasPrefix(j.Backups[p], "old/") {
			info, data, readErr := secureReadDestination(repo, p)
			backup, backupErr := os.Lstat(filepath.Join(staging, j.Backups[p]))
			if readErr == nil && backupErr == nil && os.SameFile(info, backup) && componentLinkCount(info) == 2 && info.Mode().IsRegular() && info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) == 0 {
				o = observation{true, digest(data), gitMode(info.Mode().Perm()), fmt.Sprintf("%04o", info.Mode().Perm())}
				err = nil
			}
		}
		if err != nil {
			return err
		}
		if !sameObservation(o, j.Before[p]) {
			return fmt.Errorf("restored input differs: %s", p)
		}
		if o.Exists {
			if err = syncComponentFile(repo, p); err != nil {
				return err
			}
		}
	}
	for _, p := range contracts.Keys(j.Directories) {
		if j.Directories[p].BeforeExists {
			if err := syncComponentDirectory(repo, p); err != nil {
				return err
			}
		}
	}
	return syncLocalDirectory(repo.root)
}
