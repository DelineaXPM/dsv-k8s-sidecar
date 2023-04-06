package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DelineaXPM/dsv-k8s-sidecar/magefiles/constants"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// DSV is the namespace for mage tasks related to DSV, such as client credential creation.
type DSV mg.Namespace

var (
	dsvprofilename   = os.Getenv("DSV_PROFILE_NAME")
	rolename         = "github-dsv-k8s-sidecar-tests"
	policyname       = fmt.Sprintf("secrets:%s", secretpath)
	policysubjects   = fmt.Sprintf("roles:%s", rolename)
	policyresources  = fmt.Sprintf("secrets:%s:<.*>", secretpath)
	secretpath       = fmt.Sprintf("ci:tests:%s", "dsv-k8s-sidecar")
	secretpathclient = fmt.Sprintf("clients:%s", secretpath)
	desc             = fmt.Sprintf("a secret for testing operation of with dsv-k8s-sidecar")
	clientcredfile   = filepath.Join(constants.CacheDirectory, fmt.Sprintf("%s.json", rolename))
	clientcredname   = fmt.Sprintf("%s", rolename)
	secretkey        = "food" // just simple test placeholder for now
	testsecretkey    = fmt.Sprintf("secrets:%s:%s", secretpath, secretkey)
	testsecretvalue  = `{"taco":"burrito"}` //  placeholder for testing, not sensitive, and ok to leave for now

)

func checkDSVProfileName() {
	if dsvprofilename == "" {
		pterm.Error.Println(
			"DSV_PROFILE_NAME is not set and this is required to ensure the correct dsv tenant for testing is used",
		)
		panic("DSV_PROFILE_NAME is not set")
	}
}

// ➕ SetupAll creates the policy, role, and client credentials.
func (DSV) SetupAll() error {
	magetoolsutils.CheckPtermDebug()
	checkDSVProfileName()
	pterm.Warning.Println("WIP: initial creation to help with future testing setup, may need refinement")
	logger := pterm.DefaultLogger.WithLevel(pterm.LogLevelInfo).WithCaller(true)

	// dsv role create
	logger.Info("creating role", logger.Args("rolename", rolename))

	if err := sh.RunV("dsv", "role", "create", "--name", rolename, "--profile", dsvprofilename); err != nil {
		logger.Error("unable to create role", logger.Args("rolename", rolename))
		return err
	}
	logger.Info("created role", logger.Args("rolename", rolename))

	// dsv policy create
	if err := sh.RunV("dsv", "policy", "create",
		"--path", policyname,
		"--actions", "read,list",
		"--effect", "allow",
		"--subjects", policysubjects,
		"--desc", fmt.Sprintf("scoped access for %s by %s", secretpath, rolename),
		"--resources", policyresources,
		"--profile", dsvprofilename,
	); err != nil {
		logger.Error("unable to create policy", logger.Args("policyname", rolename))
		return err
	}
	logger.Info("created policy", logger.Args("policyname", rolename))

	logger.Info("creating client credentials", logger.Args("clientcredname", clientcredname))
	err := sh.RunV(
		"dsv",
		"client",
		"create",
		"--role", rolename,
		"--plain",
		"--profile", dsvprofilename,
		"--out", fmt.Sprintf("file:%s", clientcredfile),
	)
	if err != nil {
		return err
	}
	logger.Info("created client credentials", logger.Args("clientcredname", clientcredname))

	type ClientCredentials struct {
		ClientID string `json:"clientId"`
		Secret   string `json:"clientSecret"`
	}

	b, err := os.ReadFile(clientcredfile)
	if err != nil {
		logger.Error(
			"unable to read client credentials file",
			logger.Args("clientcredfile", clientcredfile, "error", err),
		)
		return err
	}
	var clientcred ClientCredentials
	err = json.Unmarshal(b, &clientcred)
	if err != nil {
		logger.Error(
			"unable to unmarshal client credentials file",
			logger.Args("clientcredfile", clientcredfile, "error", err),
		)
		return err
	}

	logger.Info("Put in .cache/charts/dsv-k8s-controller/values.yaml", logger.Args(
		"clientID", clientcred.ClientID,
		"clientSecret", clientcred.Secret,
	))

	return nil
}

// 🔐 CreateSecret creates a secret for usage with this specific client, policy, and role setup.
// This probably needs refactoring to allow input via pterm or via file.
// At time of creation (2023-04) it's a draft task to help with better test setup for developers wanting to test and have isolated
// permissions for just this specific secret path, role, client. It's all hard coded but can improve in the future.
func (DSV) CreateSecret() error {
	magetoolsutils.CheckPtermDebug()
	checkDSVProfileName()
	pterm.Warning.Println("WIP: initial creation to help with future testing setup, may need refinement")
	logger := pterm.DefaultLogger.WithLevel(pterm.LogLevelInfo).WithCaller(true)
	logger.Info("creating secret for DSV client")
	secretkey := "food"
	if err := sh.RunV(
		"dsv",
		"secret",
		"create",
		"--path", testsecretkey,
		"--data", testsecretvalue,
		"--desc", desc,
		"--profile", dsvprofilename,
	); err != nil {
		logger.Error("unable to create secret", logger.Args("secretkey", secretkey, "error", err))
		return err
	}
	logger.Info("created secret for DSV client", logger.Args("secretkey", secretkey))
	return nil
}
