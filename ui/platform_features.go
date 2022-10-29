package ui

type platformFeature int

const (
	supportsQRScanning platformFeature = 1 << iota
)
