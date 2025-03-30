package coldbrew

// Captures 'raw' client inputs, and passes them to the client 'receiver' for mapping/tracking
type InputCapturer interface {
	Capture()
}
