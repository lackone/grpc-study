package metatext

import (
	"google.golang.org/grpc/metadata"
	"strings"
)

type MetadataTextMap struct {
	metadata.MD
}

func (m MetadataTextMap) Get(key string) string {
	if vs, ok := m.MD[key]; ok {
		return vs[0]
	}
	return ""
}

func (m MetadataTextMap) Set(key string, value string) {
	key = strings.ToLower(key)
	m.MD.Append(key, value)
}

func (m MetadataTextMap) Keys() []string {
	keys := make([]string, 0)
	for k, _ := range m.MD {
		keys = append(keys, k)
	}
	return keys
}
