package consts

const (
	// TASK 前綴
	TASK = "task"
	// FAILED 後綴
	FAILED = "failed"
	// RETRY 後綴
	RETRY = "retry"
	// RoutingKey is fuzzy find task.xxx or tash.xxx.xxx,etc.
	RoutingKey = "task.#"
)
