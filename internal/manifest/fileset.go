package manifest

// UnmarshalYAML allows FileSetRepository to be either a string or a struct.
func (t *FileSetRepository) UnmarshalYAML(unmarshal func(any) error) error {
	// Try string first
	var s string
	if err := unmarshal(&s); err == nil {
		t.Name = s
		return nil
	}

	// Try struct
	type raw FileSetRepository
	var r raw
	if err := unmarshal(&r); err != nil {
		return err
	}
	*t = FileSetRepository(r)
	return nil
}
