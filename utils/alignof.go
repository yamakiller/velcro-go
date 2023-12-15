package utils

func AlignOf(size, align uint32) uint32 {
	return (size + align - 1) & ^(align - 1)
}
