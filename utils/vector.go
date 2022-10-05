package utils

type VectorInt struct {
	Begin Int2
	End   Int2
}

func NewVectorInt(begin, end Int2) VectorInt {
	return VectorInt{begin, end}
}
