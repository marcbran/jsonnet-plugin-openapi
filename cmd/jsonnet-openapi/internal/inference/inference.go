package inference

import "context"

type SpecDocument struct {
	JSON string
}

type SpecLoader interface {
	LoadSpec(ctx context.Context, ref string) (SpecDocument, error)
}

type Bundle struct {
	ID    string
	Input []byte
}

type BundleRenderer interface {
	RenderBundles(template string, specJSON string, previousJSON string) ([]Bundle, error)
}

type Task struct {
	JobName      string
	ID           string
	Input        []byte
	Prompt       string
	OutputSchema []byte
}

type Job interface {
	Name() string
	Build(ctx context.Context, spec SpecDocument, previous Results) ([]Task, error)
}

type Results map[string]string

type Runner interface {
	Exec(ctx context.Context, task Task) ([]byte, error)
}

type Store interface {
	Load(jobName string, taskID string) ([]byte, bool, error)
	Save(jobName string, taskID string, output []byte) error
	LoadAll(jobName string) (string, error)
}

type Pipeline struct {
	Jobs   []Job
	Runner Runner
	Store  Store
	Force  bool
	Limit  int
}

func (p Pipeline) Exec(ctx context.Context, spec SpecDocument) (Results, error) {
	results := Results{}
	for _, job := range p.Jobs {
		tasks, err := job.Build(ctx, spec, results)
		if err != nil {
			return nil, err
		}
		count := 0
		for _, task := range tasks {
			if !p.Force {
				_, ok, err := p.Store.Load(job.Name(), task.ID)
				if err != nil {
					return nil, err
				}
				if ok {
					continue
				}
			}
			output, err := p.Runner.Exec(ctx, task)
			if err != nil {
				return nil, err
			}
			err = p.Store.Save(job.Name(), task.ID, output)
			if err != nil {
				return nil, err
			}
			count++
			if p.Limit > 0 && count >= p.Limit {
				break
			}
		}
		all, err := p.Store.LoadAll(job.Name())
		if err != nil {
			return nil, err
		}
		results[job.Name()] = all
	}
	return results, nil
}
