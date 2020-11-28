package github

func unboxString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func unboxBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func boxString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
