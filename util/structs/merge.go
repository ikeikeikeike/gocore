package structs

import (
	"reflect"
	"time"

	"github.com/ikeikeikeike/gocore/util/logger"
	"github.com/imdario/mergo"

	"gopkg.in/oleiade/reflections.v1"
)

type (
	timeTransfomer struct{}
)

// Transformer merges time except src's ZeroTime: https://github.com/imdario/mergo#transformers
//
// If you mind even a bit, you would be better to use
// https://github.com/jinzhu/copier which will be overwritten everything.
//
func (t timeTransfomer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(time.Time{}) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				isZero := dst.MethodByName("IsZero")
				result := isZero.Call([]reflect.Value{})
				if result[0].Bool() {
					dst.Set(src)
				}
			}
			return nil
		}
	}
	return nil
}

// Merge merges dest values
func Merge(dest interface{}, values ...interface{}) error {
	data := make(map[string]interface{})

	for _, value := range values {
		v, _ := reflections.Items(value)
		if err := mergo.Map(&data, v, mergo.WithTransformers(timeTransfomer{})); err != nil {
			logger.E("merge.go Merge: %s", err)
		}
	}

	return mergo.Map(dest, data, mergo.WithTransformers(timeTransfomer{}))
}

// OverwriteMerge merges dest values
func OverwriteMerge(dest interface{}, values ...interface{}) error {
	data := make(map[string]interface{})

	for _, value := range values {
		v, _ := reflections.Items(value)
		if err := mergo.MapWithOverwrite(&data, v, mergo.WithTransformers(timeTransfomer{})); err != nil {
			logger.E("merge.go OverwriteMerge: %s", err)
		}
	}

	return mergo.MapWithOverwrite(dest, data, mergo.WithTransformers(timeTransfomer{}))
}
