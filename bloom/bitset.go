package bloom

type BitSet interface {
	Add(items []int, opts ...Option) error
	Exists(items []int, opts ...Option) (bool, error)
	Reset(opts ...Option) error
}

func newBitSet(size int) BitSet {
	bitset := make(bitset, size)
	return &bitset
}

type bitset []bool

func (b *bitset) Add(items []int, opts ...Option) error {
	for _, item := range items {
		(*b)[item] = true
	}
	return nil
}

func (b *bitset) Exists(items []int, opts ...Option) (bool, error) {
	for _, item := range items {
		if !(*b)[item] {
			return false, nil
		}
	}
	return true, nil
}

func (b *bitset) Reset(opts ...Option) error {
	for i := range *b {
		(*b)[i] = false
	}
	return nil
}
