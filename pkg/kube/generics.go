package kube

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ObjectList is a generic equivalent of [ObjectList].
type ObjectList[T any] interface {
	client.ObjectList
	*T
}

// AddrofObject is a helper type constraint for accepting a struct value that
// implements the Object interface.
type AddrofObject[T any] interface {
	*T
	client.Object
}

// List is a generic equivalent of [Ctl.List].
func List[T any, L ObjectList[T]](ctx context.Context, ctl *Ctl, opts ...client.ListOption) (*T, error) {
	var list T
	if err := ctl.client.List(ctx, L(&list), opts...); err != nil {
		return nil, err
	}
	return &list, nil
}

// Get is a generic equivalent of [Ctl.Get].
func Get[T any, PT AddrofObject[T]](ctx context.Context, ctl *Ctl, key ObjectKey) (*T, error) {
	var obj T
	if err := ctl.client.Get(ctx, key, PT(&obj)); err != nil {
		return nil, err
	}
	return &obj, nil
}

// Get is a generic equivalent of [Ctl.Create].
func Create[T any, PT AddrofObject[T]](ctx context.Context, ctl *Ctl, obj T) (*T, error) {
	if err := ctl.Create(ctx, PT(&obj)); err != nil {
		return nil, err
	}
	return &obj, nil
}

// Get is a generic equivalent of [Ctl.Delete].
func Delete[T any, PT AddrofObject[T]](ctx context.Context, ctl *Ctl, key ObjectKey) error {
	obj := PT(new(T))
	obj.SetName(key.Name)
	obj.SetNamespace(key.Namespace)

	return ctl.client.Delete(ctx, obj)
}

func AsKey(obj Object) ObjectKey {
	return ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}
