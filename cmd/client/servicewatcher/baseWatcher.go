package servicewatcher

// Watcher base interface for all our watcher implementations
type Watcher interface {
	watch()
}
