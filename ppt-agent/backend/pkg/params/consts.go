package params

import (
	"context"
	"fmt"
)

type contextParams struct{}

var contextParamsKey = &contextParams{}

type typedContextParams struct{}

var typedContextParamsKey = &typedContextParams{}

const (
	FilePathSessionKey            = "file_path_session_key"
	WorkDirSessionKey             = "work_dir_session_key"
	UserAllPreviewFilesSessionKey = "user_all_preview_files_session_key"
	TaskIDKey                    = "task_id"
)

func InitContextParams(ctx context.Context) context.Context {
	return ctx
}

func AppendContextParams(ctx context.Context, params map[string]interface{}) context.Context {
	current := GetContextParams(ctx)
	for k, v := range params {
		current[k] = v
	}
	return context.WithValue(ctx, contextParamsKey, current)
}

func GetContextParams(ctx context.Context) map[string]interface{} {
	v := ctx.Value(contextParamsKey)
	if v == nil {
		return make(map[string]interface{})
	}
	return v.(map[string]interface{})
}

func GetTypedContextParams[T any](ctx context.Context, key string) (T, bool) {
	v := ctx.Value(typedContextParamsKey)
	if v == nil {
		var zero T
		return zero, false
	}

	params, ok := v.(map[string]interface{})
	if !ok {
		var zero T
		return zero, false
	}

	val, ok := params[key]
	if !ok {
		var zero T
		return zero, false
	}

	result, ok := val.(T)
	if !ok {
		var zero T
		return zero, false
	}

	return result, true
}

func SetTypedContextParams[T any](ctx context.Context, key string, value T) context.Context {
	v := ctx.Value(typedContextParamsKey)
	var params map[string]interface{}
	if v == nil {
		params = make(map[string]interface{})
	} else {
		var ok bool
		params, ok = v.(map[string]interface{})
		if !ok {
			params = make(map[string]interface{})
		}
	}
	params[key] = value
	return context.WithValue(ctx, typedContextParamsKey, params)
}

func MustGetContextParams(ctx context.Context, key string) interface{} {
	v := GetContextParams(ctx)[key]
	if v == nil {
		panic(fmt.Errorf("context param %s not found", key))
	}
	return v
}
