package contracts

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
	"io"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

func Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func ValidDigest(s string) bool {
	if len(s) != 71 || !strings.HasPrefix(s, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(s[7:])
	return err == nil && strings.ToLower(s) == s
}
func ValidPath(s string) bool {
	if s == "" || s == "." || path.IsAbs(s) || path.Clean(s) != s || strings.HasPrefix(s, "../") || strings.ContainsAny(s, "\\\x00") || !utf8.ValidString(s) {
		return false
	}
	for _, part := range strings.Split(s, "/") {
		if strings.EqualFold(part, ".git") || strings.ContainsAny(part, `<>:"|?*`) || strings.HasSuffix(part, ".") || strings.HasSuffix(part, " ") {
			return false
		}
		for _, r := range part {
			if r < 32 || r == 127 {
				return false
			}
		}
		stem := strings.ToUpper(strings.TrimRight(strings.SplitN(part, ".", 2)[0], ". "))
		if reservedName.MatchString(stem) {
			return false
		}
	}
	return true
}

var reservedName = regexp.MustCompile(`^(CON|PRN|AUX|NUL|CONIN\$|CONOUT\$|COM[1-9¹²³]|LPT[1-9¹²³])$`)

func PortableKey(s string) string { return norm.NFC.String(cases.Fold().String(norm.NFC.String(s))) }
func CheckPortable(paths []string) error {
	seen := map[string]string{}
	for _, p := range paths {
		if !ValidPath(p) {
			return fmt.Errorf("unsafe path %q", p)
		}
		parts := strings.Split(p, "/")
		for i := range parts {
			segment := strings.Join(parts[:i+1], "/")
			key := PortableKey(segment)
			if previous, ok := seen[key]; ok && previous != segment {
				return fmt.Errorf("portable path collision: %s and %s", previous, segment)
			}
			seen[key] = segment
		}
	}
	return nil
}
func SortedSet(values []string) bool {
	for i, s := range values {
		if s == "" || (i > 0 && values[i-1] >= s) {
			return false
		}
	}
	return true
}
func Keys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
func FramedID(domain string, values ...string) string {
	h := sha256.New()
	h.Write([]byte(domain))
	h.Write([]byte{0})
	for _, v := range values {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(v)))
		h.Write(length[:])
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Decode rejects duplicate and case-aliased fields, null typed values, unknown
// fields, trailing values and invalid UTF-8 before typed JSON decoding.
func Decode(data []byte, target any) error {
	if !utf8.Valid(data) {
		return errors.New("JSON must be UTF-8")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	value, err := readValue(d)
	if err != nil {
		return err
	}
	if _, err = d.Token(); err != io.EOF {
		return errors.New("trailing JSON data")
	}
	typ := reflect.TypeOf(target)
	if typ == nil || typ.Kind() != reflect.Pointer {
		return errors.New("decode target must be a pointer")
	}
	if err = checkShape(value, typ.Elem()); err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
func readValue(d *json.Decoder) (any, error) {
	token, err := d.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}
	switch delim {
	case '{':
		m := map[string]any{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return nil, err
			}
			s, ok := key.(string)
			if !ok {
				return nil, errors.New("invalid object key")
			}
			if _, ok = m[s]; ok {
				return nil, fmt.Errorf("duplicate JSON field %q", s)
			}
			v, err := readValue(d)
			if err != nil {
				return nil, err
			}
			m[s] = v
		}
		_, err = d.Token()
		return m, err
	case '[':
		a := []any{}
		for d.More() {
			v, err := readValue(d)
			if err != nil {
				return nil, err
			}
			a = append(a, v)
		}
		_, err = d.Token()
		return a, err
	}
	return nil, errors.New("unexpected JSON delimiter")
}
func checkShape(v any, t reflect.Type) error {
	if t.Kind() == reflect.Pointer {
		return checkShape(v, t.Elem())
	}
	if v == nil {
		return errors.New("null is not a typed value")
	}
	if reflect.PointerTo(t).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
		obj, ok := v.(map[string]any)
		if !ok {
			return errors.New("expected object")
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name != "-" && f.IsExported() {
				if name == "" {
					name = f.Name
				}
				fields[name] = f.Type
				if !strings.Contains(f.Tag.Get("json"), ",omitempty") {
					if _, exists := obj[name]; !exists {
						return fmt.Errorf("missing JSON field %q", name)
					}
				}
			}
		}
		for key, value := range obj {
			ft, ok := fields[key]
			if !ok {
				return fmt.Errorf("unknown JSON field %q", key)
			}
			if err := checkShape(value, ft); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	case reflect.Map:
		obj, ok := v.(map[string]any)
		if !ok {
			return errors.New("expected map")
		}
		for _, value := range obj {
			if err := checkShape(value, t.Elem()); err != nil {
				return err
			}
		}
	case reflect.Slice:
		a, ok := v.([]any)
		if !ok {
			return errors.New("expected array")
		}
		for _, value := range a {
			if err := checkShape(value, t.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Canonical sorts keys even for structs and retains integer precision.
func Canonical(v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var obj any
	if err = d.Decode(&obj); err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}
