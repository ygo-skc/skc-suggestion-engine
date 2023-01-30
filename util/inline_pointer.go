package util

// transforms string literal to string pointer, without this function an inline string pointer cannot be created
func InlineStringPointer(s string) *string {
	return &s
}

// transforms unsigned 16 byte int literal to unsigned 16 byte int pointer, without this function an inline unsigned 16 byte int pointer cannot be created
func InlineUInt16Pointer(i uint16) *uint16 {
	return &i
}
