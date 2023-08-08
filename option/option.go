package option

// Optional is a type alias for optional values.
//
// # Compatibility Note
//
// In the future, this type will be replaced with an opaque type. This means
// that the type might not be a pointer anymore. Because of this, it is
// recommended to use the methods provided by this package to interact with
// this type.
type Optional[T any] *optionalImpl[T]

type optionalImpl[T any] struct{ v T }

// Some returns an optional value from a non-nil value.
func Some[T any](v T) Optional[T] { return Optional[T](&optionalImpl[T]{v}) }

// None returns a nil optional value.
func None[T any]() Optional[T] { return nil }

// PtrTo returns a pointer to the given value.
func PtrTo[T any](v T) *T { return &v }

// Null is a type alias for a struct that only contains a null value.
// Only use nil for this type.
type Null *struct{}
