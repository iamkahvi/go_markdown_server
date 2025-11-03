package diff

func ConstructDocString(patches []Patch) string {
	var doc_string string
	for _, patch := range patches {
		switch patch.Type {
		case Add:
			doc_string += patch.Value
		case None:
			doc_string += patch.Value
		}
	}
	return doc_string
}
