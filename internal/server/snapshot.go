package server

// snapshotSlice copies src into dst, growing dst's backing array when needed,
// and returns the slice to iterate. Callers must be the only users of dst
// between calls (the server main loop keeps one buffer per list).
func snapshotSlice[T any](dst, src []T) []T {
	if cap(dst) < len(src) {
		dst = make([]T, len(src))
	}
	dst = dst[:len(src)]
	copy(dst, src)
	return dst
}