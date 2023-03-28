// Package cert uses cfssl to generate certs for usage in testing and setup.

package cert

import (
	// "bufio".
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"

	"github.com/bitfield/script"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// Cert contains tasks to generate cert.
type Cert mg.Namespace

type renameFile struct {
	OriginalName string
	NewName      string
}

// Generate certs using cffsl (cloudflare toolkit). Requires aqua to have installed already.
func (Cert) Generate() error {
	magetoolsutils.CheckPtermDebug()
	pterm.DefaultHeader.Println("Generate()")
	duration := time.Second * 1

	_ = os.MkdirAll(constants.CacheCertDirectory, constants.PermissionUserReadWriteExecute)
	// defer func() {
	// 	for _, item := range []string{
	// 		"ca-key.pem",
	// 		"ca.csr",
	// 		"ca.pem",
	// 		"auth-key.pem",
	// 		"auth.csr",
	// 		"auth.pem",
	// 	} {
	// 		targetFile := filepath.Join(constants.CacheDirectory, item)
	// 		if err := os.Rename(item, targetFile); err != nil {
	// 			pterm.Warning.Printfln("unable to rename file: %v", item)
	// 		}
	// 		pterm.Info.Printfln("mv %s to %s", item, targetFile)
	// 	}

	// 	pterm.Success.Printfln("generated files moved to: %s", constants.CacheDirectory)
	// 	pterm.Success.Println("(Cert) Generate()")
	// }()

	_, _, err := verifyBinaries()
	if err != nil {
		return err
	}
	pterm.Info.Println("choose which cert to generate")
	selection, err := pterm.DefaultInteractiveSelect.WithOptions(
		[]string{
			"Sidecar to Broker GRPC",
			"Sidecar To Broker Token",
		},
	).Show()
	if err != nil {
		return err
	}
	switch selection {
	case "Sidecar to Broker GRPC":
		// #############
		// # TOKEN #
		// #############.
		_, err = script.Exec("cfssl -log_dir=.artifacts gencert -loglevel=5 -initca config/cert/ca-csr.json").WriteFile(".cache/outcert.json") // Exec("cfssljson -bare ca").Stdout().
		if err != nil {
			pterm.Error.Printfln("issue running cfssl -bare ca: %v", err)
			return err
		}

		_, err = script.Exec("cfssljson -log_dir=.artifacts -f .cache/outcert.json -bare ca").Stdout()
		if err != nil {
			pterm.Error.Printfln("issue running cfssljson -bare ca: %v", err)
			return err
		}

		time.Sleep(duration)
		_, err = script.Exec("cfssl -log_dir=.artifacts gencert -loglevel=5 -ca=ca.pem -ca-key=ca-key.pem -config=config/cert/ca-config.json -profile=server config/cert/auth-csr.json").Exec("cfssljson -bare auth").Stdout()
		if err != nil {
			pterm.Error.Printfln("issue running cfssljson -bare ca: %v", err)
			return err
		}
		if err := moveFiles(constants.PrefixSidecarToBrokerGRPC); err != nil {
			pterm.Error.Printfln("terminating, due to failure in moving files: %v", err)
			return err
		}

	case "Sidecar To Broker Token":

		// #############
		// # CERTTOKEN #
		// #############.
		_, err = script.Exec("cfssl -log_dir=.artifacts gencert -loglevel=5 -initca config/certtoken/ca-csr.json").WriteFile(".cache/outcerttoken.json") // Exec("cfssljson -bare ca").Stdout().
		if err != nil {
			pterm.Error.Printfln("issue running cfssl -bare ca: %v", err)
			return err
		}
		_, err = script.Exec("cfssljson -log_dir=.artifacts -f .cache/outcerttoken.json -bare ca").Stdout()
		if err != nil {
			pterm.Error.Printfln("issue running cfssljson -bare ca: %v", err)
			return err
		}

		time.Sleep(duration)
		_, err = script.Exec("cfssl -log_dir=.artifacts gencert -loglevel=5 -ca=ca.pem -ca-key=ca-key.pem -config=config/certtoken/ca-config.json -profile=server config/certtoken/auth-csr.json").Exec("cfssljson -bare auth").Stdout()
		if err != nil {
			pterm.Error.Printfln("issue running cfssljson -bare ca: %v", err)
			return err
		}
		if err := moveFiles(constants.PrefixSidecarToBrokerToken); err != nil {
			pterm.Error.Printfln("terminating, due to failure in moving files: %v", err)
			return err
		}
	}

	return nil
}

// moveFiles moves the files to the cert cache directory, and requires a prefix so it's clear which config was used.
func moveFiles(prefix string) error {
	renameListCerts := []renameFile{
		{OriginalName: "auth-key.pem", NewName: fmt.Sprintf("%s-auth-key.pem", prefix)},
		{OriginalName: "auth.csr", NewName: fmt.Sprintf("%s-auth.csr", prefix)},
		{OriginalName: "auth.pem", NewName: fmt.Sprintf("%s-auth.pem", prefix)},
		{OriginalName: "ca-key.pem", NewName: fmt.Sprintf("%s-ca-key.pem", prefix)},
		{OriginalName: "ca.csr", NewName: fmt.Sprintf("%s-ca.csr", prefix)},
		{OriginalName: "ca.pem", NewName: fmt.Sprintf("%s-ca.pem", prefix)},
	}
	for _, item := range renameListCerts {
		targetFile := filepath.Join(constants.CacheCertDirectory, item.NewName)
		if err := os.Rename(item.OriginalName, targetFile); err != nil {
			pterm.Warning.Printfln("unable to move original: %s to new: %s", item.OriginalName, item.NewName)
		}
		pterm.Info.Printfln("mv %s to %s", item.OriginalName, targetFile)
		if err := sh.Rm(item.OriginalName); err != nil {
			pterm.Error.Printfln("unable to remove original, terminating early to avoid duplicate file confusion: %s", item.OriginalName)
			return err
		}
	}
	return nil
}

// verifyBinaries returns back the binaries for the cert tooling, and errors out if they aren't available.
// They should be installed on system, and this can be done easily with aqua included in the repo.
func verifyBinaries() (string, string, error) {
	var errCount int
	cffslbinary, err := exec.LookPath("cfssl")
	if err != nil && os.IsNotExist(err) {
		pterm.Error.Printfln("unable to find cfssl, you need to run `aqua install` and ensure aqua binaries are setup in path before running this")
		errCount++
	}
	pterm.Success.Printfln("cffslbinary: %s", cffslbinary)

	cffsljsonbinary, err := exec.LookPath("cfssljson")
	if err != nil && os.IsNotExist(err) {
		pterm.Error.Printfln("unable to find cfssljson, you need to run `aqua install` and ensure aqua binaries are setup in path before running this")
		errCount++
	}
	pterm.Success.Printfln("cffsljsonbinary: %s", cffsljsonbinary)
	if errCount > 0 {
		pterm.Error.Println("run aqua to install tooling for cfssl and cfssljson")
		return "", "", fmt.Errorf("errorcount: %d, missing required tooling", errCount)
	}
	return cffslbinary, cffsljsonbinary, nil
}

// If err != nil {
// 	pterm.Error.Printfln("gencert initca failure: %v", err)
// 	return err
// }
// return nil.

// Cmd3 := fmt.Sprintf("%s gencert -ca=ca.pem -ca-key=ca-key.pem -config=certs/config/ca-config.json -profile=server certs/config/auth-csr.json", cffslbinary)
// cmd4 := fmt.Sprintf("%s -bare auth", cffsljsonbinary)
// pterm.Debug.Printfln("cmd3: %s", cmd3)
// pterm.Debug.Printfln("cmd4: %s", cmd3)
// pterm.Debug.Println("generate the auth service certs")
// _, err = script.
// 	Exec(cmd3).
// 	Exec(cmd4).Stdout()
// if err != nil {
// 	pterm.Error.Printfln("failure: %v", err)
// 	return err
// }
// return nil
// }.

/*

Sheldon: I tried to use library directly, but the cfssljson doesn't give any command object so we'll have to use cli command control
	outputCSRJSON := filepath.Join(constants.ConfigDirectory, "ca-csr.json")
	gca := gencert.Command

	if err := gca.Main(
		[]string{
			outputCSRJSON,
		}, cli.Config{
			IsCA: true,
		},
	); err != nil {
		return err
	}



*/
