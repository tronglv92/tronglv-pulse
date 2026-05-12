package filex

type FileEntity interface {
	GetName() string
	GetData() []byte
	GetSize() int64
	FlushData()
}

type fileItem struct {
	name string
	data []byte
	size int64
}

func NewFile(name string, data []byte) FileEntity {
	return &fileItem{
		name: name,
		data: data,
		size: int64(len(data)),
	}
}

func (f *fileItem) GetName() string {
	return f.name
}

func (f *fileItem) GetData() []byte {
	return f.data
}

func (f *fileItem) GetSize() int64 {
	return f.size
}

func (f *fileItem) FlushData() {
	f.data = nil
}
