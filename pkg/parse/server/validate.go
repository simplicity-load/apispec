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
		if (r < 'a' || r > 'z') &&
			r != '-' && r != '_' {
			if r != '-' {
				return false
			}
		}
	}
	return len(s) > 0
}
