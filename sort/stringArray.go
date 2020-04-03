package sort

// StringArray wraps a slice of strings to allow function calls for convenience
type StringArray []string

// String is used for parsing an array of flags
func (s *StringArray) String() string {
	output := ""
	for i := 0; i < len(*s); i++ {
		output = output + " " + (*s)[i]
	}
	return output
}

// Set is used for parsing an array of flags
func (s *StringArray) Set(value string) error {
	*s = append(*s, value)
	return nil
}
