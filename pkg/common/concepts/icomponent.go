package concepts

type IComponent interface {
	Name() string
	Priority() int32
	SetEngine(engine IEngine)
	GetEngine() IEngine
	OnInit() error
	OnStart()
	OnCleanup()
}
