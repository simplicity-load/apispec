package server

func isAllLowerAZ(s string) bool {
	for _, r := range s {
		if r < 'a' || r > 'z' {
			return false
		}
	}
	return len(s) > 0
}

func isAllLowerA_Z(s string) bool {
	for _, r := range s {
		if ('a' <= r && r <= 'z') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return len(s) > 0
}
