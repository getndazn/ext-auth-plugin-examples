package main

import (
	impl "github.com/getndazn/ext-auth-plugin-examples/plugins/geofencing/pkg"
	"github.com/solo-io/ext-auth-plugins/api"
)

func main() {}

// Compile-time assertion
var _ api.ExtAuthPlugin = new(impl.GeoFencingPlugin)

// This is the exported symbol that Gloo will look for.
//noinspection GoUnusedGlobalVariable
var Plugin impl.GeoFencingPlugin
