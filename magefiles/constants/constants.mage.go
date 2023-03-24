package constants

// Since we are dealing with builds, having a constants file until using a config input makes it easy.

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build.

const (
	// ArtifactDirectory is a directory containing artifacts for the project and shouldn't be committed to source.
	ArtifactDirectory = ".artifacts"

	// PermissionUserReadWriteExecute is the permissions for the artifact directory.
	PermissionUserReadWriteExecute = 0o0700

	// Config directory is where the cert config and other automation task defaults are configured.
	ConfigDirectory = "config"

	// CacheDirectory is where the cache for the project is placed, ie artifacts that don't need to be rebuilt often.
	CacheDirectory = ".cache"

	// ExamplesDirectory is the directory where the kubernetes manifests are stored.
	ExamplesDirectory = "k8s"

	// CacheManifestDirectory is the directory where the cached k8 manifests are copied to.
	CacheManifestDirectory = ".cache/manifests"
)

const (
	// KindClusterName is the name of the kind cluster.
	KindClusterName = "dsvtest"
	// KindClusterName is the name of the kind cluster.
	KindContextName = "kind-dsvtest"
	// KubeconfigPath is the path to the kubeconfig file for this project, which is cached locally.
	Kubeconfig = ".cache/config"
	// KubectlNamespace is the namespace used for all kubectl commands, so that they don't operate in default or other namespace by accident.
	KubectlNamespace = "dsv"
)

const (

	// ChartsDirectory is the directory where the helm charts are placed, in sub directories.
	ChartsDirectory = "charts"
)
