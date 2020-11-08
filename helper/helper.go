package helper

// RemoveDuplicate removes duplicates entries in a list
func RemoveDuplicate(s []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	for _, i := range s {
		if _, present := keys[i]; present == false {
			keys[i] = true
			result = append(result, i)
		}
	}
	return result
}
