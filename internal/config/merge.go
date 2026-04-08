package config

// Merge combines global and project configs. Project overrides global for same names.
// Returns the merged config and any conflicts found.
func Merge(global, project *Config) (*Config, []Conflict) {
	if global == nil && project == nil {
		return &Config{
			Version:   "0.1.0",
			Aliases:   make(map[string]Alias),
			Workflows: make(map[string]Workflow),
		}, nil
	}

	if project == nil {
		return global, nil
	}

	if global == nil {
		return project, nil
	}

	merged := &Config{
		Version:   global.Version,
		Global:    global.Global,
		Aliases:   make(map[string]Alias),
		Workflows: make(map[string]Workflow),
	}

	var conflicts []Conflict

	// Copy global aliases
	for name, alias := range global.Aliases {
		merged.Aliases[name] = alias
	}

	// Override/add project aliases
	for name, alias := range project.Aliases {
		if globalAlias, exists := global.Aliases[name]; exists {
			conflicts = append(conflicts, Conflict{
				Name:       name,
				Type:       "alias",
				GlobalVal:  globalAlias.Cmd,
				ProjectVal: alias.Cmd,
			})
		}
		merged.Aliases[name] = alias // project wins
	}

	// Copy global workflows
	for name, wf := range global.Workflows {
		merged.Workflows[name] = wf
	}

	// Override/add project workflows
	for name, wf := range project.Workflows {
		if globalWf, exists := global.Workflows[name]; exists {
			conflicts = append(conflicts, Conflict{
				Name:       name,
				Type:       "workflow",
				GlobalVal:  globalWf.Description,
				ProjectVal: wf.Description,
			})
		}
		merged.Workflows[name] = wf // project wins
	}

	return merged, conflicts
}
