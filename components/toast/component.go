package toast

import (
	"context"
	"io"

	"github.com/araihu/goshtoso/components"
)

// ContainerInstance is a renderable toast container.
type ContainerInstance struct {
	cfg ContainerConfig
}

// ToastContainer returns a renderable toast container.
func ToastContainer(cfg ContainerConfig) ContainerInstance {
	return ContainerInstance{cfg: cfg}
}

// Kind identifies the component as a toast container.
func (ContainerInstance) Kind() components.Kind {
	return components.KindToastContainer
}

// Render writes the toast container markup.
func (i ContainerInstance) Render(ctx context.Context, w io.Writer) error {
	return toastContainerTemplate(i.cfg).Render(ctx, w)
}

// Instance is a renderable toast component.
type Instance struct {
	cfg Config
}

// Toast returns a renderable toast component.
func Toast(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a toast.
func (Instance) Kind() components.Kind {
	return components.KindToast
}

// Render writes the toast markup.
func (i Instance) Render(ctx context.Context, w io.Writer) error {
	return toastTemplate(i.cfg).Render(ctx, w)
}

// MessageInstance is a renderable message toast.
type MessageInstance struct {
	cfg MessageConfig
}

// MessageToast returns a renderable message toast.
func MessageToast(cfg MessageConfig) MessageInstance {
	return MessageInstance{cfg: cfg}
}

// Kind identifies the component as a message toast.
func (MessageInstance) Kind() components.Kind {
	return components.KindMessageToast
}

// Render writes the message toast markup.
func (i MessageInstance) Render(ctx context.Context, w io.Writer) error {
	return messageToastTemplate(i.cfg).Render(ctx, w)
}

// OOBInstance is a renderable out-of-band toast.
type OOBInstance struct {
	cfg Config
}

// OOBToast returns a renderable out-of-band toast.
func OOBToast(cfg Config) OOBInstance {
	return OOBInstance{cfg: cfg}
}

// Kind identifies the component as an out-of-band toast.
func (OOBInstance) Kind() components.Kind {
	return components.KindOOBToast
}

// Render writes the out-of-band toast markup.
func (i OOBInstance) Render(ctx context.Context, w io.Writer) error {
	return oobToastTemplate(i.cfg).Render(ctx, w)
}

// OOBMessageInstance is a renderable out-of-band message toast.
type OOBMessageInstance struct {
	cfg MessageConfig
}

// OOBMessageToast returns a renderable out-of-band message toast.
func OOBMessageToast(cfg MessageConfig) OOBMessageInstance {
	return OOBMessageInstance{cfg: cfg}
}

// Kind identifies the component as an out-of-band message toast.
func (OOBMessageInstance) Kind() components.Kind {
	return components.KindOOBMessageToast
}

// Render writes the out-of-band message toast markup.
func (i OOBMessageInstance) Render(ctx context.Context, w io.Writer) error {
	return oobMessageToastTemplate(i.cfg).Render(ctx, w)
}

var (
	_ components.Component = ContainerInstance{}
	_ components.Component = Instance{}
	_ components.Component = MessageInstance{}
	_ components.Component = OOBInstance{}
	_ components.Component = OOBMessageInstance{}
)
