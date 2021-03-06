package generator

import (
	"errors"
	"fmt"

	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/generator/write_strategy"
)

const Version = template.Version

var (
	EmptyTemplateError = errors.New("empty template")
	EmptyStrategyError = errors.New("empty strategy")
)

type Generator interface {
	Generate() error
}

type generationUnit struct {
	template template.Template

	writeStrategy write_strategy.Strategy
	absOutPath    string
}

func NewGenUnit(tmpl template.Template, outPath string) (*generationUnit, error) {
	err := tmpl.Prepare()
	if err != nil {
		return nil, fmt.Errorf("%s: prepare error: %v", tmpl.DefaultPath(), err)
	}
	strategy, err := tmpl.ChooseStrategy()
	if err != nil {
		return nil, err
	}
	return &generationUnit{
		template:      tmpl,
		absOutPath:    outPath,
		writeStrategy: strategy,
	}, nil
}

func (g *generationUnit) Generate() error {
	if g.template == nil {
		return EmptyTemplateError
	}
	if g.writeStrategy == nil {
		return EmptyStrategyError
	}
	code := g.template.Render()
	err := g.writeStrategy.Write(code)
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	return nil
}
