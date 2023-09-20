package kvengine

type Value struct{
	Key string
	Value []byte
	Deleted bool
}
