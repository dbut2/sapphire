package hooks

type HookService[T ~int, V any] map[T]func(V)

func (h *HookService[T, V]) RegisterHook(hook T, f func(V)) {
	if *h == nil {
		*h = make(HookService[T, V])
	}
	(*h)[hook] = f
}

func (h *HookService[T, V]) Hook(hook T, v V) {
	if *h == nil {
		*h = make(HookService[T, V])
	}
	if f, ok := (*h)[hook]; ok {
		f(v)
	}
}
