package form

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/expressions"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components"
)

// Instance is a renderable form component.
type Instance struct {
	expressionOverrides expressions.Form
	cfg                 Config
}

// Form returns a renderable form component.
func Form(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a form.
func (Instance) Kind() components.Kind {
	return components.KindForm
}

// Render writes the form markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Form: i.expressionOverrides})
	return formTemplate(i.cfg).Render(ctx, w)
}

// SectionInstance is a renderable form section component.
type SectionInstance struct {
	expressionOverrides expressions.Form
	cfg                 SectionConfig
}

// Section returns a renderable form section component.
func Section(cfg SectionConfig) SectionInstance {
	return SectionInstance{cfg: cfg}
}

// Kind identifies the component as a form section.
func (SectionInstance) Kind() components.Kind {
	return components.KindFormSection
}

// Render writes the form section markup.
func (i SectionInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Form: i.expressionOverrides})
	return sectionTemplate(i.cfg).Render(ctx, w)
}

// CollapsibleSectionInstance is a renderable collapsible form section.
type CollapsibleSectionInstance struct {
	expressionOverrides expressions.Form
	cfg                 CollapsibleSectionConfig
}

// CollapsibleSection returns a renderable collapsible form section.
func CollapsibleSection(cfg CollapsibleSectionConfig) CollapsibleSectionInstance {
	return CollapsibleSectionInstance{cfg: cfg}
}

// Kind identifies the component as a collapsible form section.
func (CollapsibleSectionInstance) Kind() components.Kind {
	return components.KindFormCollapsibleSection
}

// Render writes the collapsible form section markup.
func (i CollapsibleSectionInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Form: i.expressionOverrides})
	return collapsibleSectionTemplate(i.cfg).Render(ctx, w)
}

// FlipSectionInstance is a renderable flippable form section.
type FlipSectionInstance struct {
	expressionOverrides expressions.Form
	cfg                 FlipSectionConfig
	readView            templ.Component
}

// FlipSection returns a renderable flippable form section.
func FlipSection(cfg FlipSectionConfig, readView templ.Component) FlipSectionInstance {
	return FlipSectionInstance{cfg: cfg, readView: readView}
}

// Kind identifies the component as a flippable form section.
func (FlipSectionInstance) Kind() components.Kind {
	return components.KindFormFlipSection
}

// Render writes the flippable form section markup.
func (i FlipSectionInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Form: i.expressionOverrides})
	return flipSectionTemplate(i.cfg, i.readView).Render(ctx, w)
}

// SubSectionInstance is a renderable form subsection.
type SubSectionInstance struct {
	expressionOverrides expressions.Form
	cfg                 SubSectionConfig
}

// SubSection returns a renderable form subsection.
func SubSection(cfg SubSectionConfig) SubSectionInstance {
	return SubSectionInstance{cfg: cfg}
}

// Kind identifies the component as a form subsection.
func (SubSectionInstance) Kind() components.Kind {
	return components.KindFormSubSection
}

// Render writes the form subsection markup.
func (i SubSectionInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Form: i.expressionOverrides})
	return subSectionTemplate(i.cfg).Render(ctx, w)
}

// FieldGroupInstance is a renderable form field group.
type FieldGroupInstance struct {
	expressionOverrides expressions.Form
	cfg                 FieldGroupConfig
}

// FieldGroup returns a renderable form field group.
func FieldGroup(cfg FieldGroupConfig) FieldGroupInstance {
	return FieldGroupInstance{cfg: cfg}
}

// Kind identifies the component as a form field group.
func (FieldGroupInstance) Kind() components.Kind {
	return components.KindFormFieldGroup
}

// Render writes the form field group markup.
func (i FieldGroupInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Form: i.expressionOverrides})
	return fieldGroupTemplate(i.cfg).Render(ctx, w)
}

// FormErrorsInstance is a renderable form error summary.
type FormErrorsInstance struct {
	expressionOverrides expressions.Form
	cfg                 FormErrorsConfig
}

// FormErrors returns a renderable form error summary.
func FormErrors(cfg FormErrorsConfig) FormErrorsInstance {
	return FormErrorsInstance{cfg: cfg}
}

// Kind identifies the component as a form error summary.
func (FormErrorsInstance) Kind() components.Kind {
	return components.KindFormErrors
}

// Render writes the form error summary markup.
func (i FormErrorsInstance) Render(ctx context.Context, w io.Writer) error {
	ctx = expressions.With(ctx, expressions.Set{Form: i.expressionOverrides})
	return formErrorsTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = Instance{}
	_ components.Component = SectionInstance{}
	_ components.Component = CollapsibleSectionInstance{}
	_ components.Component = FlipSectionInstance{}
	_ components.Component = SubSectionInstance{}
	_ components.Component = FieldGroupInstance{}
	_ components.Component = FormErrorsInstance{}
)

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i Instance) WithExpressions(values expressions.Form) Instance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Form: i.expressionOverrides}, expressions.Set{Form: values}).Form
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i SectionInstance) WithExpressions(values expressions.Form) SectionInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Form: i.expressionOverrides}, expressions.Set{Form: values}).Form
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i CollapsibleSectionInstance) WithExpressions(values expressions.Form) CollapsibleSectionInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Form: i.expressionOverrides}, expressions.Set{Form: values}).Form
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i FlipSectionInstance) WithExpressions(values expressions.Form) FlipSectionInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Form: i.expressionOverrides}, expressions.Set{Form: values}).Form
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i SubSectionInstance) WithExpressions(values expressions.Form) SubSectionInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Form: i.expressionOverrides}, expressions.Set{Form: values}).Form
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i FieldGroupInstance) WithExpressions(values expressions.Form) FieldGroupInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Form: i.expressionOverrides}, expressions.Set{Form: values}).Form
	return i
}

// WithExpressions returns a copy with component-scoped text overrides. Existing
// explicit config labels take precedence. Empty fields inherit render defaults.
func (i FormErrorsInstance) WithExpressions(values expressions.Form) FormErrorsInstance {
	i.expressionOverrides = expressions.Merge(expressions.Set{Form: i.expressionOverrides}, expressions.Set{Form: values}).Form
	return i
}
