package command

type Registry struct {
	commands map[string]Command
}

func NewRegistry(commands ...Command) *Registry {
	r := &Registry{commands: map[string]Command{}}
	for _, c := range commands {
		r.Register(c)
	}
	return r
}

// Register adds c and replaces a command with the same name.
func (r *Registry) Register(c Command) {
	r.commands[c.Name] = c
}

func (r *Registry) Lookup(name string) (Command, bool) {
	c, ok := r.commands[name]
	return c, ok
}
