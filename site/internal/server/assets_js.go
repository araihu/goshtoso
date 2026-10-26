//go:build js && wasm

package server

// The ASSETS binding serves the exported files, including module assets. Avoid
// linking their handlers and embedded images into the server WASM module.
func (s *Server) setupAssetRoutes()  {}
func (s *Server) registerDogImages() {}
