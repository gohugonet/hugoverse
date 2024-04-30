package resources

type Resource interface {
	Get(filename any) (resources.Resource, error)
}
