package manifest

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// ValidateStruct recursively inspects validate tags on struct fields and
// returns the first validation error found. It automatically walks into
// nested struct and *struct fields, building up the YAML field path
// (e.g. "metadata.name", "spec.merge_strategy.squash_merge_commit_title").
// Slice fields are NOT recursed into — the caller handles element iteration.
//
// prefix is an optional contextual label prepended with ": " to error messages
// (e.g. repo name, "my-repo: rulesets[0]"). It is separate from the structural path.
//
// Supported tag syntax:
//
//	validate:"required"              — string non-empty or slice non-empty
//	validate:"oneof=a b c"           — value must be one of the listed values
//	validate:"omitempty,oneof=a b c" — skip if zero/nil, otherwise check oneof
//	validate:"unique=Name"           — slice elements must have unique values for the given field
//	validate:"exclusive=Source"       — this field and the named field cannot both be non-empty
func ValidateStruct(prefix string, v any) error {
	return validateStructRecursive(prefix, "", v)
}

// validateStructRecursive walks the struct fields, accumulating structPath
// (dot-separated YAML path) for error messages.
func validateStructRecursive(prefix, structPath string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}
	rt := rv.Type()

	for i := range rt.NumField() {
		field := rt.Field(i)
		fv := rv.Field(i)

		// Skip fields not part of YAML schema
		if field.Tag.Get("yaml") == "-" {
			continue
		}

		fieldPath := joinDot(structPath, yamlFieldName(field))

		// Process validate tag
		if tag := field.Tag.Get("validate"); tag != "" && tag != "-" {
			if err := validateField(fieldPath, fv, tag, rv); err != nil {
				return wrapPrefix(prefix, err)
			}
		}

		// Recurse into nested struct / *struct (not slices)
		switch fv.Kind() {
		case reflect.Struct:
			if err := validateStructRecursive(prefix, fieldPath, fv.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Ptr:
			if !fv.IsNil() && fv.Elem().Kind() == reflect.Struct {
				if err := validateStructRecursive(prefix, fieldPath, fv.Interface()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validateField checks a single field value against its validate tag.
// parentStruct is needed for exclusive checks (to look up sibling fields).
func validateField(fieldPath string, fv reflect.Value, tag string, parentStruct ...reflect.Value) error {
	parts := strings.Split(tag, ",")
	omitempty := slices.Contains(parts, "omitempty")

	for _, p := range parts {
		switch {
		case p == "required":
			if err := checkRequired(fieldPath, fv); err != nil {
				return err
			}
		case strings.HasPrefix(p, "oneof="):
			if omitempty && isZero(fv) {
				continue
			}
			allowed := strings.Fields(strings.TrimPrefix(p, "oneof="))
			if err := checkOneOf(fieldPath, fv, allowed); err != nil {
				return err
			}
		case strings.HasPrefix(p, "unique="):
			keyField := strings.TrimPrefix(p, "unique=")
			if err := checkUnique(fieldPath, fv, keyField); err != nil {
				return err
			}
		case strings.HasPrefix(p, "exclusive="):
			otherGoName := strings.TrimPrefix(p, "exclusive=")
			if len(parentStruct) > 0 {
				if err := checkExclusive(fieldPath, fv, parentStruct[0], otherGoName); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// MigrateDeprecated processes deprecated tags on the direct fields of a struct.
// It migrates values to their target fields and returns collected warnings.
//
// Tag syntax:
//
//	deprecated:"TargetField:message"  — migrate value to TargetField, warn
//	deprecated:":message"             — no migration target, just warn and clear
//
// Returns an error if a deprecated field and its migration target are both set.
func MigrateDeprecated(v any) ([]string, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, nil
	}

	var warnings []string

	for i := range rv.NumField() {
		field := rv.Type().Field(i)
		tag := field.Tag.Get("deprecated")
		if tag == "" {
			continue
		}

		fv := rv.Field(i)
		// Only process non-empty string fields
		if fv.Kind() != reflect.String || fv.String() == "" {
			continue
		}

		target, msg := parseDeprecatedTag(tag)

		if target != "" {
			targetField := rv.FieldByName(target)
			if !targetField.IsValid() || !targetField.CanSet() {
				return nil, fmt.Errorf("deprecated tag references invalid field %q", target)
			}
			// Check conflict: both deprecated and target are set
			if targetField.Kind() == reflect.String && targetField.String() != "" {
				yamlOld := yamlFieldName(field)
				yamlNew := ""
				if tf, ok := rv.Type().FieldByName(target); ok {
					yamlNew = yamlFieldName(tf)
				}
				return nil, fmt.Errorf("cannot specify both %q and %q", yamlOld, yamlNew)
			}
			// Migrate value
			targetField.Set(fv)
		}

		warnings = append(warnings, fmt.Sprintf("%q is deprecated, %s", yamlFieldName(field), msg))
		// Clear the deprecated field
		fv.SetString("")
	}

	return warnings, nil
}

// parseDeprecatedTag splits "TargetField:message" into target and message.
func parseDeprecatedTag(tag string) (target, msg string) {
	target, msg, found := strings.Cut(tag, ":")
	if !found {
		return "", tag
	}
	return target, msg
}

// joinDot joins two path segments with ".". Returns name if prefix is empty.
func joinDot(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

// yamlFieldName extracts the YAML field name from struct tags, falling back
// to the Go field name.
func yamlFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("yaml")
	if tag == "" || tag == "-" {
		return f.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}

// checkRequired verifies a field is non-empty.
func checkRequired(name string, fv reflect.Value) error {
	switch fv.Kind() {
	case reflect.String:
		if fv.String() == "" {
			return fmt.Errorf("%s is required", name)
		}
	case reflect.Slice:
		if fv.IsNil() || fv.Len() == 0 {
			return fmt.Errorf("%s is required", name)
		}
	case reflect.Ptr:
		if fv.IsNil() {
			return fmt.Errorf("%s is required", name)
		}
	}
	return nil
}

// checkOneOf verifies the field value is one of the allowed values.
func checkOneOf(name string, fv reflect.Value, allowed []string) error {
	var val string
	switch fv.Kind() {
	case reflect.String:
		val = fv.String()
	case reflect.Ptr:
		if fv.IsNil() {
			return nil
		}
		val = fv.Elem().String()
	default:
		return nil
	}
	if slices.Contains(allowed, val) {
		return nil
	}
	return fmt.Errorf("invalid %s %q (must be one of: %s)", name, val, strings.Join(allowed, ", "))
}

// checkUnique verifies that elements of a slice have unique values for the given field.
// e.g. validate:"unique=Name" on a []Secret checks that no two secrets share the same Name.
func checkUnique(fieldPath string, fv reflect.Value, keyField string) error {
	if fv.Kind() != reflect.Slice || fv.IsNil() {
		return nil
	}
	seen := make(map[string]bool)
	for i := range fv.Len() {
		elem := fv.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		if elem.Kind() != reflect.Struct {
			continue
		}
		kf := elem.FieldByName(keyField)
		if !kf.IsValid() || kf.Kind() != reflect.String {
			continue
		}
		val := kf.String()
		if val == "" {
			continue
		}
		if seen[val] {
			return fmt.Errorf("duplicate %s %q in %s", strings.ToLower(keyField), val, fieldPath)
		}
		seen[val] = true
	}
	return nil
}

// checkExclusive verifies that this field and another named field are not both non-empty.
// e.g. validate:"exclusive=Source" on Content checks that content and source aren't both set.
func checkExclusive(fieldPath string, fv reflect.Value, parent reflect.Value, otherGoName string) error {
	if isZero(fv) {
		return nil
	}
	other := parent.FieldByName(otherGoName)
	if !other.IsValid() || isZero(other) {
		return nil
	}
	// Look up yaml name of the other field for the error message
	otherYAML := otherGoName
	if parent.Kind() == reflect.Struct {
		if sf, ok := parent.Type().FieldByName(otherGoName); ok {
			otherYAML = yamlFieldName(sf)
		}
	}
	// Extract the field name from the path (last segment)
	thisYAML := fieldPath
	if idx := strings.LastIndex(fieldPath, "."); idx >= 0 {
		thisYAML = fieldPath[idx+1:]
	}
	return fmt.Errorf("cannot have both %s and %s", thisYAML, otherYAML)
}

// isZero checks if a value is the zero value for its type.
func isZero(fv reflect.Value) bool {
	switch fv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return fv.IsNil()
	case reflect.String:
		return fv.String() == ""
	case reflect.Slice:
		return fv.IsNil() || fv.Len() == 0
	default:
		return false
	}
}

// wrapPrefix prepends a context prefix to an error.
func wrapPrefix(prefix string, err error) error {
	if prefix == "" {
		return err
	}
	return fmt.Errorf("%s: %w", prefix, err)
}
