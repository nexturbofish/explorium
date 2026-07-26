package reflect

import (
	"context"
	"hi/spec"
)

type Reflector struct {
	config ReflectConfig
}

func NewReflector(config ReflectConfig) *Reflector {
	return &Reflector{config: config}
}

func (r *Reflector) Reflect(ctx context.Context, session *spec.Session, opts ReflectOptions) (*spec.ReflectionOutput, error) {
	return Reflect(ctx, session, r.config, opts)
}

func (r *Reflector) CompileProfile(ctx context.Context) error {
	_, err := CompileProfile(ctx, r.config.MemoryStore)
	return err
}

func (r *Reflector) CompilePalaceIndex(ctx context.Context) error {
	_, err := CompilePalaceIndex(ctx, r.config.MemoryStore)
	return err
}
