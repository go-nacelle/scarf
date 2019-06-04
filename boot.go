package scarf

import (
	"github.com/go-nacelle/nacelle"

	"github.com/go-nacelle/scarf/logging"
)

// BootAndExit creates a nacelle Bootstrapper with the given name and
// initializes and starts a gRPC server with the given endpoint set. This
// method does not return.
func BootAndExit(name string, endpointSet EndpointSet) {
	boostrapper := nacelle.NewBootstrapper(
		name,
		setupFactory(endpointSet),
	)

	boostrapper.BootAndExit()
}

func setupFactory(endpointSet EndpointSet) func(nacelle.ProcessContainer, nacelle.ServiceContainer) error {
	return func(processes nacelle.ProcessContainer, services nacelle.ServiceContainer) error {
		processes.RegisterInitializer(
			logging.NewInitializer(DefaultExtractors),
			nacelle.WithInitializerName("log decorator"),
		)

		processes.RegisterProcess(
			NewServer(endpointSet),
			nacelle.WithProcessName("server"),
		)

		return nil
	}
}
