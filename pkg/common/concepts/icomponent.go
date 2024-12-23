package concepts

type IComponent interface {
	IActor
	Name() string
	Priority() int32
	OnInit() error
	OnCleanup()
}
