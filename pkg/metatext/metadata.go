package metatext

import (
	"strings"

	"google.golang.org/grpc/metadata"
)

/*
opentracing中
func (Tracer) Extract(format interface{}, carrier interface{}) (SpanContext, error)
func (Tracer) Inject(sm SpanContext, format interface{}, carrier interface{}) error
MetadataTextMap要实现carrier接口
type carrier interface {
	Set(key string, val string)
	ForeachKey(handler func(key string, val string) error) error
}
*/

type MetadataTextMap struct {
	metadata.MD
}

func (m MetadataTextMap) ForeachKey(handler func(key, value string) error) error {
	for k, mv := range m.MD {
		for _, v := range mv {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m MetadataTextMap) Set(key, value string) {
	key = strings.ToLower(key)
	m.MD[key] = append(m.MD[key], value)
}
